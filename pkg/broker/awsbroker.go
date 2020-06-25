package broker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/awslabs/aws-servicebroker/pkg/dynamodbadapter"
	"github.com/go-errors/errors"
	"github.com/golang/glog"
	"github.com/koding/cache"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	uuid "github.com/satori/go.uuid"
)

// Runs at startup and bootstraps the broker
func NewAWSBroker(o Options, awssess GetAwsSession, clients AwsClients, getCallerId GetCallerIder, updateCatalog UpdateCataloger, pollUpdate PollUpdater) (*AwsBroker, error) {

	sess := awssess(o.KeyID, o.SecretKey, o.Region, "", o.Profile, map[string]string{})
	s3sess := awssess(o.KeyID, o.SecretKey, o.S3Region, "", o.Profile, map[string]string{})
	s3svc := clients.NewS3(s3sess)
	ddbsvc := clients.NewDdb(sess)
	stssvc := clients.NewSts(sess)
	callerid, err := getCallerId(stssvc)
	if err != nil {
		return &AwsBroker{}, err
	}
	accountid := *callerid.Account
	accountuuid := uuid.NewV5(uuid.NullUUID{}.UUID, accountid+o.BrokerID)

	glog.Infof("Running as caller identity '%+v'.", callerid)

	var db Db
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

	// setup in memory cache
	var catalogcache = cache.NewMemoryWithTTL(time.Duration(CacheTTL))
	var listingcache = cache.NewMemoryWithTTL(time.Duration(CacheTTL))
	listingcache.StartGC(time.Minute * 5)
	bd := &BucketDetailsRequest{
		o.S3Bucket,
		o.S3Key,
		o.TemplateFilter,
	}

	// retrieve AWS partition from instance metadata service
	partition, err := ec2metadata.New(sess).GetMetadata("/services/partition")

	if err != nil {
		partition = "aws" // no access to metadata service, defaults to AWS Standard Partition
	}

	// populate broker variables
	bl := AwsBroker{
		accountId:          accountid,
		keyid:              o.KeyID,
		secretkey:          o.SecretKey,
		profile:            o.Profile,
		tablename:          o.TableName,
		s3bucket:           o.S3Bucket,
		s3region:           o.S3Region,
		s3key:              addTrailingSlash(o.S3Key),
		templatefilter:     o.TemplateFilter,
		region:             o.Region,
		partition:          partition,
		s3svc:              s3svc,
		catalogcache:       catalogcache,
		listingcache:       listingcache,
		brokerid:           o.BrokerID,
		db:                 db,
		GetSession:         awssess,
		Clients:            clients,
		prescribeOverrides: o.PrescribeOverrides,
		globalOverrides:    getGlobalOverrides(o.BrokerID),
	}

	// get catalog and setup periodic updates from S3
	err = updateCatalog(listingcache, catalogcache, *bd, s3svc, db, bl, ListTemplates, ListingUpdate, MetadataUpdate)
	if err != nil {
		return &AwsBroker{}, err
	}
	go pollUpdate(600, listingcache, catalogcache, *bd, s3svc, db, bl, updateCatalog, ListTemplates)
	return &bl, nil
}

func UpdateCatalog(listingcache cache.Cache, catalogcache cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, listTemplates ListTemplateser, listingUpdate ListingUpdater, metadataUpdate MetadataUpdater) error {
	l, err := listTemplates(&bd, &bl)
	if err != nil {
		if strings.HasPrefix(err.Error(), "NoSuchBucket: The specified bucket does not exist") {
			return errors.New("Cannot access S3 Bucket, either it does not exist or the IAM user/role the broker is configured to use has no access to the bucket")
		}
		return err
	}
	err = listingUpdate(l, listingcache)
	if err != nil {
		return err
	}
	err = metadataUpdate(listingcache, catalogcache, bd, s3svc, db, MetadataUpdate, bl.templatefilter)
	if err != nil {
		return err
	}
	return nil
}

func PollUpdate(interval int, l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, bl AwsBroker, updateCatalog UpdateCataloger, listTemplates ListTemplateser) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		go updateCatalog(l, c, bd, s3svc, db, bl, listTemplates, ListingUpdate, MetadataUpdate)
	}
}

func MetadataUpdate(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater, templatefilter string) error {
	data, err := l.Get("__LISTINGS__")
	if err != nil {
		return err
	}
	for _, item := range data.([]ServiceNeedsUpdate) {
		if item.Update {
			file, err := getObjectBody(s3svc, bd.bucket, bd.prefix+item.Name+templatefilter)
			if err != nil {
				glog.Errorln(err)
				continue
			}
			if err := templateToServiceDefinition(file, db, c, item); err != nil {
				glog.Errorln(err)
			}
		} else {
			i, err := c.Get(item.Name)
			if err != nil {
				glog.Errorln(err)
			} else {
				c.Set(item.Name, i)
			}
		}
	}
	return nil
}

func ListingUpdate(l *[]ServiceLastUpdate, c cache.Cache) error {
	var services []ServiceNeedsUpdate
	for _, item := range *l {
		data, err := c.Get(item.Name)
		if err != nil {
			if err.Error() == "not found" {
				c.Set(item.Name, item.Date)
				services = append(services, ServiceNeedsUpdate{Name: item.Name, Update: true})
			} else {
				return err
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
	glog.Infof("Updating listings cache with %v", services)
	c.Set("__LISTINGS__", services)
	return nil
}

func ListTemplates(s3source *BucketDetailsRequest, b *AwsBroker) (*[]ServiceLastUpdate, error) {
	glog.Infoln("Listing objects bucket: " + s3source.bucket + " region: " + b.s3region + " prefix: " + s3source.prefix)
	ListResponse, err := b.s3svc.Client.ListObjectsV2(&s3.ListObjectsV2Input{
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

// ValidateBrokerAPIVersion still to determine supported api versions
func (b *AwsBroker) ValidateBrokerAPIVersion(version string) error {
	glog.Infof("Client OSB API Version: %q", version)
	return nil
}

// ServiceDefinitionToOsb converts apb service definition into osb.Service struct
func (db Db) ServiceDefinitionToOsb(sd CfnTemplate) osb.Service {
	// TODO: Marshal spec straight from the yaml in an osb.Plan, possibly using gjson
	glog.Infof("converting service definition %q ", sd.Metadata.Spec.Name)
	defer func() {
		if r := recover(); r != nil {
			glog.Errorln(errors.Wrap(r, 2).ErrorStack())
			glog.Errorf("Failed to convert service definition for %q", sd.Metadata.Spec.Name)
		}
	}()
	serviceid := uuid.NewV5(db.Accountuuid, sd.Metadata.Spec.Name).String()
	outp := osb.Service{
		ID:          serviceid,
		Name:        sd.Metadata.Spec.Name,
		Description: stripTemplateID(sd.Description),
		Tags:        sd.Metadata.Spec.Tags,
		Bindable:    true,
		Metadata: map[string]interface{}{
			"displayName":         sd.Metadata.Spec.DisplayName,
			"providerDisplayName": sd.Metadata.Spec.ProviderDisplayName,
			"documentationUrl":    sd.Metadata.Spec.DocumentationUrl,
			"imageUrl":            sd.Metadata.Spec.ImageUrl,
			"longDescription":     sd.Metadata.Spec.LongDescription,
			"outputsAsIs":         sd.Metadata.Spec.OutputsAsIs,
			"cloudFoundry":        sd.Metadata.Spec.CloudFoundry,
			"bindViaLambda":       sd.Metadata.Spec.BindViaLambda,
		},
		PlanUpdatable: aws.Bool(false),
	}

	var plans []osb.Plan
	params := cfnParamsToOsb(sd)
	for k, p := range sd.Metadata.Spec.ServicePlans {
		planid := uuid.NewV5(db.Accountuuid, "service__"+sd.Metadata.Spec.Name+"__plan__"+k).String()
		plan := osb.Plan{
			ID:          planid,
			Name:        k,
			Description: p.Description,
			Free:        aws.Bool(false),
			Bindable:    aws.Bool(true),
			Metadata: map[string]interface{}{
				"cost":            p.Cost,
				"displayName":     p.DisplayName,
				"longDescription": p.LongDescription,
			},
			Schemas: &osb.Schemas{ServiceInstance: &osb.ServiceInstanceSchema{}},
		}
		propsForCreate := make(map[string]interface{})
		var openshiftFormCreate []OpenshiftFormDefinition
		for nk, nv := range nonCfnParamDefs {
			openshiftFormCreate = openshiftFormAppend(openshiftFormCreate, nk, nv.(map[string]interface{}))
			nonCfnParam := make(map[string]interface{})
			for nnk, nnv := range nv.(map[string]interface{}) {
				if nnk != "display_group" {
					nonCfnParam[nnk] = nnv
				}
			}
			propsForCreate[nk] = nonCfnParam
		}
		propsForUpdate := make(map[string]interface{})
		requiredForCreate := make([]string, 0)
		requiredForUpdate := make([]string, 0)
		prescribed := make(map[string]string)
		var openshiftFormUpdate []OpenshiftFormDefinition
		for paramName, paramValue := range params {
			include := true
			for planParam, planValue := range p.ParameterValues {
				if planParam == paramName {
					include = false
					prescribed[planParam] = planValue
				}
			}
			required := false
			if paramValue.(map[string]interface{})["required"] != nil {
				required = paramValue.(map[string]interface{})["required"].(bool)
			}
			if include {
				openshiftFormCreate = openshiftFormAppend(openshiftFormCreate, paramName, paramValue.(map[string]interface{}))
				createParam := map[string]interface{}{}
				for nk, nv := range paramValue.(map[string]interface{}) {
					createParam[nk] = nv
				}
				for planDefaultParam, planDefaultValue := range p.ParameterDefaults {
					if planDefaultParam == paramName {
						glog.V(10).Infof("Updating default with plan default for plan %q param %q\n", k, paramName)
						createParam["default"] = planDefaultValue
					}
				}
				for _, v := range []string{"required", "display_group"} {
					delete(createParam, v)
				}
				propsForCreate[paramName] = createParam
				if required {
					requiredForCreate = append(requiredForCreate, paramName)
				}
				if stringInSlice(paramName, sd.Metadata.Spec.UpdatableParameters) {
					openshiftFormUpdate = openshiftFormAppend(openshiftFormUpdate, paramName, paramValue.(map[string]interface{}))
					updateParam := make(map[string]interface{})
					for nnk, nnv := range paramValue.(map[string]interface{}) {
						if nnk != "required" && nnk != "display_group" && nnk != "default" {
							updateParam[nnk] = nnv
						}
					}
					propsForUpdate[paramName] = updateParam
					if required {
						requiredForUpdate = append(requiredForUpdate, paramName)
					}
				}
			}
		}
		plan.Schemas.ServiceInstance.Create = &osb.InputParametersSchema{
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": propsForCreate,
				"$schema":    "http://json-schema.org/draft-06/schema#",
				"prescribed": prescribed,
			},
		}
		if len(openshiftFormCreate) > 0 {
			plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})["openshift_form_definition"] = openshiftFormCreate
		}
		if len(requiredForCreate) > 0 {
			// Cloud Foundry does not allow "required" to be an empty slice
			plan.Schemas.ServiceInstance.Create.Parameters.(map[string]interface{})["required"] = requiredForCreate
		}
		if len(propsForUpdate) > 0 {
			plan.Schemas.ServiceInstance.Update = &osb.InputParametersSchema{
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": propsForUpdate,
					"$schema":    "http://json-schema.org/draft-06/schema#",
					"prescribed": prescribed,
				},
			}
			if len(openshiftFormUpdate) > 0 {
				plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})["openshift_form_definition"] = openshiftFormCreate
			}
			if len(requiredForUpdate) > 0 {
				// Cloud Foundry does not allow "required" to be an empty slice
				plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})["required"] = requiredForUpdate
			}
		}
		plans = append(plans, plan)
	}
	outp.Plans = plans
	glog.Infof("done converting service definition %q ", sd.Metadata.Spec.Name)
	return outp
}

func (b *AwsBroker) generateS3HTTPUrl(serviceDefName string) *string {
	if b.partition == "aws-cn" {
		// AWS China Partition
		objKey := b.s3key + strings.TrimSuffix(serviceDefName, "-apb") + b.templatefilter
		objURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com.cn/%s", b.s3bucket, b.s3region, objKey)
		return aws.String(objURL)
	} else {
		// AWS Standard Partition and GovCloud Partition
		objKey := b.s3key + strings.TrimSuffix(serviceDefName, "-apb") + b.templatefilter
		objURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", b.s3bucket, objKey)
		if b.s3region != "us-east-1" {
			objURL = fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", b.s3bucket, b.s3region, objKey)
		}
		return aws.String(objURL)
	}
}
