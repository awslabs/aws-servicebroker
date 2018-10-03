# AWS Service Broker - Amazon Route 53 Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/NetworkingContentDelivery_AmazonRoute53_LARGE.png" width="108"> <p align="center">Amazon Route 53 is a highly available and scalable Domain Name System (DNS) web service.
https://aws.amazon.com/documentation/route53/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [hostedzone](#param-hostedzone)
  * [recordset](#param-recordset)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [hostedzone](#example-hostedzone)
  * [recordset](#example-recordset)

<a id="parameters" />

# Parameters

<a id="param-hostedzone" />

## hostedzone



Pricing: https://aws.amazon.com/route53/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
NewHostedZoneName|Name of the hosted zone|string


### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
HostedZoneName|Name of the hosted zone which the records are to be created in|
HostedZoneId|Id of the hosted zone which the records are to be created in|
TimeToLive|How long the resolved record should be cached by resolvers|360
Type|Type of record|A
RecordName|Name of the record|
ResourceRecord|Value of the record|
AliasTarget|Alias resource record sets only: Information about the domain to which you are redirecting traffic.|
<a id="param-recordset" />

## recordset



Pricing: https://aws.amazon.com/route53/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
HostedZoneName|Name of the hosted zone which the records are to be created in|string
HostedZoneId|Id of the hosted zone which the records are to be created in|string
RecordName|Name of the record|string
ResourceRecord|Value of the record|string
AliasTarget|Alias resource record sets only: Information about the domain to which you are redirecting traffic.|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
TimeToLive|How long the resolved record should be cached by resolvers|360|
Type|Type of record|A|A, AAAA, CAA, CNAME, MX, NS, PTR, SOA, SPF, SRV, TXT

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
NewHostedZoneName|Name of the hosted zone|
<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
RECORD_NAME|
HOSTED_ZONE_ID|

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-hostedzone" />

## hostedzone

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: route53-hostedzone-minimal-example
spec:
  clusterServiceClassExternalName: route53
  clusterServicePlanExternalName: hostedzone
  parameters:
    NewHostedZoneName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: route53-hostedzone-complete-example
spec:
  clusterServiceClassExternalName: route53
  clusterServicePlanExternalName: hostedzone
  parameters:
    NewHostedZoneName: [VALUE] # REQUIRED
```
<a id="example-recordset" />

## recordset

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: route53-recordset-minimal-example
spec:
  clusterServiceClassExternalName: route53
  clusterServicePlanExternalName: recordset
  parameters:
    HostedZoneName: [VALUE] # REQUIRED
    HostedZoneId: [VALUE] # REQUIRED
    RecordName: [VALUE] # REQUIRED
    ResourceRecord: [VALUE] # REQUIRED
    AliasTarget: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: route53-recordset-complete-example
spec:
  clusterServiceClassExternalName: route53
  clusterServicePlanExternalName: recordset
  parameters:
    HostedZoneName: [VALUE] # REQUIRED
    HostedZoneId: [VALUE] # REQUIRED
    RecordName: [VALUE] # REQUIRED
    ResourceRecord: [VALUE] # REQUIRED
    AliasTarget: [VALUE] # REQUIRED
    TimeToLive: 360 # OPTIONAL
    Type: A # OPTIONAL
```

