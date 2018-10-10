package dynamodbadapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/awslabs/aws-servicebroker/pkg/serviceinstance"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	uuid "github.com/satori/go.uuid"
)

// Item types
const (
	itemTypeParameter       = "parameter"
	itemTypeService         = "service"
	itemTypeServiceBinding  = "servicebinding"
	itemTypeServiceInstance = "serviceinstance"
)

// DdbDataStore is a DynamoDB implementation of DataStore.
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
			"type":        {S: aws.String(itemTypeService)},
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
			"type":   {S: aws.String(itemTypeParameter)},
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
			"type":            {S: aws.String(itemTypeServiceInstance)},
		},
	}
	_, err = db.Ddb.PutItem(&putInput)
	if err != nil {
		return err
	}
	return nil
}

// DeleteServiceInstance deletes the service instance.
func (db DdbDataStore) DeleteServiceInstance(sid string) error {
	return db.deleteItem(sid, itemTypeServiceInstance)
}

// GetServiceBinding returns the specified service binding.
func (db DdbDataStore) GetServiceBinding(id string) (*serviceinstance.ServiceBinding, error) {
	resp, err := db.Ddb.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(id)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
		ProjectionExpression: aws.String("servicebinding"),
		TableName:            aws.String(db.Tablename),
	})
	if err != nil {
		return nil, err
	} else if len(resp.Item) == 0 {
		return nil, nil
	}

	var sb serviceinstance.ServiceBinding
	err = dynamodbattribute.Unmarshal(resp.Item["servicebinding"], &sb)
	return &sb, err
}

// PutServiceBinding stores the service binding.
func (db DdbDataStore) PutServiceBinding(sb serviceinstance.ServiceBinding) error {
	msb, err := dynamodbattribute.Marshal(sb)
	if err != nil {
		return err
	}
	_, err = db.Ddb.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":             {S: aws.String(sb.ID)},
			"userid":         {S: aws.String(db.Accountuuid.String())},
			"servicebinding": msb,
			"type":           {S: aws.String(itemTypeServiceBinding)},
		},
		TableName: aws.String(db.Tablename),
	})
	return err
}

// DeleteServiceBinding deletes the service binding.
func (db DdbDataStore) DeleteServiceBinding(id string) error {
	return db.deleteItem(id, itemTypeServiceBinding)
}

func (db DdbDataStore) deleteItem(id, itemType string) error {
	// Ensure the item we're deleting has the expected type
	expr, _ := expression.NewBuilder().
		WithCondition(expression.Name("type").Equal(expression.Value(itemType))).
		Build()

	_, err := db.Ddb.DeleteItem(&dynamodb.DeleteItemInput{
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"id":     {S: aws.String(id)},
			"userid": {S: aws.String(db.Accountuuid.String())},
		},
		TableName: aws.String(db.Tablename),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			glog.Errorf("item %s does not have type %s", id, itemType)
			return nil // Consider this a success since the expected item is gone
		}
		return err
	}
	return nil
}
