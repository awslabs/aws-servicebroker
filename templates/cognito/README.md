# AWS Service Broker - Amazon Cognito

<img align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src=https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png width="108"><p align="center">Amazon Cognito lets you add user sign-up, sign-in, and access control to your web and mobile apps quickly and easily. Amazon Cognito scales to millions of users and supports sign-in with social identity providers, such as Facebook, Google, and Amazon, and enterprise identity providers via SAML 2.0.</p>&nbsp;

Table of contents
=================

* [Parameters](#parameters)
  * [production](#param-production)
  * [dev](#param-dev)
  * [custom](#param-custom)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [production](#example-production)
  * [dev](#example-dev)
  * [custom](#example-custom)

<a id="parameters"></a>

# Parameters

<a id = "param-production"></a>

Creates an Amazon Cognito optimised for production use.  
Pricing: https://aws.amazon.com/cognito/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create RDS instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto

<a id = "param-dev"></a>

Creates an Amazon Cognito optimised for dev/test use.  
Pricing: https://aws.amazon.com/cognito/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create RDS instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto

<a id = "param-custom"></a>

Creates an Amazon Cognito with custom configuration.  
Pricing: https://aws.amazon.com/cognito/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
AuthName|Name for Cognito Resources|Auto|

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create RDS instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

<a id="bind-credentials"></a>

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add them as additional parameters

<a id ="example-production"></a>

## production

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-production-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: production
  parameters: 
    AuthName: Auto
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-production-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: production
  parameters: 
    AuthName: Auto
```


<a id="example-dev"></a>

## dev

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-dev-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: dev
  parameters: 
    AuthName: Auto
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-dev-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: dev
  parameters: 
    AuthName: Auto
```


<a id = "example-custom"></a>

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-custom-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: custom
  parameters: 
    AuthName: Auto
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: cognito-custom-complete-example
spec: 
  clusterServiceClassExternalName: cognito
  clusterServicePlanExternalName: custom
  parameters: 
    AuthName: Auto
```


