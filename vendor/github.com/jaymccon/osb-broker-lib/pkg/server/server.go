package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"time"

	auth "github.com/abbot/go-http-auth"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/bcrypt"

	"github.com/pmorie/osb-broker-lib/pkg/rest"
)

type BasicAuth struct {
	User string
	Pass string
}

func (b *BasicAuth) Secret(user, realm string) string {
	if user == b.User {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(b.Pass), bcrypt.DefaultCost)
		if err == nil {
			return string(hashedPassword)
		}
	}
	return ""
}

// Server is the server for the OSB REST API and the metrics API. A Server glues
// the HTTP operations to their implementations.
type Server struct {
	// Router is a mux.Router that registers the handlers for the HTTP
	// operations:
	//
	// - OSB API
	// - metrics API
	Router *mux.Router
}

// New creates a new Router and registers all the necessary endpoints and handlers.
func New(api *rest.APISurface, reg prom.Gatherer, enableBasicAuth bool, secret func(user, realm string) string) *Server {
	router := mux.NewRouter()

	if api.EnableCORS {
		router.Methods("OPTIONS").HandlerFunc(api.OptionsHandler)
	}

	registerAPIHandlers(router, api, enableBasicAuth, secret)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	return &Server{
		Router: router,
	}
}

// NewHTTPHandler creates a new Router and registers API handlers
func NewHTTPHandler(api *rest.APISurface, enableBasicAuth bool, secret func(user, realm string) string) http.Handler {
	router := mux.NewRouter()
	registerAPIHandlers(router, api, enableBasicAuth, secret)
	return router
}

func getHandleFunc(handler func(w http.ResponseWriter, r *http.Request), enableBasicAuth bool, secret func(user, realm string) string) func(w http.ResponseWriter, r *http.Request) {
	if enableBasicAuth {
		authenticator := auth.NewBasicAuthenticator("aws-service-broker", secret)
		return auth.JustCheck(authenticator, handler)
	} else {
		return handler
	}
}

// registerAPIHandlers registers the APISurface endpoints and handlers.
func registerAPIHandlers(router *mux.Router, api *rest.APISurface, enableBasicAuth bool, secret func(user, realm string) string) {
	router.HandleFunc("/v2/catalog", getHandleFunc(api.GetCatalogHandler, enableBasicAuth, secret)).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/last_operation", getHandleFunc(api.LastOperationHandler, enableBasicAuth, secret)).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", getHandleFunc(api.ProvisionHandler, enableBasicAuth, secret)).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}", getHandleFunc(api.DeprovisionHandler, enableBasicAuth, secret)).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}", getHandleFunc(api.UpdateHandler, enableBasicAuth, secret)).Methods("PATCH")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", getHandleFunc(api.BindHandler, enableBasicAuth, secret)).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", getHandleFunc(api.GetBindingHandler, enableBasicAuth, secret)).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation", getHandleFunc(api.BindingLastOperationHandler, enableBasicAuth, secret)).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", getHandleFunc(api.UnbindHandler, enableBasicAuth, secret)).Methods("DELETE")
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}

// Run creates the HTTP handler and begins to listen on the specified address.
func (s *Server) Run(ctx context.Context, addr string) error {
	listenAndServe := func(srv *http.Server) error {
		return srv.ListenAndServe()
	}
	return s.run(ctx, addr, listenAndServe)
}

// RunTLS creates the HTTPS handler based on the certifications that were passed
// and begins to listen on the specified address.
func (s *Server) RunTLS(ctx context.Context, addr string, cert string, key string) error {
	var decodedCert, decodedKey []byte
	var tlsCert tls.Certificate
	var err error
	decodedCert, err = base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return err
	}
	decodedKey, err = base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	tlsCert, err = tls.X509KeyPair(decodedCert, decodedKey)
	if err != nil {
		return err
	}
	listenAndServe := func(srv *http.Server) error {
		srv.TLSConfig = new(tls.Config)
		srv.TLSConfig.Certificates = []tls.Certificate{tlsCert}
		return srv.ListenAndServeTLS("", "")
	}
	return s.run(ctx, addr, listenAndServe)
}

// RunTLSWithTLSFiles creates the HTTPS handler based on the certification
// files that were passed and begins to listen on the specified address.
func (s *Server) RunTLSWithTLSFiles(ctx context.Context, addr string, certFilePath string, keyFilePath string) error {
	listenAndServe := func(srv *http.Server) error {
		return srv.ListenAndServeTLS(certFilePath, keyFilePath)
	}
	return s.run(ctx, addr, listenAndServe)
}

func (s *Server) run(ctx context.Context, addr string, listenAndServe func(srv *http.Server) error) error {
	glog.Infof("Starting server on %s\n", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if srv.Shutdown(c) != nil {
			srv.Close()
		}
	}()
	return listenAndServe(srv)
}
