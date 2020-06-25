package broker

import (
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
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
							"param":          map[string]interface{}{"type": "integer"},
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param", "override_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"param":          map[string]interface{}{"type": "integer"},
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
							"param":     map[string]interface{}{"type": "integer"},
							"req_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"param":     map[string]interface{}{"type": "integer"},
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

	msg = "required should be removed if all required params are overridden"
	b := AwsBroker{
		brokerid:           "awsservicebroker",
		prescribeOverrides: true,
		globalOverrides:    map[string]string{"override_param": "overridden", "req_param": "overridden"},
	}
	psvcs = prescribeOverrides(b, services)
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"param": map[string]interface{}{"type": "integer"},
						},
							"$schema": "http://json-schema.org/draft-06/schema#",
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"param": map[string]interface{}{"type": "integer"},
						},
							"$schema": "http://json-schema.org/draft-06/schema#",
						},
					},
				},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)

	msg = "should succeed when there are no required params"
	b = AwsBroker{
		brokerid:           "awsservicebroker",
		prescribeOverrides: true,
		globalOverrides:    map[string]string{"override_param": "overridden"},
	}
	psvcs = prescribeOverrides(b, []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"override_param": map[string]interface{}{"type": "string"},
					},
						"$schema": "http://json-schema.org/draft-06/schema#",
					},
				}},
			}},
		}},
	})
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{},
						"$schema": "http://json-schema.org/draft-06/schema#",
					},
				}},
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
	output := getOverridesFromEnv()
	assertor.Equal(make(map[string]string), output, msg)

	msg = "should return map with all the found overrides, excluding any environment variables not prefixed with PARAM_OVERRIDE_"
	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_test_param1", "testval1")
	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_test_param2", "testval2")
	os.Setenv("NOTMATCHPARAM_OVERRIDE_awsservicebroker_all_all_all_test_param3", "testval3")
	output = getOverridesFromEnv()
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
	stsClient := sts.New(session.Must(session.NewSession()))
	actual := awsCredentialsGetter(keyid, secretkey, profile, params, client, stsClient)
	expected := *credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{Client: client},
			stscreds.NewWebIdentityRoleProvider(stsClient, os.Getenv("AWS_ROLE_ARN"), os.Getenv("AWS_ROLE_SESSION_NAME"), os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")),
		})
	assertor.Equal(expected, actual, "should return credential chain creds")

	keyid, secretkey, profile = "testid", "testkey", ""
	expected = *credentials.NewStaticCredentials(keyid, secretkey, "")
	actual = awsCredentialsGetter(keyid, secretkey, profile, params, client, stsClient)
	assertor.Equal(expected, actual, "should return static creds")

	keyid, secretkey, profile = "", "", "test"
	expected = *credentials.NewChainCredentials([]credentials.Provider{&credentials.SharedCredentialsProvider{Profile: profile}})
	actual = awsCredentialsGetter(keyid, secretkey, profile, params, client, stsClient)
	assertor.Equal(expected, actual, "should return shared creds")

	keyid, secretkey, profile = "", "", ""
	params = map[string]string{"aws_access_key": "testKeyId", "aws_secret_key": "testSecretKey"}
	expected = *credentials.NewStaticCredentials("testKeyId", "testSecretKey", "")
	actual = awsCredentialsGetter(keyid, secretkey, profile, params, client, stsClient)
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
	assertor.Equal(expected, actual, "should return input marshalled into []*cloudformation.Parameter ")
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

func TestBindViaLambda(t *testing.T) {
	assertor := assert.New(t)

	assertor.False(bindViaLambda(&osb.Service{}), "should be False")
	assertor.True(bindViaLambda(
		&osb.Service{
			Metadata: map[string]interface{}{
				"bindViaLambda": true,
			},
		}))
	assertor.False(bindViaLambda(
		&osb.Service{
			Metadata: map[string]interface{}{
				"bindViaLambda": false,
			},
		}))
}

func TestInvokeLambdaBindFunc(t *testing.T) {
	tests := []struct {
		name                string
		inputCredentials    map[string]interface{}
		newLambdaF          GetLambdaClient
		expectedErr         string
		expectedCredentials map[string]interface{}
	}{
		{
			name:             "No BindLambda",
			inputCredentials: map[string]interface{}{},
			expectedErr:      "the template metadata has BindViaLambda set to true, but no BindLambda is defined in template output",
		},
		{
			name: "Non-string BindLambda",
			inputCredentials: map[string]interface{}{
				"BindLambda": 1,
			},
			expectedErr: "non string value for BindLambda in the cloudformation template",
		},
		{
			name: "Empty-string BindLambda",
			inputCredentials: map[string]interface{}{
				"BindLambda": "",
			},
			expectedErr: "the template metadata has BindViaLambda set to true, but the BindLambda output from cloudformation is an empty string",
		},
		{
			name: "lambda session is nil",
			inputCredentials: map[string]interface{}{
				"BindLambda": "MyLambdaFunc",
			},
			newLambdaF: func(s *session.Session) lambdaiface.LambdaAPI {
				return nil
			},
			expectedErr: "attempt to establish Lambda session return a nil client",
		},
		{
			name: "error in lambda script",
			inputCredentials: map[string]interface{}{
				"BindLambda": "MyLambdaFunc",
			},
			newLambdaF: func(s *session.Session) lambdaiface.LambdaAPI {
				return &mockLambda{
					lambdas: map[string]mockLambdaFunc{
						"MyLambdaFunc": func(payload []byte) ([]byte, error) {
							return []byte(`{"errorType":"SomeError","errorMessage":"there was an error"}`), nil
						},
					},
				}
			},
			expectedErr: "error in lambda function building binding: SomeError there was an error",
		},
		{
			name: "succesful bind",
			inputCredentials: map[string]interface{}{
				"BindLambda": "MyLambdaFunc",
			},
			newLambdaF: func(s *session.Session) lambdaiface.LambdaAPI {
				return &mockLambda{
					lambdas: map[string]mockLambdaFunc{
						"MyLambdaFunc": func(payload []byte) ([]byte, error) {

							return []byte(`{"MyKey":"MyVal"}`), nil
						},
					},
				}
			},
			expectedCredentials: map[string]interface{}{"MyKey": "MyVal"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			credentials, err := invokeLambdaBindFunc(nil, tc.newLambdaF, tc.inputCredentials, "bind")
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
				assert.Nil(t, credentials)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, credentials, tc.expectedCredentials)
		})
	}
}
