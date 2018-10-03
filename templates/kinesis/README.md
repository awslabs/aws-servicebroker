# AWS Service Broker - Amazon Kinesis Data Stream Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonKinesis_LARGE.png" width="108"> <p align="center">Amazon Kinesis Data Streams enables you to build custom applications that process or analyze streaming data for specialized needs. Kinesis Data Streams can continuously capture and store terabytes of data per hour from hundreds of thousands of sources such as website clickstreams, financial transactions, social media feeds, IT logs, and location-tracking events.
https://aws.amazon.com/documentation/kinesis/</p>

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

Creates a Kinesis stream

Pricing: https://aws.amazon.com/kinesis/pricing/


### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
RetentionPeriodHours|The number of hours for the data records that are stored in shards to remain accessible. The default value is 24. For more information about the stream retention period, see Changing the Data Retention Period in the Amazon Kinesis Developer Guide.|168|
ShardCount|The number of shards that the stream uses. For greater provisioned throughput, increase the number of shards.|3|
StreamEncrypted|Indicates whether the Kinesis Stream is encrypted.|True|True, False

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are
configured with a broker secret, see getting started guides for [OpenShift](/docs/getting-started-openshift.md) or
[Kubernetes](/docs/getting-started-k8s.md) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
SBArtifactS3Bucket|Name of the S3 bucket containing the AWS Service Broker Assets|awsservicebroker|
SBArtifactS3KeyPrefix|Name of the S3 key prefix containing the AWS Service Broker Assets, leave empty if assets are in the root of the bucket||

<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
STREAM_NAME|The stream name or physical ID
STREAM_ARN|The ARN of the stream

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
  name: kinesis-default-minimal-example
spec:
  clusterServiceClassExternalName: kinesis
  clusterServicePlanExternalName: default
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: kinesis-default-complete-example
spec:
  clusterServiceClassExternalName: kinesis
  clusterServicePlanExternalName: default
  parameters:
    RetentionPeriodHours: 168 # OPTIONAL
    ShardCount: 3 # OPTIONAL
    StreamEncrypted: True # OPTIONAL
```

