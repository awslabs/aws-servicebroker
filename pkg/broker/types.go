package broker

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/awslabs/aws-servicebroker/pkg/serviceinstance"
	"github.com/koding/cache"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	uuid "github.com/satori/go.uuid"
)

// Options cli options
type Options struct {
	CatalogPath        string
	KeyID              string
	SecretKey          string
	Profile            string
	TableName          string
	S3Bucket           string
	S3Region           string
	S3Key              string
	TemplateFilter     string
	Region             string
	BrokerID           string
	RoleArn            string
	PrescribeOverrides bool
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
	accountId          string
	keyid              string
	secretkey          string
	profile            string
	tablename          string
	s3bucket           string
	s3region           string
	s3key              string
	templatefilter     string
	region             string
	partition          string
	s3svc              S3Client
	ssmsvc             ssm.SSM
	catalogcache       cache.Cache
	listingcache       cache.Cache
	instances          map[string]*serviceinstance.ServiceInstance
	brokerid           string
	db                 Db
	GetSession         GetAwsSession
	Clients            AwsClients
	prescribeOverrides bool
	globalOverrides    map[string]string
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
	PutServiceDefinition(sd osb.Service) error
	GetParam(paramname string) (value string, err error)
	PutParam(paramname string, paramvalue string) error
	GetServiceDefinition(serviceuuid string) (*osb.Service, error)
	GetServiceInstance(sid string) (*serviceinstance.ServiceInstance, error)
	PutServiceInstance(si serviceinstance.ServiceInstance) error
	DeleteServiceInstance(sid string) error
	GetServiceBinding(id string) (*serviceinstance.ServiceBinding, error)
	PutServiceBinding(sb serviceinstance.ServiceBinding) error
	DeleteServiceBinding(id string) error
}

type GetAwsSession func(keyid string, secretkey string, region string, accountId string, profile string, params map[string]string) *session.Session

type GetCfnClient func(sess *session.Session) CfnClient
type GetSsmClient func(sess *session.Session) ssmiface.SSMAPI
type GetS3Client func(sess *session.Session) S3Client
type GetDdbClient func(sess *session.Session) *dynamodb.DynamoDB
type GetStsClient func(sess *session.Session) *sts.STS
type GetIamClient func(sess *session.Session) iamiface.IAMAPI
type GetLambdaClient func(sess *session.Session) lambdaiface.LambdaAPI

type AwsClients struct {
	NewCfn    GetCfnClient
	NewSsm    GetSsmClient
	NewS3     GetS3Client
	NewDdb    GetDdbClient
	NewSts    GetStsClient
	NewIam    GetIamClient
	NewLambda GetLambdaClient
}

type S3Client struct {
	Client s3iface.S3API
}

type CfnClient struct {
	Client cloudformationiface.CloudFormationAPI
}

type AwsTags []struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type GetCallerIder func(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error)
type UpdateCataloger func(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, listTemplates ListTemplateser, listingUpdate ListingUpdater, metadataUpdate MetadataUpdater) error
type PollUpdater func(interval int, l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, updateCatalog UpdateCataloger, listTemplates ListTemplateser)
type ListTemplateser func(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error)
type ListingUpdater func(l *[]ServiceLastUpdate, c cache.Cache) error
type MetadataUpdater func(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater, templatefilter string) error

type CfnTemplate struct {
	Description string `yaml:"Description,omitempty"`
	Parameters  map[string]struct {
		Description   string   `yaml:"Description,omitempty"`
		Type          string   `yaml:"Type,omitempty"`
		Default       *string  `yaml:"Default,omitempty"`
		AllowedValues []string `yaml:"AllowedValues,omitempty"`
	} `yaml:"Parameters,omitempty"`
	Outputs map[string]struct {
		Description string `yaml:"Description,omitempty"`
	} `yaml:"Outputs,omitempty"`
	Metadata struct {
		Spec struct {
			Version             string   `yaml:"Version,omitempty"`
			Tags                []string `yaml:"Tags,omitempty"`
			Name                string   `yaml:"Name,omitempty"`
			DisplayName         string   `yaml:"DisplayName,omitempty"`
			LongDescription     string   `yaml:"LongDescription,omitempty"`
			ImageUrl            string   `yaml:"ImageUrl,omitempty"`
			DocumentationUrl    string   `yaml:"DocumentationUrl,omitempty"`
			ProviderDisplayName string   `yaml:"ProviderDisplayName,omitempty"`
			OutputsAsIs         bool     `yaml:"OutputsAsIs,omitempty"`
			CloudFoundry        bool     `yaml:"CloudFoundry,omitempty"`
			BindViaLambda       bool     `yaml:"BindViaLambda"`
			Bindings            struct {
				IAM struct {
					AddKeypair bool `yaml:"AddKeypair,omitempty"`
					Policies   []struct {
						PolicyDocument map[string]interface{} `yaml:"PolicyDocument,omitempty"`
					} `yaml:"Policies,omitempty"`
				} `yaml:"IAM,omitempty"`
				CFNOutputs []string `yaml:"CFNOutputs,omitempty"`
			} `yaml:"Bindings,omitempty"`
			ServicePlans map[string]struct {
				DisplayName       string            `yaml:"DisplayName,omitempty"`
				Description       string            `yaml:"Description,omitempty"`
				LongDescription   string            `yaml:"LongDescription,omitempty"`
				Cost              string            `yaml:"Cost,omitempty"`
				ParameterValues   map[string]string `yaml:"ParameterValues,omitempty"`
				ParameterDefaults map[string]string `yaml:"ParameterDefaults,omitempty"`
			} `yaml:"ServicePlans,omitempty"`
			UpdatableParameters []string `yaml:"UpdatableParameters,omitempty"`
		} `yaml:"AWS::ServiceBroker::Specification,omitempty"`
		Interface struct {
			ParameterGroups []struct {
				Label struct {
					Name string `yaml:"default,omitempty"`
				} `yaml:"Label,omitempty"`
				Parameters []string `yaml:"Parameters,omitempty"`
			} `yaml:"ParameterGroups,omitempty"`
			ParameterLabels map[string]struct {
				Label string `yaml:"default,omitempty"`
			} `yaml:"ParameterLabels,omitempty"`
		} `yaml:"AWS::CloudFormation::Interface,omitempty"`
	} `yaml:"Metadata,omitempty"`
}

type OpenshiftFormDefinition struct {
	Type  string
	Title string
	Items []string
}
