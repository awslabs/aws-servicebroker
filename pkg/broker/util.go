package broker

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func getGlobalOverrides(brokerID string) map[string]string {
	prefix := fmt.Sprintf("%s_all_all_all_", brokerID)
	overrides := make(map[string]string)
	for k, v := range getOverridesFromEnv() {
		if strings.HasPrefix(k, prefix) {
			overrides[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return overrides
}

func prescribeOverrides(b AwsBroker, services []osb.Service) []osb.Service {
	if !b.prescribeOverrides || len(b.globalOverrides) == 0 {
		return services
	}

	var overridenKeys []string
	for k := range b.globalOverrides {
		overridenKeys = append(overridenKeys, k)
	}

	for _, service := range services {
		for _, plan := range service.Plans {
			params := plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})
			for _, k := range overridenKeys {
				delete(params["properties"].(map[string]interface{}), k)
				params["required"] = deleteFromSlice(params["required"].([]string), k)
			}
			if len(params["required"].([]string)) == 0 {
				// Cloud Foundry does not allow "required" to be an empty slice
				delete(params, "required")
			}

			if plan.Schemas.ServiceInstance.Update != nil {
				params := plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})
				for _, k := range overridenKeys {
					delete(params["properties"].(map[string]interface{}), k)
					params["required"] = deleteFromSlice(params["required"].([]string), k)
				}
				if len(params["properties"].(map[string]interface{})) == 0 {
					// If there are no updatable properties left, remove the update schema
					plan.Schemas.ServiceInstance.Update = nil
				} else if len(params["required"].([]string)) == 0 {
					// Cloud Foundry does not allow "required" to be an empty slice
					delete(params, "required")
				}
			}
		}
	}

	return services
}

func getOverridesFromEnv() map[string]string {
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

func deleteFromSlice(list []string, s string) []string {
	var out []string
	for _, v := range list {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
func toScreamingSnakeCase(s string) string {
	in := []rune(s)

	var out []rune
	for i, r := range in {
		if i > 0 && i < len(in)-1 && // If this is not the first or last rune
			unicode.IsUpper(r) && (unicode.IsLower(in[i-1]) || unicode.IsLower(in[i+1])) { // And it's an upper preceded or followed by a lower
			out = append(out, '_')
		}
		out = append(out, unicode.ToUpper(r))
	}

	return string(out)
}

func getOverrides(brokerid string, params []string, space string, service string, cluster string) (overrides map[string]string) {
	overridesEnv := getOverridesFromEnv()

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
func awsCredentialsGetter(keyid string, secretkey string, profile string, params map[string]string, client *ec2metadata.EC2Metadata) credentials.Credentials {
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
			&ec2rolecreds.EC2RoleProvider{Client: client},
		})
}

// add trailing / if needed
func addTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") == false {
		s = s + "/"
	}
	return s
}

func generateRoleArn(params map[string]string, currentAccountID string) string {
	targetRoleName := params["target_role_name"]

	if _, ok := params["target_account_id"]; ok {
		targetAccountID := params["target_account_id"]

		glog.Infof("Params 'target_account_id' present in params, assuming role in target account '%s'.", targetAccountID)
		return fmtArn(targetAccountID, targetRoleName)
	}

	glog.Infof("Params 'target_account_id' not present in params, assuming role in current account '%s'.", currentAccountID)
	return fmtArn(currentAccountID, targetRoleName)
}

// getStackName returns the stack name for a service instance. A stack name can
// contain only alphanumeric characters (case sensitive) and hyphens. It must
// start with an alphabetic character and cannot be longer than 128 characters.
func getStackName(serviceName, instanceID string) string {
	s := fmt.Sprintf("aws-service-broker-%s-%s", serviceName, instanceID)
	s = regexp.MustCompile("[^a-zA-Z0-9-]").ReplaceAllString(s, "-")
	if len(s) > 128 {
		s = s[0:128]
	}
	return s
}

func fmtArn(accountID, roleName string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)
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
		for k := range properties.(map[string]interface{}) {
			params = append(params, k)
		}
	}
	return
}

func getUpdatableParams(plan *osb.Plan) (params []string) {
	properties := plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})["properties"]
	if properties != nil {
		for k := range properties.(map[string]interface{}) {
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

func getCredentials(service *osb.Service, outputs []*cloudformation.Output, ssmSvc ssmiface.SSMAPI) (map[string]interface{}, error) {
	credentials := make(map[string]interface{})
	var ssmValues []string

	for _, o := range outputs {
		if strings.HasPrefix(aws.StringValue(o.OutputKey), cfnOutputPolicyArnPrefix) {
			continue
		}

		// The output keys "UserKeyId" and "UserSecretKey" require special handling for backward compatibility :/
		if aws.StringValue(o.OutputKey) == cfnOutputUserKeyID || aws.StringValue(o.OutputKey) == cfnOutputUserSecretKey {
			k := fmt.Sprintf("%s_%s", strings.ToUpper(service.Name), toScreamingSnakeCase(aws.StringValue(o.OutputKey)))
			credentials[k] = aws.StringValue(o.OutputValue)
			ssmValues = append(ssmValues, aws.StringValue(o.OutputValue))
		} else {
			credentials[toScreamingSnakeCase(aws.StringValue(o.OutputKey))] = aws.StringValue(o.OutputValue)
			// If the output value starts with "ssm:", we'll get the actual value from SSM
			if strings.HasPrefix(aws.StringValue(o.OutputValue), cfnOutputSSMValuePrefix) {
				ssmValues = append(ssmValues, strings.TrimPrefix(aws.StringValue(o.OutputValue), cfnOutputSSMValuePrefix))
			}
		}
	}

	if len(ssmValues) > 0 {
		resp, err := ssmSvc.GetParameters(&ssm.GetParametersInput{
			Names:          aws.StringSlice(ssmValues),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, err
		} else if len(resp.InvalidParameters) > 0 {
			return nil, fmt.Errorf("invalid parameters: %v", aws.StringValueSlice(resp.InvalidParameters))
		}

		for _, p := range resp.Parameters {
			for k, v := range credentials {
				if strings.TrimPrefix(v.(string), cfnOutputSSMValuePrefix) == aws.StringValue(p.Name) {
					credentials[k] = aws.StringValue(p.Value)
				}
			}
		}
	}

	return credentials, nil
}

func getPolicyArn(outputs []*cloudformation.Output, scope string) (string, error) {
	outputKey := fmt.Sprintf("%s%s", cfnOutputPolicyArnPrefix, scope)
	for _, o := range outputs {
		if strings.EqualFold(aws.StringValue(o.OutputKey), outputKey) {
			return aws.StringValue(o.OutputValue), nil
		}
	}
	return "", fmt.Errorf("output not found: %s", outputKey)
}
