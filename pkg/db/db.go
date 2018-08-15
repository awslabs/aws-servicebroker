package db

import (
	"github.com/go-errors/errors"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/satori/go.uuid"
	"reflect"
)

// Db configuration
type Db struct {
	Accountid     string
	Accountuuid   uuid.UUID
	Brokerid      string
	DataStorePort DataStore
}

// ServiceInstance details of a service instance
type ServiceInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]string
	StackID   string
}

// DataStore port, any backend datastore must provide at least these interfaces
type DataStore interface {
	Lock(lockname string) bool
	IsLocked(lockname string) bool
	WaitForUnlock(lockname string) bool
	Unlock(lockname string) error
	PutServiceDefinition(sd osb.Service) error
	GetParam(paramname string) (value string, err error)
	PutParam(paramname string, paramvalue string) error
	GetServiceDefinition(serviceuuid string) (osb.Service, error)
	GetServiceInstance(sid string) (ServiceInstance, error)
	PutServiceInstance(si ServiceInstance) error
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
				props := make(map[string]interface{})
				required := make([]string, 0)
				for _, param := range k.([]interface{}) {
					name := ""
					isRequired := false
					pvals := make(map[string]interface{})
					for pk, pv := range param.(map[interface{}]interface{}) {
						if pk == "name" {
							name = pv.(string)
						} else if pk == "enum" {
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
								"type":       "object",
								"properties": props,
								"$schema":    "http://json-schema.org/draft-06/schema#",
								"required":   required,
							},
						},
					},
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

func (i *ServiceInstance) Match(other *ServiceInstance) bool {
	return reflect.DeepEqual(i, other)
}
