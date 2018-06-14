# AWS Service Broker - Amazon DynamoDB Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Database_AmazonDynamoDB_LARGE.png" width="108"> <p align="center">Amazon DynamoDB is a fully managed NoSQL database service that provides fast and predictable performance with seamless scalability. You can use Amazon DynamoDB to create a database table that can store and retrieve any amount of data, and serve any level of request traffic. Amazon DynamoDB automatically spreads the data and traffic for the table over a sufficient number of servers to handle the request capacity specified by the customer and the amount of data stored, while maintaining consistent and fast performance.
https://aws.amazon.com/documentation/dynamodb/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [hashrange](#param-hashrange)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [hashrange](#example-hashrange)

<a id="parameters" />

# Parameters

<a id="param-hashrange" />

## hashrange

DynamoDB Table with hash and range keys

Pricing: https://aws.amazon.com/dynamodb/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
HashAttributeName|Name of the Hash key|string
RangeAttributeName|Name of the Range key|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
HashAttributeType|AttributeType for  Hash key|S|S, N, B
RangeAttributeType|AttributeType for the  Range key|S|S, N, B
ReadCapacityUnits|Read ReadCapacity Units|5|
WriteCapacityUnits|Write Capacity Units|5|

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
TABLE_NAME|Name of the DynamoDB Table
TABLE_ARN|Arn of the DynamoDB Table

<a id="kubernetes-openshift-examples" />

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add
them as additional parameters

<a id="example-hashrange" />

## hashrange

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: dynamodb-hashrange-minimal-example
spec:
  clusterServiceClassExternalName: dh-dynamodb
  clusterServicePlanExternalName: hashrange
  parameters:
    HashAttributeName: [VALUE] # REQUIRED
    RangeAttributeName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: dynamodb-hashrange-complete-example
spec:
  clusterServiceClassExternalName: dh-dynamodb
  clusterServicePlanExternalName: hashrange
  parameters:
    HashAttributeName: [VALUE] # REQUIRED
    RangeAttributeName: [VALUE] # REQUIRED
    HashAttributeType: S # OPTIONAL
    RangeAttributeType: S # OPTIONAL
    ReadCapacityUnits: 5 # OPTIONAL
    WriteCapacityUnits: 5 # OPTIONAL
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***