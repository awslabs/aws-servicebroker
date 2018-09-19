# AWS Service Broker - Amazon S3 Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Storage_AmazonS3_LARGE.png" width="108"> <p align="center">Amazon Simple Storage Service (Amazon S3) is storage for the Internet. You can use Amazon S3 to store and retrieve any amount of data at any time, from anywhere on the web. You can accomplish these tasks using the simple and intuitive web interface of the AWS Management Console.
https://aws.amazon.com/documentation/s3/'</p>

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

Amazon Simple Storage Service (Amazon S3) is storage for the Internet. You can use Amazon S3 to store and retrieve any amount of data at any time, from anywhere on the web. You can accomplish these tasks using the simple and intuitive web interface of the AWS Management Console.

Pricing: https://aws.amazon.com/s3/pricing/



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
BucketName|Must contain only lowercase letters, numbers, periods (.), and hyphens. If set to Auto, a bucket name will be generated (-),Cannot end in numbers|Auto
LoggingPrefix|Must contain only lowercase letters, numbers, periods (.), and hyphens (-),Cannot end in numbers|S3AccessLogs
EnableGlacierLifeCycle|enable archiving to Glacier Storage|False
GlacierLifeCycleTransitionInDays|Define how many days objects should exist before being moved to Glacier|30
LifeCyclePrefix|Must contain only lowercase letters, numbers, periods (.), and hyphens (-),Cannot end in numbers|Archive
EnableVersioning|enable versioning|True
BucketAccessControl|define if the bucket can be accessed from public or private locations|Private
EnableLogging|enable or discable S3 logging|True
PreventDeletion|With the PreventDeletion attribute you can preserve a resource when its stack is deleted|True
<a id="param-custom" />

## custom

Amazon Simple Storage Service (Amazon S3) is storage for the Internet. You can use Amazon S3 to store and retrieve any amount of data at any time, from anywhere on the web. You can accomplish these tasks using the simple and intuitive web interface of the AWS Management Console.

Pricing: https://aws.amazon.com/s3/pricing/


### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
BucketName|Must contain only lowercase letters, numbers, periods (.), and hyphens. If set to Auto, a bucket name will be generated (-),Cannot end in numbers|Auto|
LoggingPrefix|Must contain only lowercase letters, numbers, periods (.), and hyphens (-),Cannot end in numbers|Archive|
EnableLogging|enable or discable S3 logging|True|True, False
EnableGlacierLifeCycle|enable archiving to Glacier Storage|False|True, False
GlacierLifeCycleTransitionInDays|Define how many days objects should exist before being moved to Glacier|0|
EnableVersioning|enable versioning|False|True, False
LifeCyclePrefix|Must contain only lowercase letters, numbers, periods (.), and hyphens (-),Cannot end in numbers|Archive|
BucketAccessControl|define if the bucket can be accessed from public or private locations|Private|Private, PublicRead, PublicReadWrite, AuthenticatedRead, LogDeliveryWrite, BucketOwnerRead, BucketOwnerFullControl, AwsExecRead
PreventDeletion|With the PreventDeletion attribute you can preserve a resource when its stack is deleted|True|True, False

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
BUCKET_NAME|Name of the sample Amazon S3 bucket.
BUCKET_ARN|Name of the Amazon S3 bucket
LOGGING_BUCKET_NAME|Name of the logging bucket.

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
  name: s3-production-minimal-example
spec:
  clusterServiceClassExternalName: dh-s3
  clusterServicePlanExternalName: production
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: s3-production-complete-example
spec:
  clusterServiceClassExternalName: dh-s3
  clusterServicePlanExternalName: production
  parameters:
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: s3-custom-minimal-example
spec:
  clusterServiceClassExternalName: dh-s3
  clusterServicePlanExternalName: custom
  parameters:
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: s3-custom-complete-example
spec:
  clusterServiceClassExternalName: dh-s3
  clusterServicePlanExternalName: custom
  parameters:
    BucketName: Auto # OPTIONAL
    LoggingPrefix: Archive # OPTIONAL
    EnableLogging: True # OPTIONAL
    EnableGlacierLifeCycle: False # OPTIONAL
    GlacierLifeCycleTransitionInDays: 0 # OPTIONAL
    EnableVersioning: False # OPTIONAL
    LifeCyclePrefix: Archive # OPTIONAL
    BucketAccessControl: Private # OPTIONAL
    PreventDeletion: True # OPTIONAL
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***