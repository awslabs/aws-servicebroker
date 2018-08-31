package broker

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

var nonCfnParams = []string{
	"aws_access_key",
	"aws_secret_key",
	"target_account_id",
	"target_role_name",
}

func GetOverridesFromEnv() map[string]string {
	var Overrides = make(map[string]string)

	for _, item := range os.Environ() {
		envvar := strings.Split(item, "=")
		if strings.HasPrefix(envvar[0], "PARAM_OVERRIDE_") {
			key := strings.TrimPrefix(envvar[0], "PARAM_OVERRIDE_")
			Overrides[key] = envvar[1]
			glog.V(10).Infof("%q=%q\n", key, envvar[1])
		}
	}
	return Overrides
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

func getParams(in interface{}) (keys []string) {
	p := in.(map[string]interface{})
	params, ok := p["properties"]
	if !ok {
		panic("unable to find properties keys")
	}
	innerparams := params.(map[string]interface{})
	for k := range innerparams {
		keys = append(keys, k)
	}
	return keys
}

func getOverrides(brokerid string, params []string, space string, service string, cluster string) (overrides map[string]string) {
	overrides_env := GetOverridesFromEnv()

	var services []string
	var namespaces []string
	var clusters []string
	if service != "all" {
		services = append(services, "all")
	}
	if space != "all" {
		namespaces = append(namespaces, "all")
	}
	if cluster != "all" {
		clusters = append(clusters, "all")
	}
	overrides = make(map[string]string)
	services = append(services, service)
	namespaces = append(namespaces, space)
	clusters = append(clusters, cluster)
	for _, c := range clusters {
		for _, n := range namespaces {
			for _, s := range services {
				for _, p := range params {
					paramname := brokerid + "_" + c + "_" + n + "_" + s + "_" + p
					// removing getting overrides from dynamo for the time being
					/*
						v, err := b.db.DataStorePort.GetParam(paramname)
						if err != nil {
							glog.Infof("Unable to fetch parameter override for %#+v", paramname)
							glog.Infoln(err.Error())
						}
						if v != "" {
							overrides[p] = v
						}
					*/
					if _, ok := overrides_env[paramname]; ok {
						overrides[p] = overrides_env[paramname]
					}
				}
			}
		}
	}
	glog.Infof("Overrides: '%+v'.", overrides)
	return overrides
}

// Build aws credentials using global or override keys, or the credential chain
func AwsCredentialsGetter(keyid string, secretkey string, profile string, params map[string]string) credentials.Credentials {
	if _, ok := params["aws_access_key"]; ok {
		keyid = params["aws_access_key"]
		glog.V(10).Infof("Using override credentials with keyid %q\n", keyid)
	}
	if _, ok := params["aws_secret_key"]; ok {
		secretkey = params["aws_secret_key"]
	}
	if keyid != "" && secretkey != "" {
		glog.Infof("Found 'aws_access_key' and 'aws_secret_key' in params, using credentials keyid '%q'.", keyid)
		return *credentials.NewStaticCredentials(keyid, secretkey, "")
	} else if profile != "" {
		glog.Infof("Profile specified, using profile '%q'.", profile)
		return *credentials.NewChainCredentials([]credentials.Provider{&credentials.SharedCredentialsProvider{Profile: profile}})
	}
	glog.Infof("Did not find 'aws_access_key' and 'aws_secret_key' in params, using default chain.")
	return *credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.Must(session.NewSession()))},
		})
}

// add trailing / if needed
func AddTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") == false {
		s = s + "/"
	}
	return s
}

func generateRoleArn(params map[string]string, currentAccountId string) string {
	targetRoleName := params["target_role_name"]

	if _, ok := params["target_account_id"]; ok {
		targetAccountId := params["target_account_id"]

		glog.Infof("Params 'target_account_id' present in params, assuming role in target account '%s'.", targetAccountId)
		return fmtArn(targetAccountId, targetRoleName)
	}

	glog.Infof("Params 'target_account_id' not present in params, assuming role in current account '%s'.", currentAccountId)
	return fmtArn(currentAccountId, targetRoleName)
}

func fmtArn(accountId, roleName string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, roleName)
}

func toCFNParams(params map[string]string) []*cloudformation.Parameter {
	var cfnParams []*cloudformation.Parameter
	for k, v := range params {
		if stringInSlice(k, nonCfnParams) {
			continue
		}
		cfnParams = append(cfnParams, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}
	return cfnParams
}

func newAsyncError() osb.HTTPStatusCodeError {
	return newHTTPStatusCodeError(http.StatusUnprocessableEntity, osb.AsyncErrorMessage, osb.AsyncErrorDescription)
}

func newHTTPStatusCodeError(statusCode int, msg, desc string) osb.HTTPStatusCodeError {
	err := osb.HTTPStatusCodeError{
		StatusCode: statusCode,
	}
	if msg != "" {
		err.ErrorMessage = &msg
	}
	if desc != "" {
		err.Description = &desc
	}
	glog.Error(err)
	return err
}
