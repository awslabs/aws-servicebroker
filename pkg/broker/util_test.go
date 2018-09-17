package broker

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func clearOverrides() {
	// TODO: this breaks parallel testing, should mock out os.*Env functions
	for _, item := range os.Environ() {
		envvar := strings.Split(item, "=")
		if strings.HasPrefix(envvar[0], "PARAM_OVERRIDE_") {
			os.Unsetenv(envvar[0])
		}
	}
}

func TestPrescribeOverrides(t *testing.T) {
	assertor := assert.New(t)

	services := []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param":      map[string]interface{}{"type": "string"},
						"override_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param", "override_param"},
					},
				}},
			}},
		}},
	}

	g := map[string]string{"override_param": "overridden"}

	msg := "params should not be modified when prescribeOverrides is false"
	psvcs := prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: false, globalOverrides: g}, services)
	expected := []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param":      map[string]interface{}{"type": "string"},
						"override_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param", "override_param"},
					},
				}},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)

	msg = "override_param should be removed when prescribeOverrides is true"
	psvcs = prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: true, globalOverrides: g}, services)
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param"},
					},
				}},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)

	services = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param", "override_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param", "override_param"},
						},
					},
				},
			}},
		}},
	}

	msg = "override_param should be removed from Update params too when prescribeOverrides is true"
	psvcs = prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: true, globalOverrides: g}, services)
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
				},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)
	clearOverrides()
}

func TestGetOverridesFromEnv(t *testing.T) {
	assertor := assert.New(t)

	clearOverrides()

	msg := "should return empty map if there are no overrides set"
	output := GetOverridesFromEnv()
	assertor.Equal(make(map[string]string), output, msg)

	msg = "should return map with all the found overrides, excluding any environment variables not prefixed with PARAM_OVERRIDE_"
	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_test_param1", "testval1")
	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_test_param2", "testval2")
	os.Setenv("NOTMATCHPARAM_OVERRIDE_awsservicebroker_all_all_all_test_param3", "testval3")
	output = GetOverridesFromEnv()
	assertor.Equal(map[string]string{
		"awsservicebroker_all_all_all_test_param1": "testval1",
		"awsservicebroker_all_all_all_test_param2": "testval2",
	},
		output,
		msg,
	)
	clearOverrides()
}

func TestStringInSlice(t *testing.T) {
	assertor := assert.New(t)

	assertor.Equal(true, stringInSlice("present", []string{"somestr", "present", "anotherstr"}), "should return true")

	assertor.Equal(false, stringInSlice("notpresent", []string{"somestr", "present", "anotherstr"}), "should return false")
}

func TestToScreamingSnakeCase(t *testing.T) {
	assertor := assert.New(t)

	assertor.Equal("SCREAMING_SNAKE", toScreamingSnakeCase("ScreamingSnake"), "should convert camel to snake")

	assertor.Equal("AWS_TEST", toScreamingSnakeCase("AWSTest"), "Shouldn't put an underscore between consecutive caps")

}

func TestGetOverrides(t *testing.T) {
	assertor := assert.New(t)

	clearOverrides()
	brokerid, space, service, cluster := "awsservicebroker", "all", "all", "all"
	params := []string{"test_param1", "test_param2"}

	output := getOverrides(brokerid, params, space, service, cluster)
	assertor.Equal(make(map[string]string), output, "should return an empty slice if there's no matching overrides")

	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_test_param1", "testval1")
	output = getOverrides(brokerid, params, space, service, cluster)
	assertor.Equal(map[string]string{"test_param1": "testval1"}, output, "should return only items with matching overrides")

	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_notrightservice_test_param2", "testval2")
	output = getOverrides(brokerid, params, space, service, cluster)
	assertor.Equal(map[string]string{"test_param1": "testval1"}, output, "should return only items with matching overrides")

	brokerid, space, service, cluster = "awsservicebroker", "should", "not", "match"
	output = getOverrides(brokerid, params, space, service, cluster)
	assertor.Equal(map[string]string{"test_param1": "testval1"}, output, "should return only items with matching overrides")

	clearOverrides()

}

func TestAwsCredentialsGetter(t *testing.T) {
	assertor := assert.New(t)

	keyid, secretkey, profile := "", "", ""
	params := make(map[string]string)
	client := ec2metadata.New(session.Must(session.NewSession()))
	actual := AwsCredentialsGetter(keyid, secretkey, profile, params, client)
	expected := *credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{Client: client},
		})
	assertor.Equal(expected, actual, "should return credential chain creds")

	keyid, secretkey, profile = "testid", "testkey", ""
	expected = *credentials.NewStaticCredentials(keyid, secretkey, "")
	actual = AwsCredentialsGetter(keyid, secretkey, profile, params, client)
	assertor.Equal(expected, actual, "should return static creds")

	keyid, secretkey, profile = "", "", "test"
	expected = *credentials.NewChainCredentials([]credentials.Provider{&credentials.SharedCredentialsProvider{Profile: profile}})
	actual = AwsCredentialsGetter(keyid, secretkey, profile, params, client)
	assertor.Equal(expected, actual, "should return shared creds")

	keyid, secretkey, profile = "", "", ""
	params = map[string]string{"aws_access_key": "testKeyId", "aws_secret_key": "testSecretKey"}
	expected = *credentials.NewStaticCredentials("testKeyId", "testSecretKey", "")
	actual = AwsCredentialsGetter(keyid, secretkey, profile, params, client)
	assertor.Equal(expected, actual, "should return static creds")
}

func TestToCFNParams(t *testing.T) {
	assertor := assert.New(t)

	params := map[string]string{"pkey": "pval"}
	actual := toCFNParams(params)
	expected := []*cloudformation.Parameter{
		{
			ParameterKey:   aws.String("pkey"),
			ParameterValue: aws.String("pval"),
		},
	}
	assertor.Equal(expected, actual, "should return input mashalled into []*cloudformation.Parameter ")
}

func TestNewHTTPStatusCodeError(t *testing.T) {
	assertor := assert.New(t)

	code, msg, desc := 499, "testmsg", "test desc"
	expected := osb.HTTPStatusCodeError{StatusCode: code, ErrorMessage: &msg, Description: &desc}
	actual := newHTTPStatusCodeError(code, msg, desc)
	assertor.Equal(expected, actual, "should return a HTTPStatusCodeError with code, msg and desc matching the input")
}

func TestGetCluster(t *testing.T) {
	assertor := assert.New(t)

	context := map[string]interface{}{
		"platform":          osb.PlatformCloudFoundry,
		"organization_guid": "test-test",
	}
	assertor.Equal("testtest", getCluster(context), "should strip dashes from cf guid")

	context = map[string]interface{}{
		"platform":  osb.PlatformKubernetes,
		"clusterid": "testtest",
	}
	assertor.Equal("testtest", getCluster(context), "should return clusterid from context")

	context = map[string]interface{}{
		"platform":          "other",
		"organization_guid": "testtest",
	}
	assertor.Equal("unknown", getCluster(context), "should return unknown")
}

type mockSsmGetParameters struct {
	ssmiface.SSMAPI
	Resp ssm.GetParametersOutput
}

func (mockSsmGetParameters) GetParameters(in *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	params := make([]*ssm.Parameter, 0)
	for _, n := range in.Names {
		params = append(params, &ssm.Parameter{
			Name:  aws.String(*n),
			Value: aws.String("val-" + *n),
		})
	}
	return &ssm.GetParametersOutput{Parameters: params}, nil
}

func TestGetCredentials(t *testing.T) {
	assertor := assert.New(t)

	service := osb.Service{
		Name: "testsvc",
	}
	outputs := []*cloudformation.Output{
		{
			OutputKey:   aws.String(cfnOutputPolicyArnPrefix + "Test"),
			OutputValue: aws.String("testpolicyval"),
		},
		{
			OutputKey:   aws.String(cfnOutputUserKeyID),
			OutputValue: aws.String("testkeyval"),
		},
		{
			OutputKey:   aws.String("Test"),
			OutputValue: aws.String("testasisval"),
		},
		{
			OutputKey:   aws.String("TestSsmVal"),
			OutputValue: aws.String(cfnOutputSSMValuePrefix + "testssmval"),
		},
	}
	ssmSvc := mockSsmGetParameters{}

	expected := map[string]interface{}{
		"TEST":                "testasisval",
		"TESTSVC_USER_KEY_ID": "val-testkeyval",
		"TEST_SSM_VAL":        "val-testssmval",
	}
	actual, err := getCredentials(&service, outputs, ssmSvc)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "not getting expected output")
}
