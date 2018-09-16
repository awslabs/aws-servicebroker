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

func prescribeOverrides(b AwsBroker, services []osb.Service) []osb.Service {
	if !b.prescribeOverrides {
		return services
	} else {
		// TODO: Alot of duplication of code with ServiceDefinitionToOsb, should cleanup
		for s, service := range services {
			for p, plan := range service.Plans {
				availableParams := getAvailableParams(&plan)
				overrides := getOverrides(b.brokerid, availableParams, "all", "all", "all")
				overrideKeys := make([]string, 0)
				for o := range overrides {
					overrideKeys = append(overrideKeys, o)
				}
				glog.Infoln(overrideKeys)
				schemas := map[string]map[string]interface{}{
					"create": plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{}),
				}
				if plan.Schemas.ServiceInstance.Update != nil {
					schemas["update"] = plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})
				}
				for schemaName, schema := range schemas {
					props := make(map[string]interface{})
					required := make([]string, 0)
					for k, v := range schema {
						switch k {
						case "properties":
							for pk, pv := range v.(map[string]interface{}) {
								if !stringInSlice(pk, overrideKeys) {
									props[pk] = pv
								}
							}
						case "required":
							glog.Infoln(v)
							for _, r := range v.([]string) {
								if !stringInSlice(r, overrideKeys) {
									required = append(required, r)
								}
							}
						}
					}
					if schemaName == "create" {
						plan.Schemas = &osb.Schemas{
							ServiceInstance: &osb.ServiceInstanceSchema{
								Create: &osb.InputParametersSchema{
									Parameters: map[string]interface{}{
										"type":       "object",
										"properties": props,
										"$schema":    "http://json-schema.org/draft-06/schema#",
										"required":   required,
									},
								},
							},
						}
					} else if schemaName == "update" {
						if len(props) > 0 {
							plan.Schemas.ServiceInstance.Update = &osb.InputParametersSchema{
								Parameters: map[string]interface{}{
									"type":       "object",
									"properties": props,
									"$schema":    "http://json-schema.org/draft-06/schema#",
									"required":   required,
								},
							}
						}
					}
				}
				services[s].Plans[p] = plan
			}
		}
		return services
	}
}

func GetOverridesFromEnv() map[string]string {
	var Overrides = make(map[string]string)

	for _, item := range os.Environ() {
		envvar := strings.Split(item, "=")
		if strings.HasPrefix(envvar[0], "PARAM_OVERRIDE_") {
			key := strings.TrimPrefix(envvar[0], "PARAM_OVERRIDE_")
			if envvar[1] != "" {
				Overrides[key] = envvar[1]
				glog.V(10).Infof("%q=%q\n", key, envvar[1])
			}
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

func getOverrides(brokerid string, params []string, space string, service string, cluster string) (overrides map[string]string) {
	overridesEnv := GetOverridesFromEnv()

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
					if _, ok := overridesEnv[paramname]; ok {
						overrides[p] = overridesEnv[paramname]
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

func getCluster(context map[string]interface{}) string {
	switch context["platform"] {
	case osb.PlatformCloudFoundry:
		return strings.Replace(context["organization_guid"].(string), "-", "", -1)
	case osb.PlatformKubernetes:
		return context["clusterid"].(string)
	default:
		return "unknown"
	}
}

func getNamespace(context map[string]interface{}) string {
	switch context["platform"] {
	case osb.PlatformCloudFoundry:
		return strings.Replace(context["space_guid"].(string), "-", "", -1)
	case osb.PlatformKubernetes:
		return context["namespace"].(string)
	default:
		return "unknown"
	}
}

func getPlan(service *osb.Service, planID string) *osb.Plan {
	for _, p := range service.Plans {
		if p.ID == planID {
			return &p
		}
	}
	return nil
}

func getPlanDefaults(plan *osb.Plan) map[string]string {
	defaults := make(map[string]string)
	for k, v := range plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})["properties"].(map[string]interface{}) {
		if d, ok := v.(map[string]interface{})["default"]; ok {
			defaults[k] = paramValue(d)
		}
	}
	return defaults
}

func getAvailableParams(plan *osb.Plan) (params []string) {
	properties := plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})["properties"]
	if properties != nil {
		for k, _ := range properties.(map[string]interface{}) {
			params = append(params, k)
		}
	}
	return
}

func getUpdatableParams(plan *osb.Plan) (params []string) {
	properties := plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})["properties"]
	if properties != nil {
		for k, _ := range properties.(map[string]interface{}) {
			params = append(params, k)
		}
	}
	return
}

func getRequiredParams(plan *osb.Plan) (params []string) {
	required := plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})["required"]
	if required != nil {
		for _, p := range required.([]interface{}) {
			params = append(params, p.(string))
		}
	}
	return
}

func paramValue(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
