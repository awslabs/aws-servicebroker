package dynamodbadapter

import (
	"fmt"

	"github.com/awslabs/aws-service-broker/pkg/serviceinstance"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	uuid "github.com/satori/go.uuid"
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
func (db DdbDataStore) GetServiceDefinition(serviceuuid string) (*osb.Service, error) {
	resp, err := db.Ddb.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(serviceuuid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
		TableName: aws.String(db.Tablename),
	})
	if err != nil {
		return nil, err
	} else if len(resp.Item) == 0 {
		return nil, nil
	}

	var item ServiceItem
	err = dynamodbattribute.UnmarshalMap(resp.Item, &item)
	return &item.Service, err
}

// GetServiceInstance fetches given service instance from Dynamo
func (db DdbDataStore) GetServiceInstance(sid string) (*serviceinstance.ServiceInstance, error) {
	resp, err := db.Ddb.GetItem(&dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true), // Ensure we have the latest version of the service instance
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(sid)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
		ProjectionExpression: aws.String("serviceinstance"),
		TableName:            aws.String(db.Tablename),
	})
	if err != nil {
		return nil, err
	} else if len(resp.Item) == 0 {
		return nil, nil
	}

	var si serviceinstance.ServiceInstance
	err = dynamodbattribute.Unmarshal(resp.Item["serviceinstance"], &si)
	return &si, err
}

// PutServiceInstance stores given service instance in Dynamo
func (db DdbDataStore) PutServiceInstance(si serviceinstance.ServiceInstance) error {
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
