package broker

import (
	"net/http"
	"sync"

	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	dbConnector "github.com/awslabs/aws-service-broker/pkg/db"
	"github.com/awslabs/aws-service-broker/pkg/dynamodbadapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang/glog"
	"github.com/koding/cache"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
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

// BusinessLogic holds configuration, caches and aws service clients
type BusinessLogic struct {
	sync.RWMutex
	keyid          string
	secretkey      string
	profile        string
	tablename      string
	s3bucket       string
	s3region       string
	s3key          string
	templatefilter string
	region         string
	s3svc          s3.S3
	ssmsvc         ssm.SSM
	catalogcache   cache.Cache
	listingcache   cache.Cache
	instances      map[string]*dbConnector.ServiceInstance
	brokerid       string
	db             dbConnector.Db
	rolearn        string
	overrides      map[string]string
}

// CacheTTL TTL for catalog cache record expiry
var CacheTTL = 1 * time.Hour

// Create AWS Session
func createSession(keyid string, secretkey string, region string, profile string) *session.Session {
	var defaultCreds credentials.Credentials
	if keyid != "" {
		defaultCreds = *credentials.NewStaticCredentials(keyid, secretkey, "")
	} else if profile != "" {
		defaultCreds = *credentials.NewChainCredentials([]credentials.Provider{&credentials.SharedCredentialsProvider{Profile: profile}})
	} else {
		defaultCreds = *credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.Must(session.NewSession()))},
			})
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: &defaultCreds,
	}))
	return sess
}

// NewBusinessLogic runs at broker startup and bootstraps the broker
func NewBusinessLogic(o Options) (*BusinessLogic, error) {

	sess := createSession(o.KeyID, o.SecretKey, o.Region, o.Profile)
	s3sess := createSession(o.KeyID, o.SecretKey, o.S3Region, o.Profile)
	var s3svc = s3.New(s3sess)
	var ddbsvc = dynamodb.New(sess)
	stsClient := sts.New(sess)
	callerid, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		panic(err)
	}
	accountid := *callerid.Account
	accountuuid := uuid.NewV5(uuid.NullUUID{}.UUID, accountid+o.BrokerID)

	var db dbConnector.Db
	db.Brokerid = o.BrokerID
	db.Accountid = accountid
	db.Accountuuid = accountuuid

	// connect DynamoDB  adapter to storage port
	db.DataStorePort = dynamodbadapter.DdbDataStore{
		Accountid:   accountid,
		Accountuuid: accountuuid,
		Brokerid:    o.BrokerID,
		Region:      o.Region,
		Ddb:         *ddbsvc,
		Tablename:   o.TableName,
	}

	var catalogcache = cache.NewMemoryWithTTL(time.Duration(CacheTTL))
	var listingcache = cache.NewMemoryWithTTL(time.Duration(CacheTTL))
	var overrides = make(map[string]string)

	for _, item := range os.Environ() {
		envvar := strings.Split(item, "=")
		if strings.HasPrefix(envvar[0], "PARAM_OVERRIDE_") {
			key := strings.TrimPrefix(envvar[0], "PARAM_OVERRIDE_")
			overrides[key] = envvar[1]
			glog.V(10).Infof("%q=%q\n", key, envvar[1])
		}
	}

	listingcache.StartGC(time.Minute * 5)
	bd := &BucketDetailsRequest{
		o.S3Bucket,
		o.S3Key,
		o.TemplateFilter,
	}
	s3key := o.S3Key
	if strings.HasSuffix(o.S3Key, "/") == false {
		s3key = s3key + "/"
	}
	bl := BusinessLogic{
		keyid:          o.KeyID,
		secretkey:      o.SecretKey,
		profile:        o.Profile,
		tablename:      o.TableName,
		s3bucket:       o.S3Bucket,
		s3region:       o.S3Region,
		s3key:          s3key,
		templatefilter: o.TemplateFilter,
		region:         o.Region,
		s3svc:          *s3svc,
		catalogcache:   catalogcache,
		listingcache:   listingcache,
		brokerid:       o.BrokerID,
		db:             db,
		rolearn:        o.RoleArn,
		overrides:      overrides,
	}
	updateCatalog(listingcache, catalogcache, *bd, *s3svc, db, bl)
	go pollUpdate(600, listingcache, catalogcache, *bd, *s3svc, db, bl)
	return &bl, nil
}

func updateCatalog(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc s3.S3, db dbConnector.Db, bl BusinessLogic) {
	l, err := listTemplates(&bd, &bl)
	if err != nil {
		if strings.HasPrefix(err.Error(), "NoSuchBucket: The specified bucket does not exist") {
			glog.Errorln("Cannot access S3 Bucket, either it does not exist or the IAM user/role the broker is configured to use has no access to the bucket")
			syscall.Exit(2)
		}
		panic(err)
	}
	if db.DataStorePort.Lock("catalogUpdate") {
		listingUpdate(l, listingcache)
		db.DataStorePort.Unlock("catalogUpdate")
	} else {
		if db.DataStorePort.WaitForUnlock("catalogUpdate") == false {
			if db.DataStorePort.Lock("catalogUpdate") {
				listingUpdate(l, listingcache)
				db.DataStorePort.Unlock("catalogUpdate")
			}
		}
	}
	metadataUpdate(listingcache, catalogcache, bd, s3svc, db)
}

func pollUpdate(interval int, l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc s3.S3, db dbConnector.Db, bl BusinessLogic) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		go updateCatalog(l, c, bd, s3svc, db, bl)
	}
}

// ServiceNeedsUpdate if Update == true the metadata should be refreshed from s3
type ServiceNeedsUpdate struct {
	Name   string
	Update bool
}

func metadataUpdate(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc s3.S3, db dbConnector.Db) {
	data, err := l.Get("__LISTINGS__")
	if err != nil {
		panic(err)
	}
	var lockretry []string
	for _, item := range data.([]ServiceNeedsUpdate) {
		if item.Update {
			if db.DataStorePort.Lock("ServiceSpec-" + item.Name) {
				key := bd.prefix + item.Name + "-spec.yaml"
				obj, err := s3svc.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(bd.bucket),
					Key:    aws.String(key),
				})
				if err != nil {
					glog.V(10).Infoln(bd.prefix + item.Name + "-spec.yaml")
					glog.V(10).Infoln(bd.bucket + "/" + key)
					panic(err)
				} else {
					file, err := ioutil.ReadAll(obj.Body)
					if err != nil {
						panic(err)
					} else {
						var i map[string]interface{}
						yamlerr := yaml.Unmarshal(file, &i)
						if yamlerr != nil {
							panic(yamlerr)
						} else {
							osbdef := db.ServiceDefinitionToOsb(i)
							if osbdef.Name != "" {
								err = db.DataStorePort.PutServiceDefinition(osbdef)
								if err == nil {
									c.Set(item.Name, osbdef)
								} else {
									glog.V(10).Infoln(item)
									glog.V(10).Infoln(osbdef)
									glog.Errorln(err)
								}
							} else {
								glog.Errorf("invalid service definition for %q returned", i["name"].(string))
								glog.Errorln(i)
								glog.Errorln(osbdef)
							}
						}
					}
				}
				db.DataStorePort.Unlock("ServiceSpec-" + item.Name)
			} else {
				lockretry = append(lockretry, "ServiceSpec-"+item.Name)
			}
		} else {
			i, geterr := c.Get(item.Name)
			if geterr != nil {
				glog.Errorln(geterr)
			} else {
				c.Set(item.Name, i)
			}
		}
	}
	failedlock := false
	if len(lockretry) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(lockretry))
		for _, item := range lockretry {
			go func() {
				defer wg.Done()
				if db.DataStorePort.WaitForUnlock(item) == false {
					failedlock = true
				} else {
					serviceuuid := uuid.NewV5(db.Accountuuid, item).String()
					sdef, err := db.DataStorePort.GetServiceDefinition(serviceuuid)
					if err != nil {
						glog.Errorf("failed to get service definition for %q", item)
						glog.Errorln(err)
					} else {
						c.Set(serviceuuid, sdef)
					}
				}
			}()
		}
		wg.Wait()
		if failedlock {
			metadataUpdate(l, c, bd, s3svc, db)
		}
	}
}

// ServiceLastUpdate date when a service exposed by the broker was last updated from s3
type ServiceLastUpdate struct {
	Name string
	Date time.Time
}

func listingUpdate(l *[]ServiceLastUpdate, c cache.Cache) {
	var services []ServiceNeedsUpdate
	for _, item := range *l {
		data, err := c.Get(item.Name)
		if err != nil {
			if err.Error() == "not found" {
				c.Set(item.Name, item.Date)
				services = append(services, ServiceNeedsUpdate{Name: item.Name, Update: true})
			} else {
				panic(err)
			}
		} else {
			if data.(time.Time).Unix() < item.Date.Unix() {
				c.Set(item.Name, item.Date)
				services = append(services, ServiceNeedsUpdate{Name: item.Name, Update: true})
			} else {
				services = append(services, ServiceNeedsUpdate{Name: item.Name, Update: false})
			}
		}
	}
	glog.Infof("Updating listings cache with %q", services)
	c.Set("__LISTINGS__", services)
}

var _ broker.Interface = &BusinessLogic{}

func listTemplates(s3source *BucketDetailsRequest, b *BusinessLogic) (*[]ServiceLastUpdate, error) {
	glog.Infoln("Listing objects bucket: " + s3source.bucket + " region: " + b.s3region + " prefix: " + s3source.prefix)
	ListResponse, err := b.s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s3source.bucket),
		Prefix: aws.String(s3source.prefix),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to list objects, %v\n", err)
		}
		return nil, err
	}
	numberOfRecords := 0
	for _, s3obj := range ListResponse.Contents {
		if strings.HasSuffix(*s3obj.Key, s3source.suffix) {
			numberOfRecords = numberOfRecords + 1
		}
	}
	glog.Infof("Found %x objects\n", numberOfRecords)
	s := make([]ServiceLastUpdate, 0, numberOfRecords)
	for _, s3obj := range ListResponse.Contents {
		if strings.HasSuffix(*s3obj.Key, s3source.suffix) {
			s = append(s, ServiceLastUpdate{
				Name: strings.TrimSuffix(strings.TrimPrefix(*s3obj.Key, s3source.prefix), s3source.suffix),
				Date: *s3obj.LastModified,
			})
		}
	}
	return &s, nil
}

// GetCatalog is executed on a /v2/catalog/ osb api call
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management
func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	response := &broker.CatalogResponse{}

	var services []osb.Service
	l, _ := b.listingcache.Get("__LISTINGS__")
	glog.Infoln(l)
	for _, s := range l.([]ServiceNeedsUpdate) {
		sd, err := b.catalogcache.Get(s.Name)
		if err != nil {
			if err.Error() == "not found" {
				glog.Errorf("Failed to fetch %q from the cache, item not found", s.Name)
			} else {
				glog.Errorln(err)
			}
		} else {
			services = append(services, sd.(osb.Service))
			glog.Infof("ServiceClass: %q %q", sd.(osb.Service).Name, sd.(osb.Service).ID)
			for _, plan := range sd.(osb.Service).Plans {
				glog.Infof("  ServicePlan %q %q", plan.Name, plan.ID)
			}
		}
	}
	osbResponse := &osb.CatalogResponse{Services: services}

	//glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
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

func (b *BusinessLogic) getOverrides(params []string, space string, service string, cluster string) (overrides map[string]string) {
	overrides_env := make(map[string]string)
	for k, v := range b.overrides {
		glog.V(10).Infof("%q: %q", k, v)
		overrides_env[k] = v
	}

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
					paramname := b.brokerid + "_" + c + "_" + n + "_" + s + "_" + p
					v, err := b.db.DataStorePort.GetParam(paramname)
					if err != nil {
						glog.Infof("Unable to fetch parameter override for %#+v", paramname)
						glog.Infoln(err.Error())
					}
					if v != "" {
						overrides[p] = v
					}
					if _, ok := overrides_env[paramname]; ok {
						overrides[p] = overrides_env[paramname]
					}
				}
			}
		}
	}
	glog.Infoln(overrides)
	return overrides
}

func (b *BusinessLogic) getAwsClient(params map[string]string) (cfnsvc *cloudformation.CloudFormation, ssmsvc *ssm.SSM) {
	var defaultCreds credentials.Credentials
	region := aws.String(b.region)
	keyid := b.keyid
	secretkey := b.secretkey
	if _, ok := params["aws_access_key"]; ok {
		keyid = params["aws_access_key"]
	}
	if _, ok := params["aws_secret_key"]; ok {
		secretkey = params["aws_secret_key"]
	}
	glog.V(10).Infoln(params)
	glog.V(10).Infoln(secretkey)
	if keyid != "" {
		glog.V(10).Infof("Using override credentials with keyid %q\n", keyid)
		defaultCreds = *credentials.NewStaticCredentials(keyid, secretkey, "")
	} else {
		defaultCreds = *credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.Must(session.NewSession()))},
			})
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      region,
		Credentials: &defaultCreds,
	}))
	return cloudformation.New(sess), ssm.New(sess)
}

// Provision is executed when the osb api receives PUT /v2/service_instances/:instance_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#provisioning
func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	lockid := "serviceInstance__provision__" + request.InstanceID
	gotlock := b.db.DataStorePort.Lock(lockid)
	if gotlock == false {
		if b.db.DataStorePort.WaitForUnlock(lockid) == false {
			gotlock = b.db.DataStorePort.Lock(lockid)
		}
	}
	if gotlock {
		response := broker.ProvisionResponse{}
		instance := &dbConnector.ServiceInstance{
			ID:        request.InstanceID,
			ServiceID: request.ServiceID,
			PlanID:    request.PlanID,
		}
		if request.AcceptsIncomplete {
			response.Async = true
		}
		servicedef, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
		var plandef osb.Plan
		for _, v := range servicedef.Plans {
			if v.ID == request.PlanID {
				plandef = v
			}
		}
		if err != nil {
			panic(err)
		}
		i, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)

		// Check to see if this is the same instance
		if err != nil {
			panic(err)
		} else if i.ID != "" {
			if i.Match(instance) {
				response.Exists = true
				return &response, nil
			}
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		} else {
			var tags []*cloudformation.Tag
			tags = append(tags, &cloudformation.Tag{
				Key:   aws.String("ServiceBrokerId"),
				Value: aws.String(b.region + "::" + b.brokerid),
			})
			tags = append(tags, &cloudformation.Tag{
				Key:   aws.String("ServiceBrokerInstanceId"),
				Value: aws.String(instance.ID),
			})
			var Cap []*string
			Cap = append(Cap, aws.String("CAPABILITY_IAM"))
			Cap = append(Cap, aws.String("CAPABILITY_NAMED_IAM"))
			//glog.Infoln(plandef.Schemas.ServiceInstance.Create.Parameters)

			params := getParams(plandef.Schemas.ServiceInstance.Create.Parameters)
			glog.V(10).Infoln(params)
			ns := "all"
			cluster := "all"
			if _, ok := request.Context["platform"]; ok {
				if request.Context["platform"].(string) == "cloudfoundry" {
					ns = request.Context["space_guid"].(string)
					ns = strings.Replace(ns, "-", "", -1)
					cluster = request.Context["organization_guid"].(string)
					cluster = strings.Replace(cluster, "-", "", -1)
				} else if request.Context["platform"].(string) == "kubernetes" {
					ns = request.Context["namespace"].(string)
					cluster = request.Context["clusterid"].(string)
				}
			}
			completeparams := b.getOverrides(params, ns, servicedef.Name, cluster)
			glog.V(10).Infoln(completeparams)
			for k, p := range request.Parameters {
				completeparams[k] = p.(string)
			}
			glog.V(10).Infoln(completeparams)
			cfnsvc, _ := b.getAwsClient(completeparams)
			instance.Params = make(map[string]string)
			for k, v := range completeparams {
				instance.Params[k] = v
			}
			glog.V(10).Infoln(instance.Params)
			rolearn := b.rolearn
			if _, ok := completeparams["aws_cloudformation_role_arn"]; ok {
				rolearn = completeparams["aws_cloudformation_role_arn"]
			}
			nonCfnParamarams := []string{"aws_cloudformation_role_arn", "aws_access_key", "aws_secret_key", "SBArtifactS3KeyPrefix", "SBArtifactS3Bucket", "region"}
			for _, k := range nonCfnParamarams {
				if _, ok := completeparams[k]; ok {
					delete(completeparams, k)
				}
			}
			var inputParams []*cloudformation.Parameter
			for k, p := range completeparams {
				param := cloudformation.Parameter{
					ParameterKey:   aws.String(k),
					ParameterValue: aws.String(p),
				}
				glog.V(10).Infoln(param)
				glog.V(10).Infof("%q: %q\n", k, p)
				inputParams = append(inputParams, &param)
			}
			glog.V(10).Infoln(inputParams)
			stackInput := cloudformation.CreateStackInput{
				Capabilities: Cap,
				Parameters:   inputParams,
				RoleARN:      aws.String(rolearn),
				StackName:    aws.String("CfnServiceBroker-" + servicedef.Name + "-" + instance.ID),
				Tags:         tags,
				TemplateURL:  aws.String("https://s3.amazonaws.com/" + b.s3bucket + "/" + b.s3key + strings.TrimSuffix(servicedef.Name, "-apb") + b.templatefilter),
			}
			glog.V(10).Infoln(stackInput)
			results, err := cfnsvc.CreateStack(&stackInput)
			if err != nil {
				glog.Errorln(err)
				b.db.DataStorePort.Unlock(lockid)
				return &response, err
			}
			instance.StackID = *results.StackId
			err = b.db.DataStorePort.PutServiceInstance(*instance)
			if err != nil {
				glog.Errorln(err)
				b.db.DataStorePort.Unlock(lockid)
				return &response, err
			}
		}
		b.db.DataStorePort.Unlock(lockid)
		return &response, nil
	}
	description := "Failed to get lock for instanceId" + string(request.InstanceID)
	return nil, osb.HTTPStatusCodeError{
		StatusCode:  http.StatusExpectationFailed,
		Description: &description,
	}
}

// Deprovision executed when the osb api receives DELETE /v2/service_instances/:instance_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#deprovisioning
func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	lockid := "serviceInstance__deprovision__" + request.InstanceID
	gotlock := b.db.DataStorePort.Lock(lockid)
	response := broker.DeprovisionResponse{}
	if gotlock == false {
		if b.db.DataStorePort.WaitForUnlock(lockid) == false {
			gotlock = b.db.DataStorePort.Lock(lockid)
		}
	}
	if gotlock {
		si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
		if err != nil {
			panic(err)
		}
		if si.StackID == "" {
			errmsg := "CloudFormation stackid missing, chances are stack creation failed in an unexpected way, assuming there is nothing to deprovision"
			glog.Errorln(errmsg)
			response.Async = false
			return &response, nil
		}
		glog.V(10).Infoln(si.Params)
		cfnsvc, _ := b.getAwsClient(si.Params)
		_, err = cfnsvc.DeleteStack(&cloudformation.DeleteStackInput{StackName: aws.String(si.StackID)})

		if err != nil {
			panic(err)
		}

		b.db.DataStorePort.Unlock(lockid)

		if err != nil {
			panic(err)
		}
		if request.AcceptsIncomplete {
			response.Async = true
		}
		return &response, nil
	}
	description := "Failed to get lock for instanceId" + string(request.InstanceID)
	return nil, osb.HTTPStatusCodeError{
		StatusCode:  http.StatusExpectationFailed,
		Description: &description,
	}
}

// LastOperation executed when the osb api receives GET /v2/service_instances/:instance_id/last_operation
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#polling-last-operation
func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.Infoln(request)
	glog.Infoln(c)
	si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		panic(err)
	}
	glog.Infoln(si)
	r := broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "", Description: nil}}
	if si.StackID == "" {
		errmsg := "CloudFormation stackid missing, chances are stack creation failed in an unexpected way"
		glog.Errorln(errmsg)
		r.LastOperationResponse.State = "failed"
		r.LastOperationResponse.Description = &errmsg
		return &r, nil
	}
	glog.V(10).Infoln(si.Params)
	cfnsvc, _ := b.getAwsClient(si.Params)
	response, err := cfnsvc.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(si.StackID)})
	if err != nil {
		panic(err)
	}
	failedstates := []string{"CREATE_FAILED", "ROLLBACK_IN_PROGRESS", "ROLLBACK_FAILED", "ROLLBACK_COMPLETE", "DELETE_FAILED", "UPDATE_ROLLBACK_IN_PROGRESS", "UPDATE_ROLLBACK_FAILED", "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"}
	progressingstates := []string{"CREATE_IN_PROGRESS", "DELETE_IN_PROGRESS", "UPDATE_IN_PROGRESS", "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"}
	successfulstates := []string{"CREATE_COMPLETE", "DELETE_COMPLETE", "UPDATE_COMPLETE"}
	status := *response.Stacks[0].StackStatus
	if stringInSlice(status, failedstates) {
		glog.Errorf("CloudFormation stack failed %#+v", si.StackID)
		glog.Errorf(status)
		r.LastOperationResponse.State = "failed"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else if stringInSlice(status, progressingstates) {
		glog.Infoln("CloudFormation stack still busy...")
		glog.Infoln(status)
		r.LastOperationResponse.State = "in progress"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else if stringInSlice(status, successfulstates) {
		glog.Infoln("CloudFormation stack operation completed...")
		glog.Infoln(status)
		r.LastOperationResponse.State = "succeeded"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else {
		return nil, fmt.Errorf("unexpected cfn status %v", status)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Bind executed when the osb api receives PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#request-4
func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {

	si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	service, err := b.db.DataStorePort.GetServiceDefinition(si.ServiceID)
	if err != nil {
		panic(err)
	}
	glog.Infoln(si)
	cfnsvc, ssmsvc := b.getAwsClient(si.Params)
	cfnresponse, err := cfnsvc.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(si.StackID)})
	if err != nil {
		panic(err)
	}
	outputs := make(map[string]interface{})
	for _, o := range cfnresponse.Stacks[0].Outputs {
		fmt.Println(o)
		if *o.OutputKey == "UserKeyId" || *o.OutputKey == "UserSecretKey" {
			ssmInput := ssm.GetParameterInput{
				Name:           aws.String(*o.OutputValue),
				WithDecryption: aws.Bool(true),
			}

			ssmresponse, err := ssmsvc.GetParameter(&ssmInput)
			if err != nil {
				panic(err)
			}
			pname := strings.ToUpper(service.Name) + "_" + toSnakeCase(*o.OutputKey)
			outputs[pname] = ssmresponse.Parameter.Value
		} else {
			outputs[toSnakeCase(*o.OutputKey)] = o.OutputValue
		}
	}
	glog.Infoln(outputs)
	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: outputs,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = false
	}
	return &response, nil
}

func (b *BusinessLogic) GetBinding(request *osb.GetBindingRequest, c *broker.RequestContext) (*broker.GetBindingResponse, error) {
	glog.V(10).Infoln(request)
	glog.V(10).Infoln(c)
	return &broker.GetBindingResponse{}, nil
}

func BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.V(10).Infoln(request)
	glog.V(10).Infoln(c)
	return &broker.LastOperationResponse{}, nil
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

// Unbind executed when the osb api receives DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#unbinding
func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

// Update is not supported at present, so is just a skeleton
func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = true
	}
	return &response, nil
}

func (b *BusinessLogic) BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	return &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "", Description: nil}}, nil
}

// ValidateBrokerAPIVersion does nothing ?
func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	glog.Infof("Client OSB API Version: %q", version)
	return nil
}
