package broker

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/satori/go.uuid"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"time"
	"math/rand"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"reflect"
	"github.com/golang/glog"
	"syscall"
	"strings"
	"github.com/go-errors/errors"
)

// Db dynamodb configuration
type Db struct {
	region 		string
	accountid 	string
	accountuuid uuid.UUID
	brokerid 	string
	tablename 	string
	ddb 		dynamodb.DynamoDB
}

// ServiceItem used to unmarshal catalog entries from DynamoDb
type ServiceItem struct {
	ID          string      `json:"id"`
	Userid      string      `json:"userid"`
	Service     osb.Service `json:"service"`
	Serviceid   string      `json:"serviceid"`
	Servicename string      `json:"servicename"`
}

// Param stores a parameter value
type Param struct {
	Value        string `json:"value"`
}

// Lock attempts to gain a distributed lock using DynamoDb as the backend
func (db Db) Lock (lockname string) bool {
	lockuuid := uuid.NewV5(db.accountuuid, lockname)
	condBuilder := expression.AttributeNotExists(expression.Name("id"))
	cond, err := expression.NewBuilder().WithCondition(condBuilder).Build()
	if err != nil {
		panic(err)
	}
	putInput := dynamodb.PutItemInput{
		ConditionExpression: cond.Condition(),
		ExpressionAttributeNames: cond.Names(),
		ExpressionAttributeValues: cond.Values(),
		TableName: aws.String(db.tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id": { S: aws.String(lockuuid.String()) },
			"userid": { S: aws.String(db.accountuuid.String()) },
			"type": { S: aws.String("lock") },
		},
	}
	_, err = db.ddb.PutItem(&putInput)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ResourceNotFoundException: Requested resource not found") {
			glog.Errorln("Cannot continue, DynamoDB table " + db.tablename + " does not exist in region " + db.region + " accountid " + db.accountid )
			syscall.Exit(2)
		}
		glog.Errorln("\"" + err.Error() + "\"")
		glog.Errorln("already locked...")
		return false
	}
	return true
}

// IsLocked tells whether a given lock is currently locked
func (db Db) IsLocked (lockname string) bool {
	consistentRead := true
	lockuuid := uuid.NewV5(db.accountuuid, lockname)
	getInput := dynamodb.GetItemInput{
		ConsistentRead: &consistentRead,
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(lockuuid.String()),
			},
			"userid": {
				S: aws.String(db.accountuuid.String()),
			},
		},
	}
	result, err := db.ddb.GetItem(&getInput)
	if err != nil {
		panic(err)
	}
	if len(result.Item) == 0 {
		return false
	}
	return true
}

// WaitForUnlock blocks until given lock is unlocked or timeout is reached
func (db Db) WaitForUnlock (lockname string) bool {
	waited := 0
	timeout := 10
	for db.IsLocked(lockname) {
		rand.Seed(time.Now().UTC().UnixNano())
		delay := rand.Intn(15-5) + 5
		waited += delay
		if waited > timeout {
			db.Unlock(lockname)
			return false
		}
		time.Sleep(time.Second * time.Duration(delay))
	}
	return true
}

// Unlock release given lock
func (db Db) Unlock (lockname string) error {
	lockuuid := uuid.NewV5(db.accountuuid, lockname)
	DeleteInput := dynamodb.DeleteItemInput{
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id": { S: aws.String(lockuuid.String()) },
			"userid": { S: aws.String(db.accountuuid.String()) },
		},
	}
	_, err := db.ddb.DeleteItem(&DeleteInput)
	if err != nil {
		return err
	}
	return nil
}

// PutServiceDefinition push catalog service definition to DynamoDb
func (db Db) PutServiceDefinition(sd osb.Service) error {
	glog.Infof("putting service definition %q into dynamdb", sd.Name)
	serviceid := uuid.NewV5(db.accountuuid, sd.Name)
	si, err := dynamodbattribute.Marshal(sd)
	if err != nil {
		glog.Errorln(err)
		return err
	}
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(serviceid.String())},
			"userid": {S: aws.String(db.accountuuid.String())},
			"serviceid": {S: aws.String(serviceid.String())},
			"servicename": {S: aws.String(sd.Name)},
			"service": si,
			"type": {S: aws.String("service")},
		},
	}
	_, err = db.ddb.PutItem(&putInput)
	if err != nil {
		glog.Infoln(putInput)
		glog.Errorln(err)
		return err
	}
	glog.Infof("done putting service definition %q into dynamdb", sd.Name)
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
	serviceid := uuid.NewV5(db.accountuuid, sd["name"].(string)).String()
	outp := osb.Service{}
	outp.ID = serviceid
	outp.Name = sd["name"].(string)
	outp.Bindable = sd["bindable"].(bool)
	outp.Description = sd["description"].(string)
	outp.PlanUpdatable = &f
	metadata := make(map[string]interface{})
	for index, key := range sd["metadata"].(map[interface{}]interface {}) {
		metadata[index.(string)] = key
	}
	outp.Metadata = metadata
	var tags []string
	for _, key := range sd["tags"].([]interface{}){
		tags = append(tags, key.(string))
	}
	outp.Tags = tags
	var plans []osb.Plan
	for _, key := range sd["plans"].([]interface{}) {
		plan := osb.Plan{}
		for i, k := range key.(map[interface {}]interface {}){
			if i.(string) == "name" {
				plan.Name = k.(string)
			} else if i.(string) == "description" {
				plan.Description = k.(string)
			} else if i.(string) == "free" {
				free := k.(bool)
				plan.Free = &free
			} else if i.(string) == "metadata" {
				metadata := make(map[string]interface{})
				for i2, k2 := range k.(map[interface{}]interface {}) {
					metadata[i2.(string)] = k2
				}
				plan.Metadata = metadata
			} else if i.(string) == "parameters" {
				props := make(map[string]interface{})
				required := make([]string, 0)
				for _, param := range k.([]interface {}) {
					name := ""
					isRequired := false
					pvals := make(map[string]interface{})
					for pk, pv := range param.(map[interface {}]interface {}){
						if pk == "name" {
							name = pv.(string)
						}else if pk == "enum" {
							var enum []string
							for _, e := range pv.([]interface{}) {
								enum = append(enum, e.(string))
							}
							pvals[pk.(string)] = enum
						} else if pk == "required" {
							isRequired = pv.(bool)
						} else if pk == "type" && pv == "enum" {
							pvals[pk.(string)] = "string"
						} else if pk == "type" && pv == "int" {
							pvals[pk.(string)] = "integer"
						} else {
							pvals[pk.(string)] = pv
						}
					}
					if isRequired {
						required = append(required, name)
					}
					props[name] = pvals
				}
				plan.Schemas = &osb.Schemas{
					ServiceInstance: &osb.ServiceInstanceSchema{
						Create: &osb.InputParametersSchema{
							Parameters: map[string]interface{}{
								"type": "object",
								"properties": props,
								"$schema": "http://json-schema.org/draft-06/schema#",
								"required": required,
							},
						},
					},
				}
			}
		}
		planid := uuid.NewV5(db.accountuuid, "service__" + sd["name"].(string) + "__plan__" + plan.Name).String()
		plan.ID = planid
		plans = append(plans, plan)
	}
	outp.Plans = plans
	glog.Infof("done converting service definition %q ", sd["name"].(string))
	return outp
}

// ServiceInstance deatils of a service instance
type ServiceInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]string
	StackID   string
}

func (i *ServiceInstance) match(other *ServiceInstance) bool {
	return reflect.DeepEqual(i, other)
}

// GetParam fetches parameter from Dynamo
func (db Db) GetParam(paramname string) (value string, err error) {
	paramuuid := uuid.NewV5(db.accountuuid, paramname).String()
	getInput := dynamodb.GetItemInput{
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(paramuuid)},
			"userid": {S: aws.String(db.accountuuid.String())},
		},
	}
	result, err := db.ddb.GetItem(&getInput)
	if err != nil {
		return "", err
	}
	if len(result.Item) == 0 {
		return "", fmt.Errorf("parameter does not exist")
	}

	item := Param{}
	glog.Infoln("unmarshalling item")
	glog.Infoln(result.Item)
	dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return "", err
	}
	if item.Value == "" {
		return "", fmt.Errorf("could not unmarshal service definition")
	}
	return item.Value, nil
}

// PutParam puts parameters into Dynamo
func (db Db) PutParam(paramname string, paramvalue string) error {
	paramuuid := uuid.NewV5(db.accountuuid, paramname).String()
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(paramuuid)},
			"userid": {S: aws.String(db.accountuuid.String())},
			"value": {S: aws.String(paramvalue)},
			"type": {S: aws.String("parameter")},
		},
	}
	_, err := db.ddb.PutItem(&putInput)
	if err != nil {
		return err
	}
	return nil
}

// GetServiceDefinition fetches given catalog service definition from Dynamo
func (db Db) GetServiceDefinition(serviceuuid string) (osb.Service, error) {
	servicedef := osb.Service{}
	getInput := dynamodb.GetItemInput{
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(serviceuuid)},
			"userid": {S: aws.String(db.accountuuid.String())},
		},
	}
	result, err := db.ddb.GetItem(&getInput)
	if err != nil {
		return servicedef, err
	}
	if len(result.Item) == 0 {
		return servicedef, fmt.Errorf("Service Definition does not exist")
	}

	item := ServiceItem{}
	//glog.Infoln("Debug: unmarshalling item")
	//glog.Infoln(result.Item)
	dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return servicedef, err
	}
	if item.Service.ID == "" {
		return servicedef, fmt.Errorf("could not unmarshal service definition")
	}
	return item.Service, nil
}

// GetServiceInstance fetches given service instance from Dynamo
func (db Db) GetServiceInstance(sid string) (ServiceInstance, error) {
	var si ServiceInstance
	input := dynamodb.GetItemInput{
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id": { S: aws.String(sid) },
			"userid": { S: aws.String(db.accountuuid.String()) },
		},
	}
	outp, err := db.ddb.GetItem(&input)
	if err != nil {
		panic(err)
		return si, err
	}
	dynamodbattribute.Unmarshal(outp.Item["serviceinstance"], &si)
	if err != nil {
		panic(err)
		return si, err
	}
	return si, nil
}

// PutServiceInstance stores given service instance in Dynamo
func (db Db) PutServiceInstance(si ServiceInstance) error {
	msi, err := dynamodbattribute.Marshal(si)
	if err != nil {
		return err
	}
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(si.ID)},
			"userid": {S: aws.String(db.accountuuid.String())},
			"serviceinstance": msi,
			"type": {S: aws.String("serviceinstance")},
		},
	}
	_, err = db.ddb.PutItem(&putInput)
	if err != nil {
		return err
	}
	return nil
}