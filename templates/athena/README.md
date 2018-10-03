# AWS Service Broker - Amazon Athena. Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonAthena_LARGE.png" width="108"> <p align="center">Amazon Athena is an interactive query service that makes it easy to analyze data in Amazon S3 using standard SQL. Athena is serverless, so there is no infrastructure to manage, and you pay only for the queries that you run.
https://aws.amazon.com/documentation/athena/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [default](#param-default)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [default](#example-default)

<a id="parameters" />

# Parameters

<a id="param-default" />

## default

Athena table using an existing S3 source

Pricing: https://aws.amazon.com/athena/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
S3Source|The S3 location for the source data.|string
TableColumns|The columns and their types in the format: (col_name data_type [COMMENT col_comment] [, ...] )|string
RowFormat|The row format of the source data.|DELIMITED, SERDE
SerdeName|SERDE Name, only applicable if "Row Format" is set to SERDE.|string
SerdeProperties|SERDE Properties in the format ("property_name" = "property_value", "property_name" = "property_value" [, ...] ). Only applicable if "Row Format" is set to SERDE.|string
AthenaDBName|Athena Database name, will be created if it does not exist|string
TableName|Athena table name|string


### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2

<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-default" />

## default

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: athena-default-minimal-example
spec:
  clusterServiceClassExternalName: athena
  clusterServicePlanExternalName: default
  parameters:
    S3Source: [VALUE] # REQUIRED
    TableColumns: [VALUE] # REQUIRED
    RowFormat: [VALUE] # REQUIRED
    SerdeName: [VALUE] # REQUIRED
    SerdeProperties: [VALUE] # REQUIRED
    AthenaDBName: [VALUE] # REQUIRED
    TableName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: athena-default-complete-example
spec:
  clusterServiceClassExternalName: athena
  clusterServicePlanExternalName: default
  parameters:
    S3Source: [VALUE] # REQUIRED
    TableColumns: [VALUE] # REQUIRED
    RowFormat: [VALUE] # REQUIRED
    SerdeName: [VALUE] # REQUIRED
    SerdeProperties: [VALUE] # REQUIRED
    AthenaDBName: [VALUE] # REQUIRED
    TableName: [VALUE] # REQUIRED
```

