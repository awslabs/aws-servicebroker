package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"syscall"

	"github.com/golang/glog"
	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/awslabs/aws-servicebroker/pkg/broker"
	"github.com/jaymccon/osb-broker-lib/pkg/server"
	"github.com/pmorie/osb-broker-lib/pkg/metrics"
	"github.com/pmorie/osb-broker-lib/pkg/rest"
)

var options struct {
	broker.Options

	Port              int
	Insecure          bool
	TLSCert           string
	TLSKey            string
	TLSCertFile       string
	TLSKeyFile        string
	EnableBasicAuth   bool
	BasicAuthUser     string
	BasicAuthPassword string
}

func init() {
	flag.IntVar(&options.Port, "port", 8443, "use '--port' option to specify the port for broker to listen on")
	flag.BoolVar(&options.Insecure, "insecure", false, "use --insecure to use HTTP vs HTTPS.")
	flag.StringVar(&options.TLSCertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	flag.StringVar(&options.TLSKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	flag.StringVar(&options.TLSCert, "tlsCert", "", "base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used.")
	flag.StringVar(&options.TLSKey, "tlsKey", "", "base-64 encoded PEM block to use as the private key matching the TLS certificate.")
	flag.BoolVar(&options.EnableBasicAuth, "enableBasicAuth", false, "Enable HTTP Basic Authentication")
	flag.StringVar(&options.BasicAuthUser, "basicAuthUser", "", "HTTP Basic Authentication user")
	flag.StringVar(&options.BasicAuthPassword, "basicAuthPass", "", "HTTP Basic Authentication password")
	broker.AddFlags(&options.Options)
	flag.Parse()
}

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), "0.1.0")
		return nil
	}
	if (options.TLSCert != "" || options.TLSKey != "") &&
		(options.TLSCert == "" || options.TLSKey == "") {
		fmt.Println("To use TLS with specified cert or key data, both --tlsCert and --tlsKey must be used")
		return nil
	}

	matched, _ := regexp.MatchString("^[[:alnum:]]*$", options.BrokerID)
	if !matched {
		glog.Fatalln("brokerId can only contain letters and numbers")
	}

	addr := ":" + strconv.Itoa(options.Port)

	clients := broker.AwsClients{
		NewCfn:    broker.AwsCfnClientGetter,
		NewS3:     broker.AwsS3ClientGetter,
		NewSsm:    broker.AwsSsmClientGetter,
		NewSts:    broker.AwsStsClientGetter,
		NewDdb:    broker.AwsDdbClientGetter,
		NewIam:    broker.AwsIamClientGetter,
		NewLambda: broker.AwsLambdaClientGetter,
	}

	awsBroker, err := broker.NewAWSBroker(options.Options, broker.AwsSessionGetter, clients, broker.GetCallerId, broker.UpdateCatalog, broker.PollUpdate)
	if err != nil {
		glog.Fatalln(err)
	}

	// Prom. metrics
	reg := prom.NewRegistry()
	osbMetrics := metrics.New()
	reg.MustRegister(osbMetrics)

	api, err := rest.NewAPISurface(awsBroker, osbMetrics)
	if err != nil {
		return err
	}
	if options.BasicAuthUser == "" {
		options.BasicAuthUser = os.Getenv("SECURITY_USER_NAME")
	}
	if options.BasicAuthPassword == "" {
		options.BasicAuthPassword = os.Getenv("SECURITY_USER_PASSWORD")
	}
	auth := server.BasicAuth{User: options.BasicAuthUser, Pass: options.BasicAuthPassword}
	s := server.New(api, reg, options.EnableBasicAuth, auth.Secret)

	glog.Infof("Starting broker!")

	if options.Insecure {
		err = s.Run(ctx, addr)
	} else {
		if options.TLSCert != "" && options.TLSKey != "" {
			glog.V(4).Infof("Starting secure broker with TLS cert and key data")
			err = s.RunTLS(ctx, addr, options.TLSCert, options.TLSKey)
		} else {
			if options.TLSCertFile == "" || options.TLSKeyFile == "" {
				glog.Error("unable to run securely without TLS Certificate and Key. Please review options and if running with TLS, specify --tls-cert-file and --tls-private-key-file or --tlsCert and --tlsKey.")
				return nil
			}
			glog.V(4).Infof("Starting secure broker with file based TLS cert and key")
			err = s.RunTLSWithTLSFiles(ctx, addr, options.TLSCertFile, options.TLSKeyFile)
		}
	}
	return err
}

func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			glog.Infof("Received SIGTERM, exiting gracefully...")
			f()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}
