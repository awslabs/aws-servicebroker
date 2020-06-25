package broker

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/awslabs/aws-servicebroker/pkg/serviceinstance"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/stretchr/testify/assert"
)

func TestGetCatalog(t *testing.T) {
	assertor := assert.New(t)

	opts := Options{
		TableName:          "testtable",
		S3Bucket:           "abucket",
		S3Region:           "us-east-1",
		S3Key:              "tempates/test",
		Region:             "us-east-1",
		BrokerID:           "awsservicebroker",
		PrescribeOverrides: false,
	}
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
	bl.listingcache.Set("__LISTINGS__", []ServiceNeedsUpdate{{Name: "test", Update: false}})

	expected := &broker.CatalogResponse{CatalogResponse: osb.CatalogResponse{}}
	actual, err := bl.GetCatalog(&broker.RequestContext{})
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should return empty catalog")

	svc := osb.Service{
		ID:          "test-id",
		Name:        "test",
		Description: "blah",
		Plans: []osb.Plan{
			{
				ID:      "planid",
				Name:    "planname",
				Schemas: &osb.Schemas{},
			},
		},
	}

	bl.catalogcache.Set("test", svc)
	expected = &broker.CatalogResponse{CatalogResponse: osb.CatalogResponse{Services: []osb.Service{svc}}}
	actual, err = bl.GetCatalog(&broker.RequestContext{})
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should return a single service matching the mock")
}

type mockDataStoreProvision struct{}

func (db mockDataStoreProvision) PutServiceDefinition(sd osb.Service) error { return nil }
func (db mockDataStoreProvision) GetParam(paramname string) (value string, err error) {
	return "some-value", nil
}
func (db mockDataStoreProvision) PutParam(paramname string, paramvalue string) error { return nil }
func (db mockDataStoreProvision) PutServiceInstance(si serviceinstance.ServiceInstance) error {
	for _, v := range si.Params {
		if v == "err" {
			return errors.New("test failure")
		}
	}
	return nil
}
func (db mockDataStoreProvision) GetServiceDefinition(serviceuuid string) (*osb.Service, error) {

	if serviceuuid == "test-lambda-service-id" {

		return &osb.Service{
			ID:       "test-lambda-service-id",
			Name:     "test-service-name",
			Metadata: map[string]interface{}{"bindViaLambda": true},
		}, nil

	}
	if serviceuuid == "test-service-id" {
		return &osb.Service{
			ID:   "test-service-id",
			Name: "test-service-name",
			Plans: []osb.Plan{
				{ID: "test-plan-id", Name: "test-plan-name", Schemas: &osb.Schemas{ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
							"region":         map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []interface{}{"req_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"req_param": map[string]interface{}{"type": "string"},
							},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
				}}},
			},
		}, nil
	} else if serviceuuid == "err" {
		return nil, errors.New("test failure")
	} else if serviceuuid == "noplan" {
		return &osb.Service{}, nil
	}
	return nil, nil
}
func (db mockDataStoreProvision) GetServiceInstance(sid string) (*serviceinstance.ServiceInstance, error) {
	switch sid {
	case "err":
		return nil, errors.New("test failure")
	case "err-stack":
		return &serviceinstance.ServiceInstance{ID: "err-stack", StackID: "err", PlanID: "test-plan-id", Params: map[string]string{"req_param": "a-value"}}, nil
	case "exists":
		return &serviceinstance.ServiceInstance{ID: "exists", StackID: "an-id", PlanID: "test-plan-id", Params: map[string]string{"req_param": "a-value"}}, nil
	case "foo-plan":
		return &serviceinstance.ServiceInstance{ID: "foo-plan", StackID: "an-id", PlanID: "foo"}, nil
	default:
		return nil, nil
	}
}
func (db mockDataStoreProvision) DeleteServiceInstance(id string) error { return nil }
func (db mockDataStoreProvision) GetServiceBinding(id string) (*serviceinstance.ServiceBinding, error) {
	switch id {
	case "err":
		return nil, errors.New("test failure")
	case "err-instance":
		return &serviceinstance.ServiceBinding{
			ID:         "err-instance",
			InstanceID: "err",
			PolicyArn:  "exists",
			RoleName:   "exists",
		}, nil
	case "err-role-name":
		return &serviceinstance.ServiceBinding{
			ID:         "err-role-name",
			InstanceID: "exists",
			PolicyArn:  "exists",
			RoleName:   "err",
		}, nil
	case "exists":
		return &serviceinstance.ServiceBinding{
			ID:         "exists",
			InstanceID: "exists",
		}, nil
	case "exists-role-name":
		return &serviceinstance.ServiceBinding{
			ID:         "exists-role-name",
			InstanceID: "exists",
			PolicyArn:  "exists",
			RoleName:   "exists",
		}, nil
	case "foo-instance":
		return &serviceinstance.ServiceBinding{
			ID:         "foo-instance",
			InstanceID: "foo",
			PolicyArn:  "exists",
			RoleName:   "exists",
		}, nil
	case "foo-role-name":
		return &serviceinstance.ServiceBinding{
			ID:         "foo-role-name",
			InstanceID: "exists",
			PolicyArn:  "exists",
			RoleName:   "foo",
		}, nil
	default:
		return nil, nil
	}
}
func (db mockDataStoreProvision) PutServiceBinding(sb serviceinstance.ServiceBinding) error {
	return nil
}
func (db mockDataStoreProvision) DeleteServiceBinding(id string) error { return nil }

func TestProvision(t *testing.T) {
	assertor := assert.New(t)

	opts := Options{
		TableName:          "testtable",
		S3Bucket:           "abucket",
		S3Region:           "us-east-1",
		S3Key:              "tempates/test",
		Region:             "us-east-1",
		BrokerID:           "awsservicebroker",
		PrescribeOverrides: true,
	}
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}
	bl.globalOverrides = map[string]string{"override_param": "some_value"}
	provReq := &osb.ProvisionRequest{
		InstanceID:          "test-instance-id",
		ServiceID:           "test-service-id",
		PlanID:              "test-plan-id",
		OriginatingIdentity: &osb.OriginatingIdentity{},
		AcceptsIncomplete:   true,
		Parameters: map[string]interface{}{
			"region":       "us-east-1",
			"anotherParam": "pval",
		},
	}
	reqContext := &broker.RequestContext{}

	expectedErr := newHTTPStatusCodeError(http.StatusBadRequest, "", "The parameter anotherParam is not available.")
	_, err := bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with missing parameter error")

	provReq.Parameters = map[string]interface{}{
		"region": "us-east-1",
	}
	expectedErr = newHTTPStatusCodeError(http.StatusBadRequest, "", "The parameter req_param is required.")
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with required parameter error")

	provReq.Parameters = map[string]interface{}{
		"region":    "us-east-1",
		"req_param": "pval",
	}
	expected := &broker.ProvisionResponse{ProvisionResponse: osb.ProvisionResponse{Async: true}}
	actual, err := bl.Provision(provReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should return empty provision response")

	expectedErr = osb.HTTPStatusCodeError{
		StatusCode:   422,
		ErrorMessage: aws.String("AsyncRequired"),
		Description:  aws.String("This service plan requires client support for asynchronous service operations."),
	}
	_, err = bl.Provision(&osb.ProvisionRequest{AcceptsIncomplete: false}, &broker.RequestContext{})
	assertor.Equal(expectedErr, err, "err should be 422")

	expectedErr = newHTTPStatusCodeError(http.StatusBadRequest, "", "The service plan test-plan-id was not found.")
	provReq.ServiceID = "noplan"
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with missing plan error")

	expectedErr = newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service err: test failure")
	provReq.ServiceID = "err"
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with 500 test error")

	expectedErr = newHTTPStatusCodeError(http.StatusBadRequest, "", "The service nonexist was not found.")
	provReq.ServiceID = "nonexist"
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with 500 error")

	expectedErr = newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance err: test failure")
	provReq.ServiceID = "test-service-id"
	provReq.InstanceID = "err"
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with 500 error")

	expectedErr = newHTTPStatusCodeError(http.StatusConflict, "", "Service instance exists already exists but with different attributes.")
	provReq.ServiceID = "test-service-id"
	provReq.InstanceID = "exists"
	_, err = bl.Provision(provReq, reqContext)
	assertor.Equal(expectedErr, err, "should fail with 500 error")

}

func TestDeprovision(t *testing.T) {
	tests := []struct {
		name        string
		request     *osb.DeprovisionRequest
		expectedErr error
	}{
		{
			name: "async_required",
			request: &osb.DeprovisionRequest{
				AcceptsIncomplete: false,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
			},
			expectedErr: newAsyncError(),
		},
		{
			name: "error_getting_instance",
			request: &osb.DeprovisionRequest{
				AcceptsIncomplete: true,
				InstanceID:        "err",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance err: test failure"),
		},
		{
			name: "instance_not_found",
			request: &osb.DeprovisionRequest{
				AcceptsIncomplete: true,
				InstanceID:        "foo",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusGone, "", "The service instance foo was not found."),
		},
		{
			name: "error_deleting_stack",
			request: &osb.DeprovisionRequest{
				AcceptsIncomplete: true,
				InstanceID:        "err-stack",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to delete the CloudFormation stack err: test failure"),
		},
		{
			name: "success",
			request: &osb.DeprovisionRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := NewAWSBroker(Options{}, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
			b.db.DataStorePort = mockDataStoreProvision{}

			resp, err := b.Deprovision(tt.request, &broker.RequestContext{})
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Async)
			}
		})
	}
}

func TestLastOperation(t *testing.T) {
	tests := []struct {
		name              string
		request           *osb.LastOperationRequest
		stackStatus       string
		stackStatusReason string
		expectedState     osb.LastOperationState
		expectedDesc      *string
		expectedErr       error
	}{
		{
			name: "error_getting_instance",
			request: &osb.LastOperationRequest{
				InstanceID: "err",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance err: test failure"),
		},
		{
			name: "instance_not_found",
			request: &osb.LastOperationRequest{
				InstanceID: "foo",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusGone, "", "The service instance foo was not found."),
		},
		{
			name: "error_describing_stack",
			request: &osb.LastOperationRequest{
				InstanceID: "err-stack",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to describe the CloudFormation stack err: test failure"),
		},
		{
			name: "create_in_progress",
			request: &osb.LastOperationRequest{
				InstanceID: "exists",
			},
			stackStatus:   cloudformation.StackStatusCreateInProgress,
			expectedState: osb.StateInProgress,
		},
		{
			name: "create_complete",
			request: &osb.LastOperationRequest{
				InstanceID: "exists",
			},
			stackStatus:   cloudformation.StackStatusCreateComplete,
			expectedState: osb.StateSucceeded,
		},
		{
			name: "delete_complete",
			request: &osb.LastOperationRequest{
				InstanceID: "exists",
			},
			stackStatus:   cloudformation.StackStatusDeleteComplete,
			expectedState: osb.StateSucceeded,
		},
		{
			name: "update_rollback_complete",
			request: &osb.LastOperationRequest{
				InstanceID: "exists",
			},
			stackStatus:       cloudformation.StackStatusUpdateRollbackComplete,
			stackStatusReason: "foo",
			expectedState:     osb.StateFailed,
			expectedDesc:      aws.String("foo"),
		},
		{
			name: "rollback_in_progress",
			request: &osb.LastOperationRequest{
				InstanceID: "exists",
			},
			stackStatus:       cloudformation.StackStatusRollbackInProgress,
			stackStatusReason: "foo",
			expectedState:     osb.StateFailed,
			expectedDesc:      aws.String("foo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := AwsClients{
				NewCfn: func(sess *session.Session) CfnClient {
					return CfnClient{
						Client: mockCfn{
							DescribeStacksResponse: cloudformation.DescribeStacksOutput{
								Stacks: []*cloudformation.Stack{
									{
										StackStatus:       aws.String(tt.stackStatus),
										StackStatusReason: aws.String(tt.stackStatusReason),
									},
								},
							},
							DescribeStackEventsResponse: cloudformation.DescribeStackEventsOutput{
								StackEvents: []*cloudformation.StackEvent{
									{
										LogicalResourceId:    aws.String("testId"),
										ResourceStatus:       aws.String(tt.stackStatus),
										ResourceStatusReason: aws.String(tt.stackStatusReason),
									},
								},
							},
						},
					}
				},
				NewDdb: mockAwsDdbClientGetter,
				NewIam: mockAwsIamClientGetter,
				NewS3:  mockAwsS3ClientGetter,
				NewSts: mockAwsStsClientGetter,
			}

			b, _ := NewAWSBroker(Options{}, mockGetAwsSession, clients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
			b.db.DataStorePort = mockDataStoreProvision{}

			resp, err := b.LastOperation(tt.request, &broker.RequestContext{Request: &http.Request{Header: http.Header{}}})
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedState, resp.State)
				assert.Equal(t, tt.expectedDesc, resp.Description)
			}
		})
	}
}

func toDescribeStacksOutput(outputs map[string]string) cloudformation.DescribeStacksOutput {
	var cfnOutputs []*cloudformation.Output
	for k, v := range outputs {
		cfnOutputs = append(cfnOutputs, &cloudformation.Output{
			OutputKey:   aws.String(k),
			OutputValue: aws.String(v),
		})
	}
	return cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				Outputs: cfnOutputs,
			},
		},
	}
}

func TestBind(t *testing.T) {
	tests := []struct {
		name           string
		request        *osb.BindRequest
		cfnOutputs     map[string]string
		ssmParams      map[string]string
		expectedCreds  map[string]interface{}
		expectedExists bool
		expectedErr    error
		bindViaLambda  bool
		lambdas        map[string]mockLambdaFunc
	}{
		{
			name: "unsupported_parameter",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
				Parameters: map[string]interface{}{"foo": "bar"},
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The parameter foo is not supported."),
		},
		{
			name: "error_getting_binding",
			request: &osb.BindRequest{
				BindingID:  "err",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service binding err: test failure"),
		},
		{
			name: "existing_binding",
			request: &osb.BindRequest{
				BindingID:  "exists",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
			},
			expectedExists: true,
		},
		{
			name: "conflicting_binding",
			request: &osb.BindRequest{
				BindingID:  "exists",
				InstanceID: "foo",
				ServiceID:  "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusConflict, "", "Service binding exists already exists but with different attributes."),
		},
		{
			name: "error_getting_service",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "err",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service err: test failure"),
		},
		{
			name: "service_not_found",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "foo",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service foo was not found."),
		},
		{
			name: "error_getting_instance",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "err",
				ServiceID:  "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance err: test failure"),
		},
		{
			name: "instance_not_found",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "foo",
				ServiceID:  "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service instance foo was not found."),
		},
		{
			name: "error_describing_stack",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "err-stack",
				ServiceID:  "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to describe the CloudFormation stack err: test failure"),
		},
		{
			name: "error_getting_credentials",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
			},
			cfnOutputs: map[string]string{
				"BucketName":            "mystack-mybucket-kdwwxmddtr2g",
				"BucketAccessKeyId":     "ssm:/k8s/an-id/BucketAccessKeyId",
				"BucketSecretAccessKey": "ssm:/k8s/an-id/BucketSecretAccessKey",
			},
			ssmParams: map[string]string{
				"/k8s/an-id/BucketAccessKeyId": "foo",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the credentials from CloudFormation stack an-id: invalid parameters: [/k8s/an-id/BucketSecretAccessKey]"),
		},
		{
			name: "get_credentials",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
			},
			cfnOutputs: map[string]string{
				"BucketName":            "mystack-mybucket-kdwwxmddtr2g",
				"BucketAccessKeyId":     "ssm:/k8s/an-id/BucketAccessKeyId",
				"BucketSecretAccessKey": "ssm:/k8s/an-id/BucketSecretAccessKey",
			},
			ssmParams: map[string]string{
				"/k8s/an-id/BucketAccessKeyId":     "foo",
				"/k8s/an-id/BucketSecretAccessKey": "bar",
			},
			expectedCreds: map[string]interface{}{
				"BUCKET_NAME":              "mystack-mybucket-kdwwxmddtr2g",
				"BUCKET_ACCESS_KEY_ID":     "foo",
				"BUCKET_SECRET_ACCESS_KEY": "bar",
			},
		},
		{
			name: "get_legacy_credentials",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
			},
			cfnOutputs: map[string]string{
				"BucketName":    "mystack-mybucket-kdwwxmddtr2g",
				"UserKeyId":     "/k8s/an-id/UserKeyId",
				"UserSecretKey": "/k8s/an-id/UserSecretKey",
			},
			ssmParams: map[string]string{
				"/k8s/an-id/UserKeyId":     "foo",
				"/k8s/an-id/UserSecretKey": "bar",
			},
			expectedCreds: map[string]interface{}{
				"BUCKET_NAME":                       "mystack-mybucket-kdwwxmddtr2g",
				"TEST-SERVICE-NAME_USER_KEY_ID":     "foo",
				"TEST-SERVICE-NAME_USER_SECRET_KEY": "bar",
			},
		},
		{
			name: "unsupported_scope",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
				Parameters: map[string]interface{}{
					"RoleName": "foo",
					"Scope":    "ReadOnly",
				},
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The CloudFormation stack an-id does not support binding with scope 'ReadOnly': output not found: PolicyArnReadOnly"),
		},
		{
			name: "error_attaching_role_policy",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
				Parameters: map[string]interface{}{
					"rOlEnAmE": "exists", // Also verify that RoleName is case-insensitive
				},
			},
			cfnOutputs: map[string]string{
				"PolicyArn": "err",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to attach the policy err to role exists: test failure"),
		},
		{
			name: "attach_role_policy",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-service-id",
				Parameters: map[string]interface{}{
					"RoleName": "exists",
					"sCoPe":    "ReadWrite", // Also verify that Scope is case-insensitive
				},
			},
			cfnOutputs: map[string]string{
				"PolicyArnReadWrite": "exists",
			},
			expectedCreds: make(map[string]interface{}),
		},
		{
			name: "bind_via_lambda",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-lambda-service-id",
			},
			bindViaLambda: true,
			cfnOutputs: map[string]string{
				"SecretText": "this-is-secret",
				"BindLambda": "MyLambdaFunction",
			},
			expectedCreds: map[string]interface{}{
				"PublicText": "this-is-public",
			},
			lambdas: map[string]mockLambdaFunc{"MyLambdaFunction": func(payload []byte) ([]byte, error) {
				assert.JSONEq(t, `{"BINDING_ID":"test-binding-id","BIND_LAMBDA":"MyLambdaFunction","SECRET_TEXT":"this-is-secret","RequestType":"bind", "INSTANCE_ID": "exists"}`, string(payload))
				return []byte(`{"PublicText": "this-is-public"}`), nil
			}},
		},
		{
			name: "bind_via_lambda_missing_lambda",
			request: &osb.BindRequest{
				BindingID:  "test-binding-id",
				InstanceID: "exists",
				ServiceID:  "test-lambda-service-id",
			},
			bindViaLambda: true,
			cfnOutputs: map[string]string{
				"SecretText": "this-is-secret",
				"BindLambda": "MyLambdaFunction",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "No lambda function named MyLambdaFunction could be found."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		clients := AwsClients{
			NewCfn: func(sess *session.Session) CfnClient {
				return CfnClient{
					Client: mockCfn{
						DescribeStacksResponse: toDescribeStacksOutput(tt.cfnOutputs),
					},
				}
			},
			NewDdb: mockAwsDdbClientGetter,
			NewIam: mockAwsIamClientGetter,
			NewLambda: func(sess *session.Session) lambdaiface.LambdaAPI {
				return &mockLambda{
					lambdas: tt.lambdas,
				}
			},
			NewS3: mockAwsS3ClientGetter,
			NewSsm: func(sess *session.Session) ssmiface.SSMAPI {
				return &mockSSM{
					params: tt.ssmParams,
				}
			},
			NewSts: mockAwsStsClientGetter,
		}

		b, _ := NewAWSBroker(Options{}, mockGetAwsSession, clients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
		b.db.DataStorePort = mockDataStoreProvision{}

		resp, err := b.Bind(tt.request, &broker.RequestContext{})
		if tt.expectedErr != nil {
			assert.EqualError(t, err, tt.expectedErr.Error())
		} else if assert.NoError(t, err) {
			assert.Equal(t, tt.expectedExists, resp.Exists)
			assert.Equal(t, tt.expectedCreds, resp.Credentials)
		}
		})
	}
}

func TestUnbind(t *testing.T) {

	var callCount int
	tests := []struct {
		name        string
		request     *osb.UnbindRequest
		expectedErr error
		bindViaLambda  bool
		lambdas        map[string]mockLambdaFunc
		cfnOutputs map[string]string

	}{
		{
			name: "error_getting_binding",
			request: &osb.UnbindRequest{
				BindingID: "err",
				ServiceID: "test-service-id",				
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service binding err: test failure"),
		},
		{
			name: "binding_not_found",
			request: &osb.UnbindRequest{
				BindingID: "foo",
				ServiceID: "test-service-id",				
			},
			expectedErr: newHTTPStatusCodeError(http.StatusGone, "", "The service binding foo was not found."),
		},
		{
			name: "success",
			request: &osb.UnbindRequest{
				BindingID: "exists",
				ServiceID: "test-service-id",
			},
		},
		{
			name: "error_getting_instance",
			request: &osb.UnbindRequest{
				BindingID: "err-instance",
				ServiceID: "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance err: test failure"),
		},
		{
			name: "instance_not_found",
			request: &osb.UnbindRequest{
				BindingID: "foo-instance",
				ServiceID: "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service instance foo was not found."),
		},
		{
			name: "error_detaching_role_policy",
			request: &osb.UnbindRequest{
				BindingID: "err-role-name",
				ServiceID: "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to detach the policy exists from role err: test failure"),
		},
		{
			name: "detach_role_policy",
			request: &osb.UnbindRequest{
				BindingID: "exists-role-name",
				ServiceID: "test-service-id",
			},
		},
		{
			name: "role_not_found",
			request: &osb.UnbindRequest{
				BindingID: "foo-role-name",
				ServiceID: "test-service-id",				
			},
		},
		{
			name: "unbind_via_lambda",
			request: &osb.UnbindRequest{
				BindingID: "foo-role-name",
				ServiceID: "test-lambda-service-id",				
			},
			bindViaLambda: true,
			lambdas: map[string]mockLambdaFunc{"MyLambdaFunction": func(payload []byte) ([]byte, error) {
				callCount++
				var params map[string]string
				err := json.Unmarshal(payload, &params)
				if err != nil {
					return nil, err
				}
				assert.Equal(t, params["RequestType"], "unbind")
				assert.Equal(t, params["INSTANCE_ID"], "exists")
				assert.Equal(t, params["BINDING_ID"], "foo-role-name")
				return nil, nil
			}},
			cfnOutputs: map[string]string{
				"SecretText": "this-is-secret",
				"BindLambda": "MyLambdaFunction",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount = 0
			clients := AwsClients{
				NewCfn: func(sess *session.Session) CfnClient {
					return CfnClient{
						Client: mockCfn{
							DescribeStacksResponse: toDescribeStacksOutput(tt.cfnOutputs),
						},
					}
				},
				NewDdb: mockAwsDdbClientGetter,
				NewIam: mockAwsIamClientGetter,
				NewLambda: func(sess *session.Session) lambdaiface.LambdaAPI {
					return &mockLambda{
						lambdas: tt.lambdas,
					}
				},
				NewS3: mockAwsS3ClientGetter,
				NewSsm: mockAwsSsmClientGetter,
				NewSts: mockAwsStsClientGetter,
			}
			
			b, _ := NewAWSBroker(Options{}, mockGetAwsSession, clients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
			b.db.DataStorePort = mockDataStoreProvision{}
			_, err := b.Unbind(tt.request, &broker.RequestContext{})
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			if tt.bindViaLambda {
				assert.Equal(t, 1, callCount)
			} else {
				assert.Equal(t, 0, callCount)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name          string
		request       *osb.UpdateInstanceRequest
		expectedAsync bool
		expectedErr   error
	}{
		{
			name: "async_required",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: false,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
			},
			expectedErr: newAsyncError(),
		},
		{
			name: "error_getting_instance",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "err",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service instance \"err\": test failure"),
		},
		{
			name: "instance_not_found",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "foo",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service instance \"foo\" was not found."),
		},
		{
			name: "change_plan",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
				PlanID:            aws.String("new-plan-id"),
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service plan cannot be changed from \"test-plan-id\" to \"new-plan-id\"."),
		},
		{
			name: "error_getting_service",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "err",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to get the service \"err\": test failure"),
		},
		{
			name: "service_not_found",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "foo",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service \"foo\" was not found."),
		},
		{
			name: "plan_not_found",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "foo-plan",
				ServiceID:         "test-service-id",
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The service plan \"foo\" was not found."),
		},
		{
			name: "parameter_not_updatable",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
				Parameters:        map[string]interface{}{"foo": "bar"},
			},
			expectedErr: newHTTPStatusCodeError(http.StatusBadRequest, "", "The parameter \"foo\" is not updatable."),
		},
		{
			name: "parameter_not_updated",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
				Parameters:        map[string]interface{}{"req_param": "a-value"},
			},
			expectedAsync: false,
		},
		{
			name: "error_updating_stack",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "err-stack",
				ServiceID:         "test-service-id",
				Parameters:        map[string]interface{}{"req_param": "new-value"},
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to update the CloudFormation stack \"err\": test failure"),
		},
		{
			name: "success",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
				Parameters:        map[string]interface{}{"req_param": "new-value"},
			},
			expectedAsync: true,
		},
		{
			name: "error_updating_instance",
			request: &osb.UpdateInstanceRequest{
				AcceptsIncomplete: true,
				InstanceID:        "exists",
				ServiceID:         "test-service-id",
				Parameters:        map[string]interface{}{"req_param": "err"},
			},
			expectedErr: newHTTPStatusCodeError(http.StatusInternalServerError, "", "Failed to update the service instance \"exists\": test failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := NewAWSBroker(Options{}, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
			b.db.DataStorePort = mockDataStoreProvision{}

			resp, err := b.Update(tt.request, &broker.RequestContext{})
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else if assert.NoError(t, err) {
				assert.Equal(t, tt.expectedAsync, resp.Async)
			}
		})
	}
}
