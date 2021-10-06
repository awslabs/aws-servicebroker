package broker

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/awslabs/aws-servicebroker/pkg/serviceinstance"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

const (
	instanceId = "INSTANCE_ID"
	bindingId = "BINDING_ID"
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
	glog.V(10).Infof("plan=%v", plan)
	if plan == nil {
		desc := fmt.Sprintf("The service plan %s was not found.", request.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the parameters and verify that all required parameters are set
	params := getPlanDefaults(plan)
	glog.V(10).Infof("params=%v", params)
	availableParams := getAvailableParams(plan)
	glog.V(10).Infof("availableParams=%v", availableParams)
	for k, v := range getOverrides(b.brokerid, availableParams, namespace, service.Name, cluster) {
		params[k] = v
	}
	glog.V(10).Infof("params=%v", params)
	for k, v := range getPlanPrescribedParams(plan.Schemas.ServiceInstance.Create.Parameters) {
		params[k] = paramValue(v)
	}
	glog.V(10).Infof("params=%v", params)
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
		if i.Match(instance) {
			glog.Infof("Service instance %s already exists.", instance.ID)

			glog.Infof("Checking if service instance %s is fully provisioned.", instance.ID)
			cfnSvc := getCfn(b, instance.Params)

			response := broker.ProvisionResponse{}
			status, _, err := getStackStatusAndReason(cfnSvc, i.StackID)

			switch {
			case err != nil:
				desc := fmt.Sprintf("Failed to get the stack %s: %v", i.StackID, err)
				return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
			case instanceIsFullyProvisioned(status):
				glog.Infof("Service instance %s is fully provisioned.", instance.ID)
				response.Exists = true
				return &response, nil
			case instanceProvisioningIsInProgress(status):
				glog.Infof("Service instance %s provisioning is in progress.", instance.ID)
				response.Async = true
				return &response, nil
			default:
				glog.Infof("Service instance %s provisioning failed.", instance.ID)
				return &response, newHTTPStatusCodeError(http.StatusBadRequest, "CloudFormationError", *getCfnError(i.StackID, cfnSvc))
			}
		}
		glog.V(10).Infof("i=%+v instance=%+v", *i, *instance)
		desc := fmt.Sprintf("Service instance %s already exists but with different attributes.", instance.ID)
		return nil, newHTTPStatusCodeError(http.StatusConflict, "", desc)
	}

	tags, err := buildTags(b.brokerid, request.InstanceID, cluster, namespace, params)
	if err != nil {
		desc := fmt.Sprintf("failed to parse tags: %v", err)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	urlP := b.generateS3HTTPUrl(service.Name)
	stackName := getStackName(service.Name, instance.ID)
	capabilities := []string{cloudformation.CapabilityCapabilityNamedIam}
	cfnParams := toCFNParams(params)

	// Create the CFN stack
	cfnSvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, params))
	resp, err := cfnSvc.Client.CreateStack(&cloudformation.CreateStackInput{
		Capabilities: aws.StringSlice(capabilities),
		Parameters:   cfnParams,
		StackName:    aws.String(stackName),
		Tags:         tags,
		TemplateURL:  urlP,
	})
	if err != nil {
		var url string
		if urlP != nil {
			url = *urlP
		}
		glog.Errorf("TemplateURL: %s", url)
		glog.Errorf("Tags: %v", tags)
		glog.Errorf("StackName:  %s", stackName)
		glog.Errorf("Parameters: %v", toCFNParams(params))
		glog.Errorf("Capabilities: %v ", capabilities)
		desc := fmt.Sprintf("Failed to create the CloudFormation stack: %v", err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	instance.StackID = aws.StringValue(resp.StackId)
	err = b.db.DataStorePort.PutServiceInstance(*instance)
	if err != nil {
		// Try to delete the stack
		if _, err := cfnSvc.Client.DeleteStack(&cloudformation.DeleteStackInput{StackName: aws.String(instance.StackID)}); err != nil {
			glog.Errorf("Failed to delete the CloudFormation stack %s: %v", instance.StackID, err)
		}

		desc := fmt.Sprintf("Failed to create the service instance %s: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	response := broker.ProvisionResponse{}
	response.Async = true
	return &response, nil
}

// Deprovision is executed when the OSB API receives `DELETE /v2/service_instances/:instance_id`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#deprovisioning).
func (b *AwsBroker) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	if !request.AcceptsIncomplete {
		return nil, newAsyncError()
	}

	// Get the instance
	instance, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service instance %s: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if instance == nil {
		desc := fmt.Sprintf("The service instance %s was not found.", request.InstanceID)
		return nil, newHTTPStatusCodeError(http.StatusGone, "", desc)
	}

	// Delete the CFN stack
	cfnSvc := b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, instance.Params))
	if _, err := cfnSvc.Client.DeleteStack(&cloudformation.DeleteStackInput{StackName: aws.String(instance.StackID)}); err != nil {
		desc := fmt.Sprintf("Failed to delete the CloudFormation stack %s: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	response := broker.DeprovisionResponse{}
	response.Async = true
	return &response, nil
}

// LastOperation is executed when the OSB API receives `GET /v2/service_instances/:instance_id/last_operation`
// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#polling-last-operation).
func (b *AwsBroker) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	glog.V(10).Infof("request=%+v", *request)

	// Get the instance
	instance, err := b.db.DataStorePort.GetServiceInstance(request.InstanceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service instance %s: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if instance == nil {
		// Returning 410 Gone here is only appropriate for asynchronous delete
		// operations, so hopefully that's what this operation is :)
		// (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#response-1)
		desc := fmt.Sprintf("The service instance %s was not found.", request.InstanceID)
		return nil, newHTTPStatusCodeError(http.StatusGone, "", desc)
	}

	// Get the CFN stack status
	cfnSvc := getCfn(b, instance.Params)

	status, reason, err := getStackStatusAndReason(cfnSvc, instance.StackID)
	if err != nil {
		return nil, err
	}

	response := broker.LastOperationResponse{}
	if instanceIsFullyProvisioned(status) {
		response.State = osb.StateSucceeded
		if status == cloudformation.StackStatusDeleteComplete {
			// If the resources were successfully deleted, try to delete the instance
			if err := b.db.DataStorePort.DeleteServiceInstance(instance.ID); err != nil {
				glog.Errorf("Failed to delete the service instance %s: %v", instance.ID, err)
			}
		}
	} else if instanceProvisioningIsInProgress(status) {
		response.State = osb.StateInProgress
	} else {
		glog.Errorf("CloudFormation stack %s failed with status %s: %s", instance.StackID, status, reason)
		response := broker.LastOperationResponse{}
		response.State = osb.StateFailed
		response.Description = getCfnError(instance.StackID, cfnSvc)
		if *response.Description == "" {
			response.Description = &reason
		}

		// workaround for https://github.com/kubernetes-incubator/service-catalog/issues/2505
		originatingIdentity := strings.Split(c.Request.Header.Get("X-Broker-Api-Originating-Identity"), " ")[0]
		if originatingIdentity == "kubernetes" {
			return &response, newHTTPStatusCodeError(http.StatusBadRequest, "CloudFormationError", *response.Description)
		}
	}
	return &response, nil
}

func getCfn(b *AwsBroker, instanceParams map[string]string) CfnClient {
	return b.Clients.NewCfn(b.GetSession(b.keyid, b.secretkey, b.region, b.accountId, b.profile, instanceParams))
}

func getStackStatusAndReason(cfnSvc CfnClient, stackId string) (status string, reason string, err error) {
	resp, err := cfnSvc.Client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackId),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to describe the CloudFormation stack %s: %v", stackId, err)
		return "", "", newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}
	status = aws.StringValue(resp.Stacks[0].StackStatus)
	reason = aws.StringValue(resp.Stacks[0].StackStatusReason)
	glog.V(10).Infof("stack=%s status=%s reason=%s", stackId, status, reason)

	return
}

func instanceIsFullyProvisioned(status string) bool {
	return status == cloudformation.StackStatusCreateComplete ||
		status == cloudformation.StackStatusDeleteComplete ||
		status == cloudformation.StackStatusUpdateComplete
}

func instanceProvisioningIsInProgress(status string) bool {
	return strings.HasSuffix(status, "_IN_PROGRESS") && !strings.Contains(status, "ROLLBACK")
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
	resp, err := b.Clients.NewCfn(sess).Client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(instance.StackID),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to describe the CloudFormation stack %s: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	// Verify that the binding doesn't already exist
	sb, err := b.db.DataStorePort.GetServiceBinding(binding.ID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service binding %s: %v", binding.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if sb != nil {
		if sb.Match(binding) {
			glog.Infof("Service binding %s already exists.", binding.ID)
			return createBindResponse(service, resp, b, sess, instance, binding)
		}
		desc := fmt.Sprintf("Service binding %s already exists but with different attributes.", binding.ID)
		return nil, newHTTPStatusCodeError(http.StatusConflict, "", desc)
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

	credentials, err := getBindingCredentials(service, resp, b, sess, instance, binding)
	if err != nil {
		return nil, err
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

func createBindResponse(
	service *osb.Service,
	resp *cloudformation.DescribeStacksOutput,
	b *AwsBroker,
	sess *session.Session,
	instance *serviceinstance.ServiceInstance,
	binding *serviceinstance.ServiceBinding) (*broker.BindResponse, error) {

	response := broker.BindResponse{}
	response.Exists = true

	credentials, err := getBindingCredentials(service, resp, b, sess, instance, binding)
	if err != nil {
		return nil, err
	}

	response.Credentials = credentials
	return &response, nil
}

func getBindingCredentials(
	service *osb.Service,
	resp *cloudformation.DescribeStacksOutput,
	b *AwsBroker,
	sess *session.Session,
	instance *serviceinstance.ServiceInstance,
	binding *serviceinstance.ServiceBinding) (map[string]interface{}, error) {

	// Get the credentials from the CFN stack outputs
	credentials, err := getCredentials(service, resp.Stacks[0].Outputs, b.Clients.NewSsm(sess))
	if err != nil {
		desc := fmt.Sprintf("Failed to get the credentials from CloudFormation stack %s: %v", instance.StackID, err)
		glog.Error(desc)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	if bindViaLambda(service) {
		// Copy instance and binding IDs into credentials to
		// be used as identifiers for resources we create in
		// lambda so that we can reference them when we unbind
		// (for example, you can build a unique path for an
		// IAM User with this information, and avoid the need
		// to have persist extra identifiers, or have users
		// provide them.
		credentials[instanceId] = binding.InstanceID
		credentials[bindingId] = binding.ID

		// Replace credentials with a derived set calculated by a lambda function
		credentials, err = invokeLambdaBindFunc(sess, b.Clients.NewLambda, credentials, "bind")
		if err != nil {
			glog.Error(err)
			return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", err.Error())
		}
	}
	return credentials, nil
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

	if bindViaLambda(service) {

		// Get the CFN stack outputs
		resp, err := b.Clients.NewCfn(sess).Client.DescribeStacks(&cloudformation.DescribeStacksInput{
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

		// Copy in the instance and binding IDs because we can
		// use this as a stable reference to uniquely identify
		// dynamically created resources (for exmaple, you can
		// use these in the Path prefix of an IAM User).
		credentials["INSTANCE_ID"] = binding.InstanceID
		credentials["BINDING_ID"] = binding.ID
		_, err = invokeLambdaBindFunc(sess, b.Clients.NewLambda, credentials, "unbind")
		if err != nil {
			desc := fmt.Sprintf("Error running lambda function for unbind from: %vo", err)
			return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
		}
	}

	if binding.PolicyArn != "" {
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
		desc := fmt.Sprintf("Failed to get the service instance %q: %v", request.InstanceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if instance == nil {
		desc := fmt.Sprintf("The service instance %q was not found.", request.InstanceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Verify that we're not changing the plan (this should never happen since
	// we're setting `plan_updateable: false`, but better safe than sorry)
	if request.PlanID != nil && *request.PlanID != instance.PlanID {
		desc := fmt.Sprintf("The service plan cannot be changed from %q to %q.", instance.PlanID, *request.PlanID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the service
	service, err := b.db.DataStorePort.GetServiceDefinition(request.ServiceID)
	if err != nil {
		desc := fmt.Sprintf("Failed to get the service %q: %v", request.ServiceID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	} else if service == nil {
		desc := fmt.Sprintf("The service %q was not found.", request.ServiceID)
		return nil, newHTTPStatusCodeError(http.StatusBadRequest, "", desc)
	}

	// Get the plan
	plan := getPlan(service, instance.PlanID)
	if plan == nil {
		desc := fmt.Sprintf("The service plan %q was not found.", instance.PlanID)
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
				desc := fmt.Sprintf("The parameter %q is not updatable.", k)
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
	_, err = cfnSvc.Client.UpdateStack(&cloudformation.UpdateStackInput{
		Capabilities: aws.StringSlice([]string{cloudformation.CapabilityCapabilityNamedIam}),
		Parameters:   toCFNParams(params),
		StackName:    aws.String(instance.StackID),
		TemplateURL:  b.generateS3HTTPUrl(service.Name),
	})
	if err != nil {
		desc := fmt.Sprintf("Failed to update the CloudFormation stack %q: %v", instance.StackID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	// Update the params in the DB
	instance.Params = params
	err = b.db.DataStorePort.PutServiceInstance(*instance)
	if err != nil {
		// Try to cancel the update
		if _, err := cfnSvc.Client.CancelUpdateStack(&cloudformation.CancelUpdateStackInput{StackName: aws.String(instance.StackID)}); err != nil {
			glog.Errorf("Failed to cancel updating the CloudFormation stack %q: %v", instance.StackID, err)
			glog.Errorf("Service instance %q and CloudFormation stack %q may be out of sync!", instance.ID, instance.StackID)
		}

		desc := fmt.Sprintf("Failed to update the service instance %q: %v", instance.ID, err)
		return nil, newHTTPStatusCodeError(http.StatusInternalServerError, "", desc)
	}

	response := broker.UpdateInstanceResponse{}
	response.Async = true
	return &response, nil
}

// BindingLastOperation is not implemented, as async binding is not supported.
func (b *AwsBroker) BindingLastOperation(request *osb.BindingLastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	panic("not implemented")
}

// GetBinding is not implemented, as async binding is not supported.
func (b *AwsBroker) GetBinding(request *osb.GetBindingRequest, c *broker.RequestContext) (*broker.GetBindingResponse, error) {
	panic("not implemented")
}
