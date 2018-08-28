package broker

import (
	"fmt"
	"github.com/awslabs/aws-service-broker/pkg/serviceinstance"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"net/http"
	"strings"
)

// GetCatalog is executed on a /v2/catalog/ osb api call
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management
func (b *AwsBroker) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	response := &broker.CatalogResponse{}

	var services []osb.Service
	l, _ := b.listingcache.Get("__LISTINGS__")
	glog.Infoln(l)
	for _, s := range l.([]ServiceNeedsUpdate) {
		sd, err := b.catalogcache.Get(s.Name)
		if err != nil {
			if err.Error() == "not found" {
				glog.Errorf("Failed to fetch %q from the cache, item not found", s.Name)
			} else {
				glog.Errorln(err)
			}
		} else {
			services = append(services, sd.(osb.Service))
			glog.Infof("ServiceClass: %q %q", sd.(osb.Service).Name, sd.(osb.Service).ID)
			for _, plan := range sd.(osb.Service).Plans {
				glog.Infof("  ServicePlan %q %q", plan.Name, plan.ID)
			}
		}
	}
	osbResponse := &osb.CatalogResponse{Services: services}

	//glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

// Provision is executed when the osb api receives PUT /v2/service_instances/:instance_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#provisioning
func (b *AwsBroker) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	lockid := "serviceInstance__provision__" + request.InstanceID
	gotlock := b.db.DataStorePort.Lock(lockid)
	if gotlock == false {
		if b.db.DataStorePort.WaitForUnlock(lockid) == false {
			gotlock = b.db.DataStorePort.Lock(lockid)
		}
	}
	if gotlock {
		response := broker.ProvisionResponse{}
		instance := &serviceinstance.ServiceInstance{
			ID:        request.InstanceID,
			ServiceID: request.ServiceID,
			PlanID:    request.PlanID,
		}
		if request.AcceptsIncomplete {
			response.Async = true
		}
		servicedef, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
		var plandef osb.Plan
		for _, v := range servicedef.Plans {
			if v.ID == request.PlanID {
				plandef = v
			}
		}
		if err != nil {
			panic(err)
		}
		i, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)

		// Check to see if this is the same instance
		if err != nil {
			panic(err)
		} else if i.ID != "" {
			if i.Match(instance) {
				response.Exists = true
				return &response, nil
			}
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		} else {
			var tags []*cloudformation.Tag
			tags = append(tags, &cloudformation.Tag{
				Key:   aws.String("ServiceBrokerId"),
				Value: aws.String(b.region + "::" + b.brokerid),
			})
			tags = append(tags, &cloudformation.Tag{
				Key:   aws.String("ServiceBrokerInstanceId"),
				Value: aws.String(instance.ID),
			})
			var Cap []*string
			Cap = append(Cap, aws.String("CAPABILITY_IAM"))
			Cap = append(Cap, aws.String("CAPABILITY_NAMED_IAM"))
			//glog.Infoln(plandef.Schemas.ServiceInstance.Create.Parameters)

			params := getParams(plandef.Schemas.ServiceInstance.Create.Parameters)
			glog.V(10).Infoln(params)
			ns := "all"
			cluster := "all"
			if _, ok := request.Context["platform"]; ok {
				if request.Context["platform"].(string) == "cloudfoundry" {
					ns = request.Context["space_guid"].(string)
					ns = strings.Replace(ns, "-", "", -1)
					cluster = request.Context["organization_guid"].(string)
					cluster = strings.Replace(cluster, "-", "", -1)
				} else if request.Context["platform"].(string) == "kubernetes" {
					ns = request.Context["namespace"].(string)
					cluster = request.Context["clusterid"].(string)
				}
			}
			completeparams := getOverrides(b.brokerid, params, ns, servicedef.Name, cluster)
			glog.V(10).Infoln(completeparams)
			for k, p := range request.Parameters {
				completeparams[k] = p.(string)
			}
			glog.V(10).Infoln(completeparams)
			cfnsvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, completeparams))
			instance.Params = make(map[string]string)
			for k, v := range completeparams {
				instance.Params[k] = v
			}
			glog.V(10).Infoln(instance.Params)
			nonCfnParamarams := []string{
				"aws_access_key",
				"aws_secret_key",
				"aws_session_token",
				"SBArtifactS3KeyPrefix",
				"SBArtifactS3Bucket",
				"region",
				"target_role_name",
				"target_account_id"}
			for _, k := range nonCfnParamarams {
				if _, ok := completeparams[k]; ok {
					delete(completeparams, k)
				}
			}
			var inputParams []*cloudformation.Parameter
			for k, p := range completeparams {
				param := cloudformation.Parameter{
					ParameterKey:   aws.String(k),
					ParameterValue: aws.String(p),
				}
				glog.V(10).Infoln(param)
				glog.V(10).Infof("%q: %q\n", k, p)
				inputParams = append(inputParams, &param)
			}
			glog.Infof("Input Parmas '%+v'", inputParams)
			stackInput := cloudformation.CreateStackInput{
				Capabilities: Cap,
				Parameters:   inputParams,
				StackName:    aws.String("CfnServiceBroker-" + servicedef.Name + "-" + instance.ID),
				Tags:         tags,
				TemplateURL:  b.generateS3HTTPUrl(servicedef.Name),
			}
			glog.V(10).Infoln(stackInput)
			results, err := cfnsvc.CreateStack(&stackInput)
			if err != nil {
				glog.Errorln(err)
				b.db.DataStorePort.Unlock(lockid)
				return &response, err
			}
			instance.StackID = *results.StackId
			err = b.db.DataStorePort.PutServiceInstance(*instance)
			if err != nil {
				glog.Errorln(err)
				b.db.DataStorePort.Unlock(lockid)
				return &response, err
			}
		}
		b.db.DataStorePort.Unlock(lockid)
		return &response, nil
	}
	description := "Failed to get lock for instanceId" + string(request.InstanceID)
	return nil, osb.HTTPStatusCodeError{
		StatusCode:  http.StatusExpectationFailed,
		Description: &description,
	}
}

// Deprovision executed when the osb api receives DELETE /v2/service_instances/:instance_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#deprovisioning
func (b *AwsBroker) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	lockid := "serviceInstance__deprovision__" + request.InstanceID
	gotlock := b.db.DataStorePort.Lock(lockid)
	response := broker.DeprovisionResponse{}
	if gotlock == false {
		if b.db.DataStorePort.WaitForUnlock(lockid) == false {
			gotlock = b.db.DataStorePort.Lock(lockid)
		}
	}
	if gotlock {
		si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
		if err != nil {
			panic(err)
		}
		if si.StackID == "" {
			errmsg := "CloudFormation stackid missing, chances are stack creation failed in an unexpected way, assuming there is nothing to deprovision"
			glog.Errorln(errmsg)
			response.Async = false
			return &response, nil
		}
		glog.V(10).Infoln(si.Params)
		cfnsvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, si.Params))
		_, err = cfnsvc.DeleteStack(&cloudformation.DeleteStackInput{StackName: aws.String(si.StackID)})

		if err != nil {
			panic(err)
		}

		b.db.DataStorePort.Unlock(lockid)

		if err != nil {
			panic(err)
		}
		if request.AcceptsIncomplete {
			response.Async = true
		}
		return &response, nil
	}
	description := "Failed to get lock for instanceId" + string(request.InstanceID)
	return nil, osb.HTTPStatusCodeError{
		StatusCode:  http.StatusExpectationFailed,
		Description: &description,
	}
}

// LastOperation executed when the osb api receives GET /v2/service_instances/:instance_id/last_operation
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#polling-last-operation
func (b *AwsBroker) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.Infoln(request)
	glog.Infoln(c)
	si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		panic(err)
	}
	glog.Infoln(si)
	r := broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "", Description: nil}}
	if si.StackID == "" {
		errmsg := "CloudFormation stackid missing, chances are stack creation failed in an unexpected way"
		glog.Errorln(errmsg)
		r.LastOperationResponse.State = "failed"
		r.LastOperationResponse.Description = &errmsg
		return &r, nil
	}
	glog.V(10).Infoln(si.Params)
	cfnsvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, si.Params))
	response, err := cfnsvc.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(si.StackID)})
	if err != nil {
		panic(err)
	}
	failedstates := []string{"CREATE_FAILED", "ROLLBACK_IN_PROGRESS", "ROLLBACK_FAILED", "ROLLBACK_COMPLETE", "DELETE_FAILED", "UPDATE_ROLLBACK_IN_PROGRESS", "UPDATE_ROLLBACK_FAILED", "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"}
	progressingstates := []string{"CREATE_IN_PROGRESS", "DELETE_IN_PROGRESS", "UPDATE_IN_PROGRESS", "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"}
	successfulstates := []string{"CREATE_COMPLETE", "DELETE_COMPLETE", "UPDATE_COMPLETE"}
	status := *response.Stacks[0].StackStatus
	if stringInSlice(status, failedstates) {
		glog.Errorf("CloudFormation stack failed %#+v", si.StackID)
		glog.Errorf(status)
		r.LastOperationResponse.State = "failed"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else if stringInSlice(status, progressingstates) {
		glog.Infoln("CloudFormation stack still busy...")
		glog.Infoln(status)
		r.LastOperationResponse.State = "in progress"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else if stringInSlice(status, successfulstates) {
		glog.Infoln("CloudFormation stack operation completed...")
		glog.Infoln(status)
		r.LastOperationResponse.State = "succeeded"
		r.LastOperationResponse.Description = response.Stacks[0].StackStatusReason
		return &r, nil
	} else {
		return nil, fmt.Errorf("unexpected cfn status %v", status)
	}
}

// Bind executed when the osb api receives PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#request-4
func (b *AwsBroker) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {

	si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	service, err := b.db.DataStorePort.GetServiceDefinition(si.ServiceID)
	if err != nil {
		panic(err)
	}
	glog.Infoln(si)
	sess := b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, si.Params)
	cfnsvc := b.Clients.NewCfn(sess)
	ssmsvc := b.Clients.NewSsm(sess)
	cfnresponse, err := cfnsvc.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(si.StackID)})
	if err != nil {
		panic(err)
	}
	outputs := make(map[string]interface{})
	for _, o := range cfnresponse.Stacks[0].Outputs {
		fmt.Println(o)
		if *o.OutputKey == "UserKeyId" || *o.OutputKey == "UserSecretKey" {
			ssmInput := ssm.GetParameterInput{
				Name:           aws.String(*o.OutputValue),
				WithDecryption: aws.Bool(true),
			}

			ssmresponse, err := ssmsvc.GetParameter(&ssmInput)
			if err != nil {
				panic(err)
			}
			pname := strings.ToUpper(service.Name) + "_" + toSnakeCase(*o.OutputKey)
			outputs[pname] = ssmresponse.Parameter.Value
		} else {
			outputs[toSnakeCase(*o.OutputKey)] = o.OutputValue
		}
	}
	glog.Infoln(outputs)
	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: outputs,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = false
	}
	return &response, nil
}

func (b *AwsBroker) GetBinding(request *osb.GetBindingRequest, c *broker.RequestContext) (*broker.GetBindingResponse, error) {
	glog.V(10).Infoln(request)
	glog.V(10).Infoln(c)
	return &broker.GetBindingResponse{}, nil
}

func BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.V(10).Infoln(request)
	glog.V(10).Infoln(c)
	return &broker.LastOperationResponse{}, nil
}

// Unbind executed when the osb api receives DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#unbinding
func (b *AwsBroker) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

// Update is not supported at present, so is just a skeleton
func (b *AwsBroker) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = true
	}
	return &response, nil
}

func (b *AwsBroker) BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	return &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "", Description: nil}}, nil
}
