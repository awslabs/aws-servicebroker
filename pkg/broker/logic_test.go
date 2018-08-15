package broker

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
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

func TestNewBusinessLogic(t *testing.T) {
	assert := assert.New(t)
	options := new(TestCases)
	options.GetTests("../../testcases/options.yaml")
	spew.Dump(options)
	for k, v := range *options {
		expectSuccess := true
		if strings.HasSuffix(k, "Bad") {
			expectSuccess = false
		}
		fmt.Println(expectSuccess)
		_, err := NewBusinessLogic(v)
		assert.Nil(err)
	}
}
