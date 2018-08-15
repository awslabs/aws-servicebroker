package dynamodbadapter

import (
	"fmt"
	dbConnector "github.com/awslabs/aws-service-broker/pkg/db"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"syscall"
	"time"
)

// DynamoDB implementation of DataStore Adapter
type DdbDataStore struct {
	Accountid   string
	Accountuuid uuid.UUID
	Brokerid    string
	Region      string
	Ddb         dynamodb.DynamoDB
	Tablename   string
}

// Lock attempts to gain a distributed lock using DynamoDb as the backend
func (db DdbDataStore) Lock(lockname string) bool {
	lockuuid := uuid.NewV5(db.Accountuuid, lockname)
	condBuilder := expression.AttributeNotExists(expression.Name("id"))
	cond, err := expression.NewBuilder().WithCondition(condBuilder).Build()
	if err != nil {
		panic(err)
	}
	putInput := dynamodb.PutItemInput{
		ConditionExpression:       cond.Condition(),
		ExpressionAttributeNames:  cond.Names(),
		ExpressionAttributeValues: cond.Values(),
		TableName:                 aws.String(db.Tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(lockuuid.String())},
			"userid": {S: aws.String(db.Accountuuid.String())},
			"type":   {S: aws.String("lock")},
		},
	}
	_, err = db.Ddb.PutItem(&putInput)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ResourceNotFoundException: Requested resource not found") {
			glog.Errorln("Cannot continue, DynamoDB table " + db.Tablename + " does not exist in Region " + db.Region + " Accountid " + db.Accountid)
			syscall.Exit(2)
		}
		glog.Errorln("\"" + err.Error() + "\"")
		glog.Errorln("already locked...")
		return false
	}
	return true
}

// IsLocked tells whether a given lock is currently locked
func (db DdbDataStore) IsLocked(lockname string) bool {
	consistentRead := true
	lockuuid := uuid.NewV5(db.Accountuuid, lockname)
	getInput := dynamodb.GetItemInput{
		ConsistentRead: &consistentRead,
		TableName:      aws.String(db.Tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(lockuuid.String()),
			},
			"userid": {
				S: aws.String(db.Accountuuid.String()),
			},
		},
	}
	result, err := db.Ddb.GetItem(&getInput)
	if err != nil {
		panic(err)
	}
	if len(result.Item) == 0 {
		return false
	}
	return true
}

// WaitForUnlock blocks until given lock is unlocked or timeout is reached
func (db DdbDataStore) WaitForUnlock(lockname string) bool {
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
func (db DdbDataStore) Unlock(lockname string) error {
	lockuuid := uuid.NewV5(db.Accountuuid, lockname)
	DeleteInput := dynamodb.DeleteItemInput{
		TableName: aws.String(db.Tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(lockuuid.String())},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
	}
	_, err := db.Ddb.DeleteItem(&DeleteInput)
	if err != nil {
		return err
	}
	return nil
}

// PutServiceDefinition push catalog service definition to DynamoDb
func (db DdbDataStore) PutServiceDefinition(sd osb.Service) error {
	glog.Infof("putting service definition %q into dynamdb", sd.Name)
	serviceid := uuid.NewV5(db.Accountuuid, sd.Name)
	si, err := dynamodbattribute.Marshal(sd)
	if err != nil {
		glog.Errorln(err)
		return err
	}
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.Tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id":          {S: aws.String(serviceid.String())},
			"userid":      {S: aws.String(db.Accountuuid.String())},
			"serviceid":   {S: aws.String(serviceid.String())},
			"servicename": {S: aws.String(sd.Name)},
			"service":     si,
			"type":        {S: aws.String("service")},
		},
	}
	_, err = db.Ddb.PutItem(&putInput)
	if err != nil {
		glog.Infoln(putInput)
		glog.Errorln(err)
		return err
	}
	glog.Infof("done putting service definition %q into dynamdb", sd.Name)
	return nil
}

// Param stores a parameter value
type Param struct {
	Value string `json:"value"`
}

// GetParam fetches parameter from Dynamo
func (db DdbDataStore) GetParam(paramname string) (value string, err error) {
	paramuuid := uuid.NewV5(db.Accountuuid, paramname).String()
	getInput := dynamodb.GetItemInput{
		TableName: aws.String(db.Tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(paramuuid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
	}
	result, err := db.Ddb.GetItem(&getInput)
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
func (db DdbDataStore) PutParam(paramname string, paramvalue string) error {
	paramuuid := uuid.NewV5(db.Accountuuid, paramname).String()
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.Tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(paramuuid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
			"value":  {S: aws.String(paramvalue)},
			"type":   {S: aws.String("parameter")},
		},
	}
	_, err := db.Ddb.PutItem(&putInput)
	if err != nil {
		return err
	}
	return nil
}

// ServiceItem used to unmarshal catalog entries from DynamoDb
type ServiceItem struct {
	ID          string      `json:"id"`
	Userid      string      `json:"userid"`
	Service     osb.Service `json:"service"`
	Serviceid   string      `json:"serviceid"`
	Servicename string      `json:"servicename"`
}

// GetServiceDefinition fetches given catalog service definition from Dynamo
func (db DdbDataStore) GetServiceDefinition(serviceuuid string) (osb.Service, error) {
	servicedef := osb.Service{}
	getInput := dynamodb.GetItemInput{
		TableName: aws.String(db.Tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(serviceuuid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
	}
	result, err := db.Ddb.GetItem(&getInput)
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
func (db DdbDataStore) GetServiceInstance(sid string) (dbConnector.ServiceInstance, error) {
	var si dbConnector.ServiceInstance
	input := dynamodb.GetItemInput{
		TableName: aws.String(db.Tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(sid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
	}
	outp, err := db.Ddb.GetItem(&input)
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
func (db DdbDataStore) PutServiceInstance(si dbConnector.ServiceInstance) error {
	msi, err := dynamodbattribute.Marshal(si)
	if err != nil {
		return err
	}
	putInput := dynamodb.PutItemInput{
		TableName: aws.String(db.Tablename),
		Item: map[string]*dynamodb.AttributeValue{
			"id":              {S: aws.String(si.ID)},
			"userid":          {S: aws.String(db.Accountuuid.String())},
			"serviceinstance": msi,
			"type":            {S: aws.String("serviceinstance")},
		},
	}
	_, err = db.Ddb.PutItem(&putInput)
	if err != nil {
		return err
	}
	return nil
}
