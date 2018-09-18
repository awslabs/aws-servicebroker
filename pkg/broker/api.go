package broker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/awslabs/aws-service-broker/pkg/serviceinstance"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

// GetCatalog is executed on a /v2/catalog/ osb api call
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management
func (b *AwsBroker) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	response := &broker.CatalogResponse{}

	var services []osb.Service
	l, _ := b.listingcache.Get("__LISTINGS__")
	glog.V(10).Infoln(l)
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
	osbResponse := &osb.CatalogResponse{Services: prescribeOverrides(*b, services)}

	//glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

// Provision is executed when the OSB API receives `PUT /v2/service_instances/:instance_id`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#provisioning).
func (b *AwsBroker) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	if !request.AcceptsIncomplete {
		return nil, newAsyncError()
	}

	// Get the context
	cluster := getCluster(request.Context)
	namespace := getNamespace(request.Context)

	// Get the service
	service, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service %s: %v", request.ServiceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if service == nil {
		desc := fmt.Sprintf("The service %s was not found.", request.ServiceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the plan
	plan := getPlan(service, request.PlanID)
	if plan == nil {
		desc := fmt.Sprintf("The service plan %s was not found.", request.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the parameters and verify that all required parameters are set
	params := getPlanDefaults(plan)
	availableParams := getAvailableParams(plan)
	for k, v := range getOverrides(b.brokerid, availableParams, namespace, service.Name, cluster) {
		params[k] = v
	}
	for k, v := range request.Parameters {
		if !stringInSlice(k, availableParams) {
			desc := fmt.Sprintf("The parameter %s is not available.", k)
			return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
		}
		params[k] = paramValue(v)
	}
	for _, p := range getRequiredParams(plan) {
		if _, ok := params[p]; !ok {
			desc := fmt.Sprintf("The parameter %s is required.", p)
			return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
		}
	}
	glog.V(10).Infof("params=%v", params)

	instance := &serviceinstance.ServiceInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		Params:    params,
		PlanID:    request.PlanID,
	}

	// Verify that the instance doesn't already exist
	i, err := b.db.DataStorePort.GetServiceInstance(instance.ID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service instance %s: %v", instance.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if i != nil {
		// TODO: This logic could use some love. The docs state that 200 OK MUST be
		// returned if the service instance already exists, is fully provisioned,
		// and the requested parameters are identical to the existing service
		// instance. Right now, this doesn't check whether the instance is fully
		// provisioned, and the reflect.DeepEqual check in Match will return false
		// if the parameter order is different.
		if i.Match(instance) {
			glog.Infof("Service instance %s already exists.", instance.ID)
			response := broker.ProvisionResponse{}
			response.Exists = true
			return &response, nil
		}
		glog.V(10).Infof("i=%+v instance=%+v", *i, *instance)
		desc := fmt.Sprintf("Service instance %s already exists but with different attributes.", instance.ID)
		return nil, newHTTPStatusCodeError(http.StatusConflict, "", desc)
	}

	// Create the CFN stack
	cfnSvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, params))
	resp, err := cfnSvc.CreateStack(&cloudformation.CreateStackInput{
		Capabilities: aws.StringSlice([]string{cloudformation.CapabilityCapabilityNamedIam}),
		Parameters:   toCFNParams(params),
		StackName:    aws.String(getStackName(service.Name, instance.ID)),
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("aws-service-broker:broker-id"),
				Value: aws.String(b.brokerid),
			},
			{
				Key:   aws.String("aws-service-broker:instance-id"),
				Value: aws.String(request.InstanceID),
			},
			{
				Key:   aws.String("aws-service-broker:cluster"),
				Value: aws.String(cluster),
			},
			{
				Key:   aws.String("aws-service-broker:namespace"),
				Value: aws.String(namespace),
			},
		},
		TemplateURL: b.generateS3HTTPUrl(service.Name),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to create the CloudFormation stack: %v", err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	instance.StackID = aws.StringValue(resp.StackId)
	err = b.db.DataStorePort.PutServiceInstance(*instance)
	if err != nil {
		// Try to delete the stack
		if _, err := cfnSvc.DeleteStack(&cloudformation.DeleteStackInput{StackName: aws.String(instance.StackID)}); err != nil {
			glog.Errorf("Failed to delete the CloudFormation stack %s: %v", instance.StackID, err)
		}

		desc := fmt.Sprintf("Failed to create the service instance %s: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	response := broker.ProvisionResponse{}
	response.Async = true
	return &response, nil
}

// Deprovision executed when the osb api receives DELETE /v2/service_instances/:instance_id
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#deprovisioning
func (b *AwsBroker) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	response := broker.DeprovisionResponse{}
	si, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		panic(err)
	}
	if si == nil || si.StackID == "" {
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

	if request.AcceptsIncomplete {
		response.Async = true
	}
	return &response, nil
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
	if si == nil || si.StackID == "" {
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
	failedstates := []string{"CREATE_FAILED", "ROLLBACK_IN_PROGRESS", "ROLLBACK_FAILED", "ROLLBACK_COMPLETE", "DELETE_FAILED", "UPDATE_ROLLBACK_IN_PROGRESS", "UPDATE_ROLLBACK_FAILED", "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS", "UPDATE_ROLLBACK_COMPLETE"}
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

// Bind is executed when the OSB API receives `PUT /v2/service_instances/:instance_id/service_bindings/:binding_id`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#request-4).
func (b *AwsBroker) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	binding := &serviceinstance.ServiceBinding{
		ID:         request.BindingID,
		InstanceID: request.InstanceID,
	}

	// Get the binding params
	for k, v := range request.Parameters {
		if strings.EqualFold(k, bindParamRoleName) {
			binding.RoleName = paramValue(v)
		} else if strings.EqualFold(k, bindParamScope) {
			binding.Scope = paramValue(v)
		} else {
			desc := fmt.Sprintf("The parameter %s is not supported.", k)
			return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
		}
	}

	// Verify that the binding doesn't already exist
	sb, err := b.db.DataStorePort.GetServiceBinding(binding.ID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service binding %s: %v", binding.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if sb != nil {
		if sb.Match(binding) {
			glog.Infof("Service binding %s already exists.", binding.ID)
			response := broker.BindResponse{}
			response.Exists = true
			return &response, nil
		}
		desc := fmt.Sprintf("Service binding %s already exists but with different attributes.", binding.ID)
		return nil, newHTTPStatusCodeError(http.StatusConflict, "", desc)
	}

	// Get the service (this is only required because the USER_KEY_ID and
	// USER_SECRET_KEY credentials need to be prefixed with the service name for
	// backward compatibility)
	service, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service %s: %v", request.ServiceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if service == nil {
		desc := fmt.Sprintf("The service %s was not found.", request.ServiceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the instance
	instance, err := b.db.DataStorePort.GetServiceInstance(binding.InstanceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service instance %s: %v", binding.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if instance == nil {
		desc := fmt.Sprintf("The service instance %s was not found.", binding.InstanceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	sess := b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, instance.Params)

	// Get the CFN stack outputs
	resp, err := b.Clients.NewCfn(sess).DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(instance.StackID),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to describe the CloudFormation stack %s: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	// Get the credentials from the CFN stack outputs
	credentials, err := getCredentials(service, resp.Stacks[0].Outputs, b.Clients.NewSsm(sess))
	if err != nil {
		desc := fmt.Sprintf("Failed to get the credentials from CloudFormation stack %s: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	if binding.RoleName != "" {
		policyArn, err := getPolicyArn(resp.Stacks[0].Outputs, binding.Scope)
		if err != nil {
			desc := fmt.Sprintf("The CloudFormation stack %s does not support binding with scope '%s': %v", instance.StackID, binding.Scope, err)
			return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
		}

		// Attach the scoped policy to the role
		_, err = b.Clients.NewIam(sess).AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyArn),
			RoleName:  aws.String(binding.RoleName),
		})
		if err != nil {
			desc := fmt.Sprintf("Failed to attach the policy %s to role %s: %v", policyArn, binding.RoleName, err)
			return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
		}

		binding.PolicyArn = policyArn
	}

	// Store the binding
	err = b.db.DataStorePort.PutServiceBinding(*binding)
	if err != nil {
		desc := fmt.Sprintf("Failed to store the service binding %s: %v", binding.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	return &broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: credentials,
		},
	}, nil
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

// Unbind is executed when the OSB API receives `DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#request-5).
func (b *AwsBroker) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	// Get the binding
	binding, err := b.db.DataStorePort.GetServiceBinding(request.BindingID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service binding %s: %v", request.BindingID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if binding == nil {
		desc := fmt.Sprintf("The service binding %s was not found.", request.BindingID)
		return nil, newHTTPStatusCodeError(http.StatusGone, "", desc)
	}

	if binding.PolicyArn != "" {
		instance, err := b.db.DataStorePort.GetServiceInstance(binding.InstanceID)
		if err != nil {
			desc := fmt.Sprintf("Failed to get the service instance %s: %v", binding.InstanceID, err)
			return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
		} else if instance == nil {
			desc := fmt.Sprintf("The service instance %s was not found.", binding.InstanceID)
			return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
		}

		sess := b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, instance.Params)

		// Detach the scoped policy from the role
		_, err = b.Clients.NewIam(sess).DetachRolePolicy(&iam.DetachRolePolicyInput{
			PolicyArn: aws.String(binding.PolicyArn),
			RoleName:  aws.String(binding.RoleName),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeNoSuchEntityException {
				glog.Infof("The policy %s was already detached from role %s.", binding.PolicyArn, binding.RoleName)
			} else {
				desc := fmt.Sprintf("Failed to detach the policy %s from role %s: %v", binding.PolicyArn, binding.RoleName, err)
				return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
			}
		}
	}

	// Delete the binding
	err = b.db.DataStorePort.DeleteServiceBinding(binding.ID)
	if err != nil {
		desc := fmt.Sprintf("Failed to delete the service binding %s: %v", binding.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	return &broker.UnbindResponse{}, nil
}

// Update is executed when the OSB API receives `PATCH /v2/service_instances/:instance_id`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#updating-a-service-instance).
func (b *AwsBroker) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	if !request.AcceptsIncomplete {
		return nil, newAsyncError()
	}

	// Get the service instance
	instance, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service instance %s: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if instance == nil {
		desc := fmt.Sprintf("The service instance %s was not found.", request.InstanceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Verify that we're not changing the plan (this should never happen since
	// we're setting `plan_updateable: false`, but better safe than sorry)
	if request.PlanID != nil && *request.PlanID != instance.PlanID {
		desc := fmt.Sprintf("The service plan cannot be changed from %s to %s.", instance.PlanID, *request.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the service
	service, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service %s: %v", request.ServiceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if service == nil {
		desc := fmt.Sprintf("The service %s was not found.", request.ServiceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the plan and verify that it has updatable parameters
	plan := getPlan(service, instance.PlanID)
	if plan == nil {
		desc := fmt.Sprintf("The service plan %s was not found.", instance.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	} else if plan.Schemas.ServiceInstance.Update == nil {
		desc := fmt.Sprintf("The service plan %s has no updatable parameters.", instance.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the parameters
	params := getPlanDefaults(plan)
	paramsUpdated := false
	updatableParams := getUpdatableParams(plan)
	for k, v := range instance.Params {
		params[k] = v
	}
	for k, v := range request.Parameters {
		newValue := paramValue(v)
		if params[k] != newValue {
			if !stringInSlice(k, updatableParams) {
				desc := fmt.Sprintf("The parameter %s is not updatable.", k)
				return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
			}
			params[k] = newValue
			paramsUpdated = true
		}
	}
	if !paramsUpdated {
		// Nothing to do, so return success (if we try a CFN update, it'll fail)
		return &broker.UpdateInstanceResponse{}, nil
	}
	glog.V(10).Infof("params=%v", params)

	// Update the CFN stack
	cfnSvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, params))
	_, err = cfnSvc.UpdateStack(&cloudformation.UpdateStackInput{
		Capabilities: aws.StringSlice([]string{cloudformation.CapabilityCapabilityNamedIam}),
		Parameters:   toCFNParams(params),
		StackName:    aws.String(instance.StackID),
		TemplateURL:  b.generateS3HTTPUrl(service.Name),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to update the CloudFormation stack %s: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	// Update the params in the DB
	instance.Params = params
	err = b.db.DataStorePort.PutServiceInstance(*instance)
	if err != nil {
		// Try to cancel the update
		if _, err := cfnSvc.CancelUpdateStack(&cloudformation.CancelUpdateStackInput{StackName: aws.String(instance.StackID)}); err != nil {
			glog.Errorf("Failed to cancel updating the CloudFormation stack %s: %v", instance.StackID, err)
			glog.Errorf("Service instance %s and CloudFormation stack %s may be out of sync!", instance.ID, instance.StackID)
		}

		desc := fmt.Sprintf("Failed to update the service instance %s: %v", instance.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	response := broker.UpdateInstanceResponse{}
	response.Async = true
	return &response, nil
}

func (b *AwsBroker) BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	return &broker.LastOperationResponse{LastOperationResponse: osb.LastOperationResponse{State: "", Description: nil}}, nil
}
