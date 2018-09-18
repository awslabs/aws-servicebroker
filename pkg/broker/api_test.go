package broker

import (
	"errors"
	"github.com/awslabs/aws-service-broker/pkg/serviceinstance"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
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
	return nil
}
func (db mockDataStoreProvision) GetServiceDefinition(serviceuuid string) (*osb.Service, error) {
	if serviceuuid == "test-service-id" {
		return &osb.Service{
			ID:   "test-service-id",
			Name: "test-service-name",
			Plans: []osb.Plan{
				{ID: "test-plan-id", Name: "test-plan-name", Schemas: &osb.Schemas{ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string", "required": true},
							"override_param": map[string]interface{}{"type": "string"},
							"region":         map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []interface{}{"req_param"},
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
	if sid == "err" {
		return nil, errors.New("test failure")
	} else if sid == "exists" {
		return &serviceinstance.ServiceInstance{StackID: "an-id"}, nil
	}
	return nil, nil
}
func (db mockDataStoreProvision) GetServiceBinding(id string) (*serviceinstance.ServiceBinding, error) {
	if id == "exists" {
		return &serviceinstance.ServiceBinding{
			ID:         "exists",
			InstanceID: "exists",
		}, nil
	}
	return nil, nil
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
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
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
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}

	deprovReq := &osb.DeprovisionRequest{
		InstanceID:        "test-instance-id",
		AcceptsIncomplete: true,
	}
	reqContext := &broker.RequestContext{}

	expected := &broker.DeprovisionResponse{}
	actual, err := bl.Deprovision(deprovReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")

	bl.accountId = "test"
	bl.secretkey = "testkey"

	deprovReq.InstanceID = "exists"
	expected.Async = true
	actual, err = bl.Deprovision(deprovReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")

}

func TestLastOperation(t *testing.T) {
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

	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}

	loReq := &osb.LastOperationRequest{InstanceID: "test-instance-id"}
	reqContext := &broker.RequestContext{}
	msg := "CloudFormation stackid missing, chances are stack creation failed in an unexpected way"
	expected := &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "failed", Description: &msg}}
	actual, err := bl.LastOperation(loReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")

	mockClients.NewCfn = func(sess *session.Session) CfnClient {
		return CfnClient{mockCfn{
			DescribeStacksResponse: cloudformation.DescribeStacksOutput{
				NextToken: nil,
				Stacks: []*cloudformation.Stack{
					{
						StackStatus: aws.String("CREATE_IN_PROGRESS"),
					},
				},
			},
		}}
	}
	bl, _ = NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}
	expected = &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "in progress", Description: nil}}
	loReq.InstanceID = "exists"
	actual, err = bl.LastOperation(loReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")

	mockClients.NewCfn = func(sess *session.Session) CfnClient {
		return CfnClient{mockCfn{
			DescribeStacksResponse: cloudformation.DescribeStacksOutput{
				NextToken: nil,
				Stacks: []*cloudformation.Stack{
					{
						StackStatus: aws.String("CREATE_FAILED"),
					},
				},
			},
		}}
	}
	bl, _ = NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}
	expected = &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "failed", Description: nil}}
	loReq.InstanceID = "exists"
	actual, err = bl.LastOperation(loReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")

	mockClients.NewCfn = func(sess *session.Session) CfnClient {
		return CfnClient{mockCfn{
			DescribeStacksResponse: cloudformation.DescribeStacksOutput{
				NextToken: nil,
				Stacks: []*cloudformation.Stack{
					{
						StackStatus: aws.String("CREATE_COMPLETE"),
					},
				},
			},
		}}
	}
	bl, _ = NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}
	expected = &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "succeeded", Description: nil}}
	loReq.InstanceID = "exists"
	actual, err = bl.LastOperation(loReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed even if stack is not in serviceInstance (was never created)")
}

func TestBind(t *testing.T) {
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
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}

	bindReq := &osb.BindRequest{
		BindingID:         "test-bind-id",
		InstanceID:        "exists",
		AcceptsIncomplete: true,
		ServiceID:         "test-service-id",
	}
	reqContext := &broker.RequestContext{}

	expected := &broker.BindResponse{BindResponse: osb.BindResponse{Credentials: map[string]interface{}{}}}
	actual, err := bl.Bind(bindReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed")

}

func TestUnbind(t *testing.T) {
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
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.db.DataStorePort = mockDataStoreProvision{}

	unbindReq := &osb.UnbindRequest{BindingID: "exists"}
	reqContext := &broker.RequestContext{}

	expected := &broker.UnbindResponse{UnbindResponse: osb.UnbindResponse{}}
	actual, err := bl.Unbind(unbindReq, reqContext)
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should succeed")

}
