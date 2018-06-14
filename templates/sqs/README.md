# AWS Servicebroker - Amazon SQS Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Messaging_AmazonSQS_LARGE.png" width="108"> <p align="center">Amazon Simple Queue Service (Amazon SQS) is a fully managed message queuing service that makes it easy to decouple and scale microservices, distributed systems, and serverless applications. Amazon SQS moves data between distributed application components and helps you decouple these components."
https://aws.amazon.com/documentation/sqs/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [standard](#param-standard)
  * [fifo](#param-fifo)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [standard](#example-standard)
  * [fifo](#example-fifo)

<a id="parameters" />

# Parameters

<a id="param-standard" />

## standard

Managed Standard SQS Queue

Pricing: https://aws.amazon.com/sqs/pricing/


### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
DelaySeconds|The Id of the AMI you wish to launch the instance from.|5|
MaximumMessageSize|The limit of how many bytes that a message can contain before Amazon SQS rejects it, 1024 bytes (1 KiB) to 262144 bytes (256 KiB)|262144|
MessageRetentionPeriod|The number of seconds that Amazon SQS retains a message. You can specify an integer value from 60 seconds (1 minute) to 1209600 seconds (14 days).|345600|
ReceiveMessageWaitTimeSeconds|Specifies the duration, in seconds, that the ReceiveMessage action call waits until a message is in the queue in order to include it in the response, as opposed to returning an empty response if a message is not yet available. 1 to 20|0|
UsedeadletterQueue|A dead-letter queue is a queue that other (source) queues can target for messages that can't be processed (consumed) successfully. You can set aside and isolate these messages in the dead-letter queue to determine why their processing doesn't succeed.|false|true, false
VisibilityTimeout|This should be longer than the time it would take to process and delete a message, this should not exceed 12 hours.|5|

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are
configured with a broker secret, see getting started guides for [OpenShift](/docs/getting-started-openshift.md) or
[Kubernetes](/docs/getting-started-k8s.md) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
aws_access_key|AWS Access Key to authenticate to AWS with.||
aws_secret_key|AWS Secret Key to authenticate to AWS with.||
aws_cloudformation_role_arn|IAM role ARN for use as Cloudformation Stack Role.||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
SBArtifactS3Bucket|Name of the S3 bucket containing the AWS Service Broker Assets|awsservicebroker|
SBArtifactS3KeyPrefix|Name of the S3 key prefix containing the AWS Service Broker Assets, leave empty if assets are in the root of the bucket||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
FifoQueue|If true queue will be FIFO|false
ContentBasedDeduplication|specifies whether to enable content-based deduplication, only applies to FIFO queues|false
<a id="param-fifo" />

## fifo

Managed FIFO SQS Queue

Pricing: https://aws.amazon.com/sqs/pricing/


### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
ContentBasedDeduplication|specifies whether to enable content-based deduplication, only applies to FIFO queues|true|true, false
DelaySeconds|The Id of the AMI you wish to launch the instance from.|5|
MaximumMessageSize|The limit of how many bytes that a message can contain before Amazon SQS rejects it, 1024 bytes (1 KiB) to 262144 bytes (256 KiB)|262144|
MessageRetentionPeriod|The number of seconds that Amazon SQS retains a message. You can specify an integer value from 60 seconds (1 minute) to 1209600 seconds (14 days).|345600|
ReceiveMessageWaitTimeSeconds|Specifies the duration, in seconds, that the ReceiveMessage action call waits until a message is in the queue in order to include it in the response, as opposed to returning an empty response if a message is not yet available. 1 to 20|0|
UsedeadletterQueue|A dead-letter queue is a queue that other (source) queues can target for messages that can't be processed (consumed) successfully. You can set aside and isolate these messages in the dead-letter queue to determine why their processing doesn't succeed.|false|true, false
VisibilityTimeout|This should be longer than the time it would take to process and delete a message, this should not exceed 12 hours.|5|

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are
configured with a broker secret, see getting started guides for [OpenShift](/docs/getting-started-openshift.md) or
[Kubernetes](/docs/getting-started-k8s.md) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
aws_access_key|AWS Access Key to authenticate to AWS with.||
aws_secret_key|AWS Secret Key to authenticate to AWS with.||
aws_cloudformation_role_arn|IAM role ARN for use as Cloudformation Stack Role.||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
SBArtifactS3Bucket|Name of the S3 bucket containing the AWS Service Broker Assets|awsservicebroker|
SBArtifactS3KeyPrefix|Name of the S3 key prefix containing the AWS Service Broker Assets, leave empty if assets are in the root of the bucket||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
FifoQueue|If true queue will be FIFO|true
<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
QUEUE_URL|URL of newly created SQS Queue
QUEUE_ARN|ARN of newly created SQS Queue
QUEUE_NAME|Name newly created SQS Queue
DEAD_LETTER_QUEUE_URL|URL of newly created SQS Queue
DEAD_LETTER_QUEUE_ARN|ARN of newly created SQS Queue
DEAD_LETTER_QUEUE_NAME|Name newly created SQS Queue

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-standard" />

## standard

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sqs-standard-minimal-example
spec:
  clusterServiceClassExternalName: dh-sqs
  clusterServicePlanExternalName: standard
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sqs-standard-complete-example
spec:
  clusterServiceClassExternalName: dh-sqs
  clusterServicePlanExternalName: standard
  parameters:
    DelaySeconds: 5 # OPTIONAL
    MaximumMessageSize: 262144 # OPTIONAL
    MessageRetentionPeriod: 345600 # OPTIONAL
    ReceiveMessageWaitTimeSeconds: 0 # OPTIONAL
    UsedeadletterQueue: false # OPTIONAL
    VisibilityTimeout: 5 # OPTIONAL
```
<a id="example-fifo" />

## fifo

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sqs-fifo-minimal-example
spec:
  clusterServiceClassExternalName: dh-sqs
  clusterServicePlanExternalName: fifo
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sqs-fifo-complete-example
spec:
  clusterServiceClassExternalName: dh-sqs
  clusterServicePlanExternalName: fifo
  parameters:
    ContentBasedDeduplication: true # OPTIONAL
    DelaySeconds: 5 # OPTIONAL
    MaximumMessageSize: 262144 # OPTIONAL
    MessageRetentionPeriod: 345600 # OPTIONAL
    ReceiveMessageWaitTimeSeconds: 0 # OPTIONAL
    UsedeadletterQueue: false # OPTIONAL
    VisibilityTimeout: 5 # OPTIONAL
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***