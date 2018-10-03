# AWS Service Broker - Amazon EMR Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Analytics_AmazonEMR_LARGE.png" width="108"> <p align="center">Amazon EMR provides a managed Hadoop framework that makes it easy, fast, and cost-effective to process vast amounts of data across dynamically scalable Amazon EC2 instances. You can also run other popular distributed frameworks such as Apache Spark, HBase, Presto, and Flink in Amazon EMR, and interact with data in other AWS data stores such as Amazon S3 and Amazon DynamoDB.
https://aws.amazon.com/documentation/emr/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [production](#param-production)
  * [custom](#param-custom)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [production](#example-production)
  * [custom](#example-custom)

<a id="parameters" />

# Parameters

<a id="param-production" />

## production

Creates an Amazon EMR cluster optimised for production use

Pricing: https://aws.amazon.com/emr/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
KeyName|Must be an existing Keyname|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
MasterInstanceType|Instance type to be used for the master instance.|m3.xlarge|
CoreInstanceType|Instance type to be used for core instances.|m3.xlarge|
NumberOfCoreInstances|Must be a valid number|2|
ReleaseLabel|Must be a valid EMR release  version|emr-5.7.0|
EMRApplication|Please select which application will be installed on the cluster this would be either Ganglia and spark, or Ganglia and s3 backed Hbase|Spark|Spark, Hbase

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|Must be a valid VPC ID||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
EMRClusterName|Name of the cluster, if set to "Auto" a name will be auto-generated|Auto
EMRCidr|CIDR Block for EMR subnet.|Auto
<a id="param-custom" />

## custom

Creates an Amazon EMR cluster with a custom configuration

Pricing: https://aws.amazon.com/emr/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
KeyName|Must be an existing Keyname|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
EMRClusterName|Name of the cluster, if set to "Auto" a name will be auto-generated|Auto|
MasterInstanceType|Instance type to be used for the master instance.|m3.xlarge|
CoreInstanceType|Instance type to be used for core instances.|m3.xlarge|
NumberOfCoreInstances|Must be a valid number|2|
EMRCidr|CIDR Block for EMR subnet.|Auto|
ReleaseLabel|Must be a valid EMR release  version|emr-5.7.0|
EMRApplication|Please select which application will be installed on the cluster this would be either Ganglia and spark, or Ganglia and s3 backed Hbase|Spark|Spark, Hbase

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|Must be a valid VPC ID||

<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
S3_DATA_BUCKET|
EMR_ENDPOINT|
EMR_CLUSTER_ID|

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-production" />

## production

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: emr-production-minimal-example
spec:
  clusterServiceClassExternalName: emr
  clusterServicePlanExternalName: production
  parameters:
    KeyName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: emr-production-complete-example
spec:
  clusterServiceClassExternalName: emr
  clusterServicePlanExternalName: production
  parameters:
    KeyName: [VALUE] # REQUIRED
    MasterInstanceType: m3.xlarge # OPTIONAL
    CoreInstanceType: m3.xlarge # OPTIONAL
    NumberOfCoreInstances: 2 # OPTIONAL
    ReleaseLabel: emr-5.7.0 # OPTIONAL
    EMRApplication: Spark # OPTIONAL
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: emr-custom-minimal-example
spec:
  clusterServiceClassExternalName: emr
  clusterServicePlanExternalName: custom
  parameters:
    KeyName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: emr-custom-complete-example
spec:
  clusterServiceClassExternalName: emr
  clusterServicePlanExternalName: custom
  parameters:
    KeyName: [VALUE] # REQUIRED
    EMRClusterName: Auto # OPTIONAL
    MasterInstanceType: m3.xlarge # OPTIONAL
    CoreInstanceType: m3.xlarge # OPTIONAL
    NumberOfCoreInstances: 2 # OPTIONAL
    EMRCidr: Auto # OPTIONAL
    ReleaseLabel: emr-5.7.0 # OPTIONAL
    EMRApplication: Spark # OPTIONAL
```

