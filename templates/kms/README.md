# AWS Service Broker - KMS Key Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/SecurityIdentityCompliance_AWSKMS_LARGE.png" width="108"> <p align="center">AWS Key Management Service (KMS) is a managed service that makes it easy for you to create and control the encryption keys used to encrypt your data, and uses FIPS 140-2 validated hardware security modules to protect the security of your keys.
https://aws.amazon.com/documentation/kms/</p>

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

Generates a KMS key

Pricing: https://aws.amazon.com/kms/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
KeyAdministratorRoleArn|To add an additional administrative role, specify the ARN here. By default the root user and the CloudFormation Stack role is granted administrative access to the key. Admins can update, revoke, delete the key, but cannot use it to encrypt or decrypt.|string
CloudformationRoleArn|Role used to launch this stack, this is typically configured as an AWS Service Broker Secret.|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
EnableKeyRotation|AWS KMS generates new cryptographic material for the CMK every year. AWS KMS also saves the CMK's older cryptographic material so it can be used to decrypt data that it encrypted.|true|true, false

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
KMS_KEY_ID|Id of the KMS key
KMS_KEY_ARN|Arn of the KMS key

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
  name: kms-default-minimal-example
spec:
  clusterServiceClassExternalName: kms
  clusterServicePlanExternalName: default
  parameters:
    KeyAdministratorRoleArn: [VALUE] # REQUIRED
    CloudformationRoleArn: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: kms-default-complete-example
spec:
  clusterServiceClassExternalName: kms
  clusterServicePlanExternalName: default
  parameters:
    KeyAdministratorRoleArn: [VALUE] # REQUIRED
    CloudformationRoleArn: [VALUE] # REQUIRED
    EnableKeyRotation: true # OPTIONAL
```

