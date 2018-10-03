# AWS Service Broker - Amazon SNS Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Messaging_AmazonSNS_LARGE.png" width="108"> <p align="center">Amazon Simple Notification Service (Amazon SNS) is a web service that enables applications, end-users, and devices to instantly send and receive notifications from the cloud.
https://aws.amazon.com/documentation/sns/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [topicwithsub](#param-topicwithsub)
  * [topic](#param-topic)
  * [subscription](#param-subscription)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [topicwithsub](#example-topicwithsub)
  * [topic](#example-topic)
  * [subscription](#example-subscription)

<a id="parameters" />

# Parameters

<a id="param-topicwithsub" />

## topicwithsub



Pricing: https://aws.amazon.com/sns/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
SubscriptionEndPoint|The endpoint that receives notifications from the Amazon SNS topic. If left blank no subscription will be added to the topic. The endpoint value depends on the protocol that you specify. This could be a URL, ARN or SMS-capable telephone number.|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
SubscriptionProtocol|The subscription's protocol. Examples: "http", "https", "email", "email-json", "sms", "sqs", "application", "lambda".|sqs|, http, https, email, email-json, sms, sqs, application, lambda
SubscriptionNumRetries|Number of retries in the backoff phase|3|
SubscriptionMinDelayTarget|Defines the delay associated with the first retry attempt in the backoff phase|20|
SubscriptionMaxDelayTarget|Defines the delay associated with the final retry attempt in the backoff phase|20|

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
ExistingTopicArn|If not creating a topic, define the arn for an existing topic|
CreateTopic|Should we create a topic or not ?|Yes
<a id="param-topic" />

## topic



Pricing: https://aws.amazon.com/sns/pricing/


### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
SubscriptionNumRetries|Number of retries in the backoff phase|3|
SubscriptionMinDelayTarget|Defines the delay associated with the first retry attempt in the backoff phase|20|
SubscriptionMaxDelayTarget|Defines the delay associated with the final retry attempt in the backoff phase|20|

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
ExistingTopicArn|If not creating a topic, define the arn for an existing topic|
CreateTopic|Should we create a topic or not ?|Yes
SubscriptionProtocol|The subscription's protocol. Examples: "http", "https", "email", "email-json", "sms", "sqs", "application", "lambda".|
SubscriptionEndPoint|The endpoint that receives notifications from the Amazon SNS topic. If left blank no subscription will be added to the topic. The endpoint value depends on the protocol that you specify. This could be a URL, ARN or SMS-capable telephone number.|
<a id="param-subscription" />

## subscription



Pricing: https://aws.amazon.com/sns/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
SubscriptionEndPoint|The endpoint that receives notifications from the Amazon SNS topic. If left blank no subscription will be added to the topic. The endpoint value depends on the protocol that you specify. This could be a URL, ARN or SMS-capable telephone number.|string
ExistingTopicArn|If not creating a topic, define the arn for an existing topic|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
SubscriptionProtocol|The subscription's protocol. Examples: "http", "https", "email", "email-json", "sms", "sqs", "application", "lambda".|sqs|, http, https, email, email-json, sms, sqs, application, lambda
SubscriptionNumRetries|Number of retries in the backoff phase|3|
SubscriptionMinDelayTarget|Defines the delay associated with the first retry attempt in the backoff phase|20|
SubscriptionMaxDelayTarget|Defines the delay associated with the final retry attempt in the backoff phase|20|

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
CreateTopic|Should we create a topic or not ?|No
<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
TOPIC_ARN|ARN of SNS Topic

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-topicwithsub" />

## topicwithsub

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-topicwithsub-minimal-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: topicwithsub
  parameters:
    SubscriptionEndPoint: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-topicwithsub-complete-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: topicwithsub
  parameters:
    SubscriptionEndPoint: [VALUE] # REQUIRED
    SubscriptionProtocol: sqs # OPTIONAL
    SubscriptionNumRetries: 3 # OPTIONAL
    SubscriptionMinDelayTarget: 20 # OPTIONAL
    SubscriptionMaxDelayTarget: 20 # OPTIONAL
```
<a id="example-topic" />

## topic

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-topic-minimal-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: topic
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-topic-complete-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: topic
  parameters:
    SubscriptionNumRetries: 3 # OPTIONAL
    SubscriptionMinDelayTarget: 20 # OPTIONAL
    SubscriptionMaxDelayTarget: 20 # OPTIONAL
```
<a id="example-subscription" />

## subscription

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-subscription-minimal-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: subscription
  parameters:
    SubscriptionEndPoint: [VALUE] # REQUIRED
    ExistingTopicArn: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: sns-subscription-complete-example
spec:
  clusterServiceClassExternalName: sns
  clusterServicePlanExternalName: subscription
  parameters:
    SubscriptionEndPoint: [VALUE] # REQUIRED
    ExistingTopicArn: [VALUE] # REQUIRED
    SubscriptionProtocol: sqs # OPTIONAL
    SubscriptionNumRetries: 3 # OPTIONAL
    SubscriptionMinDelayTarget: 20 # OPTIONAL
    SubscriptionMaxDelayTarget: 20 # OPTIONAL
```

