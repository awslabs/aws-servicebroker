package broker

import (
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPrescribeOverrides(t *testing.T) {
	assertor := assert.New(t)

	services := []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param":      map[string]interface{}{"type": "string"},
						"override_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param", "override_param"},
					},
				}},
			}},
		}},
	}

	os.Setenv("PARAM_OVERRIDE_awsservicebroker_all_all_all_override_param", "overridden")

	msg := "params should not be modified when prescribeOverrides is false"
	psvcs := prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: false}, services)
	expected := []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param":      map[string]interface{}{"type": "string"},
						"override_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param", "override_param"},
					},
				}},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)

	msg = "override_param should be removed when prescribeOverrides is true"
	psvcs = prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: true}, services)
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{Create: &osb.InputParametersSchema{
					Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
						"req_param": map[string]interface{}{"type": "string"},
					},
						"$schema":  "http://json-schema.org/draft-06/schema#",
						"required": []string{"req_param"},
					},
				}},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)

	services = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param", "override_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param":      map[string]interface{}{"type": "string"},
							"override_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param", "override_param"},
						},
					},
				},
			}},
		}},
	}

	msg = "override_param should be removed from Update params too when prescribeOverrides is true"
	psvcs = prescribeOverrides(AwsBroker{brokerid: "awsservicebroker", prescribeOverrides: true}, services)
	expected = []osb.Service{
		{ID: "test", Name: "test", Description: "test", Plans: []osb.Plan{
			{ID: "testplan", Name: "testplan", Description: "testplan", Schemas: &osb.Schemas{
				ServiceInstance: &osb.ServiceInstanceSchema{
					Create: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
					Update: &osb.InputParametersSchema{
						Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{
							"req_param": map[string]interface{}{"type": "string"},
						},
							"$schema":  "http://json-schema.org/draft-06/schema#",
							"required": []string{"req_param"},
						},
					},
				},
			}},
		}},
	}
	assertor.Equal(expected, psvcs, msg)
}
