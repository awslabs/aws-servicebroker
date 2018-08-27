package broker

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"testing"
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

func TestAssumeRoleRegion(t *testing.T) {
	bl := BusinessLogic{region: "us-east-2"}
	cfnclient, ssmclient := bl.getAwsClient(map[string]string{})
	assert.Equal(t, cfnclient.Endpoint, "https://cloudformation.us-east-2.amazonaws.com", "Checking cfn endpoint")
	assert.Equal(t, ssmclient.Endpoint, "https://ssm.us-east-2.amazonaws.com", "Checking ssm endpoint")
}

func TestAssumeArnGeneration(t *testing.T) {
	params := map[string]string{"target_role_name": "worker"}
	accountId := "123456654321"
	assert.Equal(t, generateRoleArn(params, accountId), "arn:aws:iam::123456654321:role/worker", "Validate role arn")
	params["target_account_id"] = "000000000000"
	assert.Equal(t, generateRoleArn(params, accountId), "arn:aws:iam::000000000000:role/worker", "Validate role arn")
}
