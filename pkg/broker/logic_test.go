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
	assert.Equal(t, ssmclient.Endpoint, "https://ssm.us-east-2.amazonaws.com", "Checking cfn endpoint")
}
