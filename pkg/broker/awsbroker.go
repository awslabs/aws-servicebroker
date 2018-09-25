package broker

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

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
	yaml "gopkg.in/yaml.v2"
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
	err = metadataUpdate(listingcache, catalogcache, bd, s3svc, db, MetadataUpdate)
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

func MetadataUpdate(l cache.Cache, c cache.Cache, bd BucketDetailsRequest, s3svc S3Client, db Db, metadataUpdate MetadataUpdater) error {
	data, err := l.Get("__LISTINGS__")
	if err != nil {
		return err
	}
	for _, item := range data.([]ServiceNeedsUpdate) {
		if item.Update {
			key := bd.prefix + item.Name + "-spec.yaml"
			obj, err := s3svc.Client.GetObject(&s3.GetObjectInput{
				Bucket: aws.String(bd.bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				return err
			} else if obj.Body == nil {
				return errors.New("s3 object body missing")
			} else {
				file, err := ioutil.ReadAll(obj.Body)
				if err != nil {
					return err
				} else {
					var i map[string]interface{}
					yamlerr := yaml.Unmarshal(file, &i)
					if yamlerr != nil {
						return yamlerr
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
		} else {
			i, geterr := c.Get(item.Name)
			if geterr != nil {
				glog.Errorln(geterr)
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
func (db Db) ServiceDefinitionToOsb(sd map[string]interface{}) osb.Service {
	// TODO: Marshal spec straight from the yaml in an osb.Plan, possibly using gjson
	glog.Infof("converting service definition %q ", sd["name"].(string))
	defer func() {
		if r := recover(); r != nil {
			glog.Errorln(errors.Wrap(r, 2).ErrorStack())
			glog.Errorf("Failed to convert service definition for %q", sd["name"].(string))
		}
	}()
	f := false
	serviceid := uuid.NewV5(db.Accountuuid, sd["name"].(string)).String()
	outp := osb.Service{}
	outp.ID = serviceid
	outp.Name = sd["name"].(string)
	outp.Bindable = sd["bindable"].(bool)
	outp.Description = sd["description"].(string)
	outp.PlanUpdatable = &f
	metadata := make(map[string]interface{})
	for index, key := range sd["metadata"].(map[interface{}]interface{}) {
		metadata[index.(string)] = key
	}
	outp.Metadata = metadata
	var tags []string
	for _, key := range sd["tags"].([]interface{}) {
		tags = append(tags, key.(string))
	}
	outp.Tags = tags
	var plans []osb.Plan
	for _, key := range sd["plans"].([]interface{}) {
		plan := osb.Plan{}
		for i, k := range key.(map[interface{}]interface{}) {
			if i.(string) == "name" {
				plan.Name = k.(string)
			} else if i.(string) == "description" {
				plan.Description = k.(string)
			} else if i.(string) == "free" {
				free := k.(bool)
				plan.Free = &free
			} else if i.(string) == "metadata" {
				metadata := make(map[string]interface{})
				for i2, k2 := range k.(map[interface{}]interface{}) {
					metadata[i2.(string)] = k2
				}
				plan.Metadata = metadata
			} else if i.(string) == "parameters" {
				propsForCreate := make(map[string]interface{})
				requiredForCreate := make([]string, 0)
				propsForUpdate := make(map[string]interface{})
				requiredForUpdate := make([]string, 0)
				for _, param := range k.([]interface{}) {
					var name string
					var required, updatable bool
					pvals := make(map[string]interface{})
					for pk, pv := range param.(map[interface{}]interface{}) {
						switch pk {
						case "name":
							name = pv.(string)
						case "required":
							required = pv.(bool)
						case "type":
							switch pv {
							case "enum":
								pvals[pk.(string)] = "string"
							case "int":
								pvals[pk.(string)] = "integer"
							default:
								pvals[pk.(string)] = pv
							}
						case "updatable":
							updatable = pv.(bool)
						default:
							pvals[pk.(string)] = pv
						}
					}
					propsForCreate[name] = pvals
					if required {
						requiredForCreate = append(requiredForCreate, name)
					}
					if updatable {
						propsForUpdate[name] = pvals
						if required {
							requiredForUpdate = append(requiredForUpdate, name)
						}
					}
				}
				plan.Schemas = &osb.Schemas{
					ServiceInstance: &osb.ServiceInstanceSchema{
						Create: &osb.InputParametersSchema{
							Parameters: map[string]interface{}{
								"type":       "object",
								"properties": propsForCreate,
								"$schema":    "http://json-schema.org/draft-06/schema#",
							},
						},
					},
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
						},
					}
					if len(requiredForUpdate) > 0 {
						// Cloud Foundry does not allow "required" to be an empty slice
						plan.Schemas.ServiceInstance.Update.Parameters.(map[string]interface{})["required"] = requiredForUpdate
					}
				}
			}
		}
		planid := uuid.NewV5(db.Accountuuid, "service__"+sd["name"].(string)+"__plan__"+plan.Name).String()
		plan.ID = planid
		plans = append(plans, plan)
	}
	outp.Plans = plans
	glog.Infof("done converting service definition %q ", sd["name"].(string))
	return outp
}

func (b *AwsBroker) generateS3HTTPUrl(serviceDefName string) *string {
	prefix := "https://s3.amazonaws.com/"
	if b.s3region != "us-east-1" {
		prefix = fmt.Sprintf("https://s3-%s.amazonaws.com/", b.s3region)
	}
	return aws.String(prefix + b.s3bucket + "/" + b.s3key + strings.TrimSuffix(serviceDefName, "-apb") + b.templatefilter)
}
