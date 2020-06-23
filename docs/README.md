AWS Service Broker Documentation
================================

![Architecture](/docs/images/architecture.png)
*Illustrates how application platforms can use the broker to provision and bind to AWS services.*

## Installation

* [Prerequisites](/docs/install_prereqs.md)
* [Installation on SUSE Cloud Application Platform](/docs/getting-started-cap.md)
* [Installation on OpenShift](/docs/getting-started-openshift.md)
* [Installation on Pivotal Cloud Foundry](/docs/getting-started-pcf.md)
* [Installation on Kubernetes](/docs/getting-started-k8s.md)
* [Installation on SAP Cloud Platform](/docs/getting-started-scp.md)

## Provisioning and binding services

Documentation for all of the available plans, their parameters and binding outputs are available in the
[AWS Service Broker GitHub repository](https://github.com/awslabs/aws-servicebroker/tree/master/templates)

## Configuration tasks

### Passing In AWS credentials via parameters

The **aws_access_key**, **aws_secret_key** can be passed in as parameters to the provision request.

If provided, they will be used in place of the aws service catalog process role.

These parameters will be stored in the DynamoDB backend.  Currently STS generated credentials
are not supported as there is no way to update them upon expiration via the
open service broker spec.

For example

```
# svcat provision my-instance-name \
	-n my-app \
	--class my-instance-class \
	--plan prd \
	-p VpcId=vpc-123451234512341234,aws_access_key=bacdbcadbcadbcad,aws_secret_key=abcdabcdabcdabcdabcdabcdabcdabcd
  Name:        my-instance-name
  Namespace:   my-app
  Status:
  Class:       my-instance-class
  Plan:        prd

Parameters:
  Name: my-ingress-sg-1535425552
  VpcId: vpc-123451234512341234
  aws_access_key: bacdbcadbcadbcad
  aws_secret_key: abcdabcdabcdabcdabcdabcdabcdabcd
```

### Managing Resources Via Assumed Role

The aws-service-broker has the ability to assume a role for all resources it manages.

This role can be in the same account, or a separate target account.

To setup the role, assume admin credentials in the account where the role will reside
and create the role for the aws-service-broker to assume.

```
service_broker_account_id=100000000001 # role where the service broker will run, will be the same as the target if in single account

aws cloudformation create-stack \
    --stack-name AwsServiceBrokerWorkerRole \
    --template-body file://setup/aws-service-broker-worker.json \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters ParameterKey=ServiceBrokerAccountId,ParameterValue=$service_broker_account_id
```

To do this you must ensure that the role the **aws-service-broker** is running allows it to assume the target role.

Get the ARN:

```
aws cloudformation describe-stacks \
        --stack-name AwsServiceBrokerWorkerRole | jq -r .Stacks[0].Outputs[0].OutputValue
```

Ensure the service broker role has the below permissions:

```json
{
  "Action": "sts:AssumeRole",
  "Resource": "arn:aws:iam::123456654321:role/aws-service-broker-worker",
  "Effect": "Allow"
}
```

Provide **target_account_id** and **target_role_name** as parameters to the provision command
to tell the service broker to assume the role in another account to provision.

```
svcat provision my-ingress-api-gw \
    -n my-app \
    --class my-class \
    --plan prd \
    -p VpcId=vpc-1234567887654321,target_account_id=123456654321,target_role_name=aws-service-broker-worker
````

### Overriding the default AWS region

The **region** can be passed in as a parameter to the provision request.

If provided, it will be used in place of the aws service catalog process region.

### Parameter Overrides

> **NOTE:** Current releases of the Service Broker have the DynamoDB mechanism disabled, please use the Environment Variable approach to prescribing overrides

The broker can override parameter values using override records in the metadata DynamoDB table, or by providing environment variables in the broker execution environment.

The broker provides a hierarchy of parameter overrides to prescribe values for common parameters like AWS credentials, region,
VPC ID or any other parameter in a service plan.

An override can be broker wide, or only apply to a particular org/cluster, space/namespace, or ServiceClass.

#### Environment Variables
The following structure is used for override Environment Variables:
```
PARAM_OVERRIDE_<BROKER_ID>_<ORG_GUID/CLUSTER_ID>_<SPACE_GUID/NAMESPACE>_<SERVICE>_<PARAMETER>=<value>
```

`<ORG_GUID/CLUSTER_ID>`, `<SPACE_GUID/NAMESPACE>`, and `<SERVICE>` can all have the literal value `all`.

#### DynamoDB

The structure of an override record is:

```json
{
    "id": "<UUID>",
    "userid": "<UUID>",
    "parameter_name": "<PARAMETER_NAME>",
    "parameter_value": "<PARAMETER_VALUE>",
    "service_class": "<SERVICECLASS_NAME>",
    "org_guid": "<CLOUDFOUNDRY_ORG_GUID>",
    "space_guid": "<CLOUDFOUNDRY_SPACE_GUID>",
    "cluster_id": "<KUBERNETES_CLUSTER_ID>",
    "namespace": "<KUBERNETES_NAMESPACE_ID>"
}
```

> Notes:
> * `id`, `userid`, `parameter_name` and `parameter_value` are required.
> * `org_guid` and `space_guid` are Cloud Foundry specific, and cannot be combined with `cluster_id` and `namespace` (Kubernetes specific)
> * If a parameter is overridden globally (none of the optional fields are provided) and the `-prescribeOverrides` flag is passed, it will be removed from the available parameters presented by the application platform's UI
> * cluster_id for kubernetes is [generated by the service catalog](https://github.com/kubernetes-incubator/service-catalog/blob/acf976260e505bedb10b7c8f18efc69833714ecc/pkg/controller/controller.go#L1317), and will change if the service catalog is removed and reinstalled.

The order of precedence for parameter values is:

1. Plan default
2. User provided
3. Global Overrides
4. ServiceClass overrides
5. Org/Cluster overrides
6. Org/Cluster + ServiceClass overrides
7. Space/Namespace overrides
8. Space/Namespace + ServiceClass overrides
9. Org/Cluster + Space/Namespace overrides
10. Org/Cluster + Space/Namespace + ServiceClass overrides

#### Examples

> Note: You need the ossp-uuid and aws-cli command line tools to run these examples

**Set a global override to provision into us-west-2 region:**

**DynamoDB:**

```bash
ACCOUNT_ID=123456789012          # Account ID for the AWS account that the broker user/role is in
BROKER_ID=aws-service-broker     # brokerId provided as an argument when launching the broker, if not specified it defaults to aws-service-broker
DYNAMODB_TABLE=awssb             # name of broker metadata table
DYNAMODB_REGION=us-east-1        # region that the dynamo table is in
cat <<EOF > "./override.json"
{
        "id": { "S": "$(uuid)" },
        "userid": { "S": "$(uuid -v 5 00000000-0000-0000-0000-000000000000 ${ACCOUNT_ID}${BROKER_ID})" },
        "parameter_name": { "S": "region" },
        "parameter_value": { "S": "us-west-2" }
}
EOF
aws dynamodb put-item --table-name ${DYNAMODB_TABLE} --region ${DYNAMODB_REGION} --item file://override.json
```

**Environment Variable:**

add an environment variable as follows (assumes the broker has been configured with a BROKER_ID of `awsservicebroker`:

```
PARAM_OVERRIDE_awsservicebroker_all_all_all_region=us-west-2
```

**Set `myns` namespace to provision into us-west-2 region:**

**DynamoDB:**
```bash
ACCOUNT_ID=123456789012          # Account ID for the AWS account that the broker user/role is in
BROKER_ID=aws-service-broker     # brokerId provided as an argument when launching the broker, if not specified it defaults to aws-service-broker
DYNAMODB_TABLE=awssb             # name of broker metadata table
DYNAMODB_REGION=us-east-1        # region that the dynamo table is in
CLUSTER_ID=$(kubectl get cm cluster-info -n catalog -o jsonpath='{$.data.id}') # Ensure your kubectl is set to the desired cluster
NAMESPACE=myns
cat <<EOF > "./override.json"
{
        "id": { "S": "$(uuid)" },
        "userid": { "S": "$(uuid -v 5 00000000-0000-0000-0000-000000000000 ${ACCOUNT_ID}${BROKER_ID})" },
        "parameter_name": { "S": "region" },
        "parameter_value": { "S": "us-west-2" },
        "cluster_id": { "S": "${CLUSTER_ID}" },
        "namespace": { "S": "${NAMESPACE}" }
}
EOF
aws dynamodb put-item --table-name ${DYNAMODB_TABLE} --region ${DYNAMODB_REGION} --item file://override.json
```
**Environment Variable:**

add an environment variable replacing the sections between `<>` with appropriate values (assumes the broker has been configured with a BROKER_ID of `awsservicebroker`:

```
PARAM_OVERRIDE_awsservicebroker_<ORG_GUID/CLUSTER_ID>_<SPACE_GUID/NAMESPACE>_all_region=us-west-2
```



### Custom Catalog

You can configure the broker to point to your own S3 bucket (which can be private or public) containing
CloudFormation templates and ServiceClass specs. The bucket, prefix and AWS region that the broker scans for ServiceClasses is configured using the
`-s3Bucket`, `-s3Key` and `-s3Region` commandline switches.

* [Example -spec.yaml file](/docs/examples/example-main.yaml)

#### Generating unique credentials for each bind request

It is possible to define a CloudFormation template that defines an AWS Lambda function to be run whenever the AWS Service Broker recieves a bind or unbind request.  This in turn allows for the bindings to be defined uniquely rather than returning the current state of the CloudFormation outputs for each new binding. In order to use this functionality you must define three things in your template:

1. You must place the key/value pair `BindViaLambda: true` in your `Metadata` section.
2. You must define a lambda function as a custom resource within you template.
3. You must output the name or Arn of your lambda function, in the `Outputs` section of your template, with the key `BindLambda`. 

* [Example -spec.yaml file with Lambda generated bindings](/docs/examples/example-with-lambda-bindings-main.yaml)


### Template Metadata Generator

A tool that examines the parameters of the JSON Cloudformation template and automatically creates the Service Broker Metadata (AWS::ServiceBroker::Specification) for the template and outputs in JSON and YAML format. It will include all parameters in the template into the metadata.

https://github.com/ndrest-amzn/ServiceBrokerMetaGen

### Template Documentation Generator

A tool tha automatically creates the Service Broker doc Readme.md from the Cloudformation Template provided. It will include all parameters in the template into the Readme.md.

https://github.com/ndrest-amzn/ServiceBrokerDocGen
