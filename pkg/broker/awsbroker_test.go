package broker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/awslabs/aws-servicebroker/pkg/serviceinstance"
	"github.com/koding/cache"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

type TestCases map[string]Options

func (T *TestCases) GetTests(f string) error {
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &T)
	if err != nil {
		log.Printf("Unmarshal: %v", err)
		return err
	}
	return nil
}

func mockGetAwsSession(keyid string, secretkey string, region string, accountID string, profile string, params map[string]string) *session.Session {
	sess := mock.Session
	conf := aws.NewConfig()
	conf.Region = aws.String(region)
	return sess.Copy(conf)
}

func mockAwsCfnClientGetter(sess *session.Session) CfnClient {
	return CfnClient{mockCfn{
		DescribeStacksResponse: cloudformation.DescribeStacksOutput{},
	}}
}

func mockAwsStsClientGetter(sess *session.Session) *sts.STS {
	conf := aws.NewConfig()
	conf.Region = sess.Config.Region
	return &sts.STS{Client: mock.NewMockClient(conf)}
}

func mockAwsS3ClientGetter(sess *session.Session) S3Client {
	conf := aws.NewConfig()
	conf.Region = sess.Config.Region
	return S3Client{s3iface.S3API(&s3.S3{Client: mock.NewMockClient(conf)})}
}

func mockAwsDdbClientGetter(sess *session.Session) *dynamodb.DynamoDB {
	conf := aws.NewConfig()
	conf.Region = sess.Config.Region
	return &dynamodb.DynamoDB{Client: mock.NewMockClient(conf)}
}

type mockIAM struct {
	iamiface.IAMAPI
}

func (c *mockIAM) AttachRolePolicy(input *iam.AttachRolePolicyInput) (*iam.AttachRolePolicyOutput, error) {
	if aws.StringValue(input.RoleName) != "exists" || aws.StringValue(input.PolicyArn) != "exists" {
		return nil, errors.New("test failure")
	}
	return &iam.AttachRolePolicyOutput{}, nil
}

func (c *mockIAM) DetachRolePolicy(input *iam.DetachRolePolicyInput) (*iam.DetachRolePolicyOutput, error) {
	if aws.StringValue(input.RoleName) == "err" || aws.StringValue(input.PolicyArn) == "err" {
		return nil, errors.New("test failure")
	} else if aws.StringValue(input.RoleName) == "exists" && aws.StringValue(input.PolicyArn) == "exists" {
		return &iam.DetachRolePolicyOutput{}, nil
	}
	return nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", nil)
}

func mockAwsIamClientGetter(sess *session.Session) iamiface.IAMAPI {
	return &mockIAM{}
}

type mockLambdaFunc func(payload []byte) ([]byte, error)

type mockLambda struct {
	lambdaiface.LambdaAPI
	lambdas map[string]mockLambdaFunc
}

func (ml *mockLambda) Invoke(ii *lambda.InvokeInput) (*lambda.InvokeOutput, error) {
	f, found := ml.lambdas[aws.StringValue(ii.FunctionName)]
	if !found {
		return nil, fmt.Errorf("No lambda function named %s could be found.", aws.StringValue(ii.FunctionName))
	}
	result, err := f(ii.Payload)
	if err != nil {
		return nil, fmt.Errorf("Error in Lambda function: %s", err.Error())
	}
	return &lambda.InvokeOutput{Payload: result}, nil
}

func mockAwsLambdaClientGetter(sess *session.Session) lambdaiface.LambdaAPI {
	return &mockLambda{}
}

type mockSSM struct {
	ssmiface.SSMAPI
	params map[string]string
}

func (c *mockSSM) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	output := ssm.GetParametersOutput{}
	for _, n := range input.Names {
		if aws.StringValue(n) == "err" {
			return nil, errors.New("test failure")
		} else if v, ok := c.params[aws.StringValue(n)]; ok {
			output.Parameters = append(output.Parameters, &ssm.Parameter{Name: n, Value: &v})
		} else {
			output.InvalidParameters = append(output.InvalidParameters, n)
		}
	}
	return &output, nil
}

func mockAwsSsmClientGetter(sess *session.Session) ssmiface.SSMAPI {
	return &mockSSM{}
}

var mockClients = AwsClients{
	NewCfn:    mockAwsCfnClientGetter,
	NewDdb:    mockAwsDdbClientGetter,
	NewIam:    mockAwsIamClientGetter,
	NewLambda: mockAwsLambdaClientGetter,
	NewS3:     mockAwsS3ClientGetter,
	NewSsm:    mockAwsSsmClientGetter,
	NewSts:    mockAwsStsClientGetter,
}

func mockGetAccountID(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error) {
	return &sts.GetCallerIdentityOutput{Account: aws.String("123456789012")}, nil
}

func mockGetAccountIDFail(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error) {
	return &sts.GetCallerIdentityOutput{}, errors.New("I should be failing")
}

func mockUpdateCatalog(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, listTemplates ListTemplateser, listingUpdate ListingUpdater, metadataUpdate MetadataUpdater) error {
	return nil
}

func mockUpdateCatalogFail(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, listTemplates ListTemplateser, listingUpdate ListingUpdater, metadataUpdate MetadataUpdater) error {
	return errors.New("I failed")
}

func mockPollUpdate(interval int, l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, updateCatalog UpdateCataloger, listTemplates ListTemplateser) {

}

// mock implementation of DataStore Adapter
type mockDataStore struct{}

func (db mockDataStore) PutServiceDefinition(sd osb.Service) error { return nil }
func (db mockDataStore) GetParam(paramname string) (value string, err error) {
	return "some-value", nil
}
func (db mockDataStore) PutParam(paramname string, paramvalue string) error          { return nil }
func (db mockDataStore) PutServiceInstance(si serviceinstance.ServiceInstance) error { return nil }
func (db mockDataStore) GetServiceDefinition(serviceuuid string) (*osb.Service, error) {
	service := osb.Service{
		ID:                  "",
		Name:                "",
		Description:         "",
		Tags:                nil,
		Requires:            nil,
		Bindable:            false,
		BindingsRetrievable: false,
		PlanUpdatable:       nil,
		Plans:               nil,
		DashboardClient: &osb.DashboardClient{
			ID:          "",
			Secret:      "",
			RedirectURI: "",
		},
		Metadata: nil,
	}
	return &service, nil
}
func (db mockDataStore) GetServiceInstance(sid string) (*serviceinstance.ServiceInstance, error) {
	si := serviceinstance.ServiceInstance{
		ID:        "",
		ServiceID: "",
		PlanID:    "",
		Params:    nil,
		StackID:   "",
	}
	return &si, nil
}
func (db mockDataStore) DeleteServiceInstance(sid string) error { return nil }
func (db mockDataStore) GetServiceBinding(id string) (*serviceinstance.ServiceBinding, error) {
	return nil, nil
}
func (db mockDataStore) PutServiceBinding(sb serviceinstance.ServiceBinding) error { return nil }
func (db mockDataStore) DeleteServiceBinding(id string) error                      { return nil }

func TestNewAwsBroker(t *testing.T) {
	assert := assert.New(t)
	options := new(TestCases)
	options.GetTests("../../testcases/options.yaml")

	for _, v := range *options {
		// Shouldn't error
		bl, err := NewAWSBroker(v, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
		assert.Nil(err)

		// check values are as expected
		assert.Equal(v.KeyID, bl.keyid)
		assert.Equal(v.SecretKey, bl.secretkey)
		assert.Equal(v.Profile, bl.secretkey)
		assert.Equal(v.Profile, bl.profile)
		assert.Equal(v.TableName, bl.tablename)
		assert.Equal(v.S3Bucket, bl.s3bucket)
		assert.Equal(v.S3Region, bl.s3region)
		assert.Equal(addTrailingSlash(v.S3Key), bl.s3key)
		assert.Equal(v.TemplateFilter, bl.templatefilter)
		assert.Equal(v.Region, bl.region)
		assert.Equal(v.BrokerID, bl.brokerid)
		assert.Equal("123456789012", bl.db.Accountid)
		assert.Equal(uuid.NewV5(uuid.NullUUID{}.UUID, "123456789012"+v.BrokerID), bl.db.Accountuuid)
		assert.Equal(v.BrokerID, bl.db.Brokerid)

		// Should error
		_, err = NewAWSBroker(v, mockGetAwsSession, mockClients, mockGetAccountIDFail, mockUpdateCatalog, mockPollUpdate)
		assert.Error(err)

		// Should error
		_, err = NewAWSBroker(v, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalogFail, mockPollUpdate)
		assert.Error(err)
	}
}

func mockListTemplates(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error) {
	return &[]ServiceLastUpdate{}, nil
}

func mockListTemplatesFailNoBucket(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error) {
	return &[]ServiceLastUpdate{}, errors.New("NoSuchBucket: The specified bucket does not exist")
}

func mockListTemplatesFail(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error) {
	return &[]ServiceLastUpdate{}, errors.New("ListTemplates failed")
}

func mockListingUpdate(l *[]ServiceLastUpdate, c cache.Cache) error {
	return nil
}

func mockListingUpdateFail(l *[]ServiceLastUpdate, c cache.Cache) error {
	return errors.New("ListingUpdate failed")
}

func mockMetadataUpdate(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater, templatefilter string) error {
	return nil
}

func mockMetadataUpdateFail(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater, templatefilter string) error {
	return errors.New("MetadataUpdate failed")
}

func TestUpdateCatalog(t *testing.T) {
	assert := assert.New(t)
	options := new(TestCases)
	options.GetTests("../../testcases/options.yaml")
	var bl *AwsBroker
	var bd *BucketDetailsRequest
	for _, v := range *options {
		bl, _ = NewAWSBroker(v, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
		bd = &BucketDetailsRequest{
			v.S3Bucket,
			v.S3Key,
			v.TemplateFilter,
		}
	}

	bl.db.DataStorePort = mockDataStore{}

	err := UpdateCatalog(bl.listingcache, bl.catalogcache, *bd, bl.s3svc, bl.db, *bl, mockListTemplates, mockListingUpdate, mockMetadataUpdate)
	assert.Nil(err)

	err = UpdateCatalog(bl.listingcache, bl.catalogcache, *bd, bl.s3svc, bl.db, *bl, mockListTemplatesFailNoBucket, mockListingUpdate, mockMetadataUpdate)
	assert.EqualError(err, "Cannot access S3 Bucket, either it does not exist or the IAM user/role the broker is configured to use has no access to the bucket")

	err = UpdateCatalog(bl.listingcache, bl.catalogcache, *bd, bl.s3svc, bl.db, *bl, mockListTemplatesFail, mockListingUpdate, mockMetadataUpdate)
	assert.EqualError(err, "ListTemplates failed")

	err = UpdateCatalog(bl.listingcache, bl.catalogcache, *bd, bl.s3svc, bl.db, *bl, mockListTemplates, mockListingUpdateFail, mockMetadataUpdate)
	assert.EqualError(err, "ListingUpdate failed")

	err = UpdateCatalog(bl.listingcache, bl.catalogcache, *bd, bl.s3svc, bl.db, *bl, mockListTemplates, mockListingUpdate, mockMetadataUpdateFail)
	assert.EqualError(err, "MetadataUpdate failed")
}

type mockS3 struct {
	s3iface.S3API
	GetObjectResp s3.GetObjectOutput
}

func (m mockS3) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &m.GetObjectResp, nil
}

type mockCfn struct {
	cloudformationiface.CloudFormationAPI
	DescribeStacksResponse      cloudformation.DescribeStacksOutput
	DescribeStackEventsResponse cloudformation.DescribeStackEventsOutput
	CreateStackResponse         cloudformation.CreateStackOutput
	DeleteStackResponse         cloudformation.DeleteStackOutput
	UpdateStackResponse         cloudformation.UpdateStackOutput
}

func (m mockCfn) DescribeStacks(in *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if aws.StringValue(in.StackName) == "err" {
		return nil, errors.New("test failure")
	}
	return &m.DescribeStacksResponse, nil
}

func (m mockCfn) DescribeStackEvents(in *cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error) {
	if aws.StringValue(in.StackName) == "err" {
		return nil, errors.New("test failure")
	}
	return &m.DescribeStackEventsResponse, nil
}

func (m mockCfn) CreateStack(in *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return &m.CreateStackResponse, nil
}

func (m mockCfn) DeleteStack(in *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	if aws.StringValue(in.StackName) == "err" {
		return nil, errors.New("test failure")
	}
	return &m.DeleteStackResponse, nil
}

func (m mockCfn) UpdateStack(in *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	if aws.StringValue(in.StackName) == "err" {
		return nil, errors.New("test failure")
	}
	return &m.UpdateStackResponse, nil
}

func (m mockCfn) CancelUpdateStack(in *cloudformation.CancelUpdateStackInput) (*cloudformation.CancelUpdateStackOutput, error) {
	return &cloudformation.CancelUpdateStackOutput{}, nil
}

func TestMetadataUpdate(t *testing.T) {
	assert := assert.New(t)
	options := new(TestCases)
	options.GetTests("../../testcases/options.yaml")
	var bl *AwsBroker
	var bd *BucketDetailsRequest
	for _, v := range *options {
		bl, _ = NewAWSBroker(v, mockGetAwsSession, mockClients, mockGetAccountID, mockUpdateCatalog, mockPollUpdate)
		bd = &BucketDetailsRequest{
			v.S3Bucket,
			v.S3Key,
			v.TemplateFilter,
		}
	}
	bl.db.DataStorePort = mockDataStore{}

	s3svc := S3Client{
		Client: mockS3{GetObjectResp: s3.GetObjectOutput{}},
	}

	// test "__LISTINGS__" not in cache
	err := MetadataUpdate(bl.listingcache, bl.catalogcache, *bd, s3svc, bl.db, MetadataUpdate, "-main.yaml")
	assert.EqualError(err, "not found")

	// test empty s3 body
	var serviceUpdates []ServiceNeedsUpdate
	serviceUpdates = append(serviceUpdates, ServiceNeedsUpdate{
		Name:   "test-service",
		Update: true,
	})
	bl.listingcache.Set("__LISTINGS__", serviceUpdates)
	err = MetadataUpdate(bl.listingcache, bl.catalogcache, *bd, s3svc, bl.db, MetadataUpdate, "-main.yaml")
	assert.Equal(err, nil, "should handle empty s3 objects without erroring")

	// test object not yaml
	s3obj := s3.GetObjectOutput{Body: ioutil.NopCloser(strings.NewReader("test"))}
	s3svc = S3Client{
		Client: mockS3{GetObjectResp: s3obj},
	}
	err = MetadataUpdate(bl.listingcache, bl.catalogcache, *bd, s3svc, bl.db, MetadataUpdate, "-main.yaml")
	assert.Equal(err, nil, "should handle bad templates without erroring")

	// TODO: test success and more failure scenarios
}

func TestAssumeArnGeneration(t *testing.T) {
	params := map[string]string{"target_role_name": "worker"}

	// AWS Standard Partition
	partition := "aws"
	accountID := "123456654321"
	assert.Equal(t, generateRoleArn(params, accountID, partition), "arn:aws:iam::123456654321:role/worker", "Validate role arn")
	params["target_account_id"] = "000000000000"
	assert.Equal(t, generateRoleArn(params, accountID, partition), "arn:aws:iam::000000000000:role/worker", "Validate role arn")

}

func TestAssumeArnGenerationChinaPartition(t *testing.T) {
	params := map[string]string{"target_role_name": "worker"}

	// AWS China Partition
	partition := "aws-cn"
	accountID := "123456654321"
	assert.Equal(t, generateRoleArn(params, accountID, partition), "arn:aws-cn:iam::123456654321:role/worker", "Validate role arn")
	params["target_account_id"] = "000000000000"
	assert.Equal(t, generateRoleArn(params, accountID, partition), "arn:aws-cn:iam::000000000000:role/worker", "Validate role arn")
}
