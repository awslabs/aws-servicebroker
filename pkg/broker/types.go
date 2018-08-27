package broker

import (
	"github.com/awslabs/aws-service-broker/pkg/serviceinstance"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/koding/cache"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

// Options cli options
type Options struct {
	CatalogPath    string
	KeyID          string
	SecretKey      string
	Profile        string
	TableName      string
	S3Bucket       string
	S3Region       string
	S3Key          string
	TemplateFilter string
	Region         string
	BrokerID       string
	RoleArn        string
}

// BucketDetailsRequest describes the details required to fetch metadata and templates from s3
type BucketDetailsRequest struct {
	bucket string
	prefix string
	suffix string
}

// AwsBroker holds configuration, caches and aws service clients
type AwsBroker struct {
	sync.RWMutex
	accountId      string
	keyid          string
	secretkey      string
	profile        string
	tablename      string
	s3bucket       string
	s3region       string
	s3key          string
	templatefilter string
	region         string
	s3svc          S3Client
	ssmsvc         ssm.SSM
	catalogcache   cache.Cache
	listingcache   cache.Cache
	instances      map[string]*serviceinstance.ServiceInstance
	brokerid       string
	db             Db
	GetSession     GetAwsSession
	Clients        AwsClients
}

// ServiceNeedsUpdate if Update == true the metadata should be refreshed from s3
type ServiceNeedsUpdate struct {
	Name   string
	Update bool
}

// ServiceLastUpdate date when a service exposed by the broker was last updated from s3
type ServiceLastUpdate struct {
	Name string
	Date time.Time
}

// Db configuration
type Db struct {
	Accountid     string
	Accountuuid   uuid.UUID
	Brokerid      string
	DataStorePort DataStore
}

// DataStore port, any backend datastore must provide at least these interfaces
type DataStore interface {
	Lock(lockname string) bool
	IsLocked(lockname string) bool
	WaitForUnlock(lockname string) bool
	Unlock(lockname string) error
	PutServiceDefinition(sd osb.Service) error
	GetParam(paramname string) (value string, err error)
	PutParam(paramname string, paramvalue string) error
	GetServiceDefinition(serviceuuid string) (osb.Service, error)
	GetServiceInstance(sid string) (serviceinstance.ServiceInstance, error)
	PutServiceInstance(si serviceinstance.ServiceInstance) error
}

type GetAwsSession func(keyid string, secretkey string, region string, accountId string, profile string, params map[string]string) *session.Session

type GetCfnClient func(sess *session.Session) *cloudformation.CloudFormation
type GetSsmClient func(sess *session.Session) *ssm.SSM
type GetS3Client func(sess *session.Session) S3Client
type GetDdbClient func(sess *session.Session) *dynamodb.DynamoDB
type GetStsClient func(sess *session.Session) *sts.STS

type AwsClients struct {
	NewCfn GetCfnClient
	NewSsm GetSsmClient
	NewS3  GetS3Client
	NewDdb GetDdbClient
	NewSts GetStsClient
}

type S3Client struct {
	Client s3iface.S3API
}

type GetCallerIder func(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error)
type UpdateCataloger func(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, listTemplates ListTemplateser, listingUpdate ListingUpdater, metadataUpdate MetadataUpdater) error
type PollUpdater func(interval int, l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, updateCatalog UpdateCataloger, listTemplates ListTemplateser)
type ListTemplateser func(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error)
type ListingUpdater func(l *[]ServiceLastUpdate, c cache.Cache) error
type MetadataUpdater func(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater) error
