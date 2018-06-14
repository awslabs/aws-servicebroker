# AWS Service Broker - Amazon Lex Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonLex_LARGE.png" width="108"> <p align="center">Amazon Lex is a service for building conversational interfaces into any application using voice and text. Amazon Lex provides the advanced deep learning functionalities of automatic speech recognition (ASR) for converting speech to text, and natural language understanding (NLU) to recognize the intent of the text, to enable you to build applications with highly engaging user experiences and lifelike conversational interactions.
https://aws.amazon.com/documentation/lex/</p>

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

Creates an Amazon Lex bot

Pricing: https://aws.amazon.com/lex/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
LexS3Bucket|Name of the S3 bucket containing the bot and (optionally) intent/custom slot type json documents|string
IntentsKey|S3 key to a json document containing a list of Lex intents to create for example: [{"name": "intent1", ...}, {"name": "intent2", ...}] . For more information on the intent structure, see https://docs.aws.amazon.com/lex/latest/dg/API_PutIntent.html. If no intents are required leave this field empty|string
CustomSlotTypesKey|S3 key to a json document containing a list of Lex custom slot types to create for example: [{"name": "slot1", ...}, {"name": "slot2", ...}] . For more information on the slot type structure, see https://docs.aws.amazon.com/lex/latest/dg/API_PutSlotType.html. If no custom slot types are required leave this field empty|string
BotKey|S3 key to a json document containing a Lex bot definition to create. For more information on the bot structure, see https://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html.|string


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
  name: lex-default-minimal-example
spec:
  clusterServiceClassExternalName: dh-lex
  clusterServicePlanExternalName: default
  parameters:
    LexS3Bucket: [VALUE] # REQUIRED
    IntentsKey: [VALUE] # REQUIRED
    CustomSlotTypesKey: [VALUE] # REQUIRED
    BotKey: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: lex-default-complete-example
spec:
  clusterServiceClassExternalName: dh-lex
  clusterServicePlanExternalName: default
  parameters:
    LexS3Bucket: [VALUE] # REQUIRED
    IntentsKey: [VALUE] # REQUIRED
    CustomSlotTypesKey: [VALUE] # REQUIRED
    BotKey: [VALUE] # REQUIRED
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***