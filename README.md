# AWS Service Broker

## Usage

### Prerequisistes

#### Application platform

Aims to be compatible with Pivotal Cloud Foundry 1.12+, OpenShift 3.9+ and kubernetes 1.8+ (requires k8s-servicecatalog)

A minikube env with servicecatalog, svcat and broker image can be started for testing by running 

```bash
make functional-test
```

Assuming no errors, this should shell you into a container with a running minikube, service-catalog, svcat, 
aws-servicebroker binary and pre-configured kubectl, type `exit` to destroy the container

#### S3 Bucket

S3 bucket containing CloudFormation templates and required osb schema files, this can be private, or a public bucket can 
also be used, the `awsservicebrokeralpha` bucket (s3Key prefix `pcf/templates`) contains initial builds of the aws service classes formatted for this broker and can be used for testing. 

#### DynamoDB Table

Table can be created with the following aws cli command:

```bash
aws dynamodb create-table --attribute-definitions \
AttributeName=id,AttributeType=S AttributeName=userid,AttributeType=S \
AttributeName=type,AttributeType=S --key-schema AttributeName=id,KeyType=HASH \
AttributeName=userid,KeyType=RANGE --global-secondary-indexes \
'IndexName=type-userid-index,KeySchema=[{AttributeName=type,KeyType=HASH},{AttributeName=userid,KeyType=RANGE}],Projection={ProjectionType=INCLUDE,NonKeyAttributes=[id,userid,type,locked]},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}' \
--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
--region us-east-1 --table-name awssb
```

You can customize the table name as needed and pass in your table name using –tableName

#### IAM 
 
The user or role that the broker runs as requires at least the following policy 
(will scope this down further before public release).

If you do not intend to assume a role for stack creation, this role will also require any permissions
needed to manage resources it creates.
 
```json
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Action": [
            "cloudformation:*",
            "s3:*",
            "dynamodb:*",
            "sts:AssumeRole",
            "ssm:*"
         ],
         "Resource": [
            "*"
         ],
         "Effect": "Allow"
      }
   ]
}
```

#### Building the binary

```bash
make build # OSX binary
make linux # linux binary
```

### Launching the broker

Here is an example of a minimal set of parameters to launch the broker with, you can add -keyId and -secretKey 
for hardcoded credentials, if no credentials are provided the broker will search the cred chain.

```bash
aws-servicebroker \
    -insecure \
    -alsologtostderr \
    -region us-east-1 \
    -s3Bucket awsservicebrokeralpha \
    -s3Key pcf/templates/ \
    -s3Region us-west-2 \
    -port 3199 \
    -tableName awssb \
    -enableBasicAuth \
    -basicAuthUser admin \
    -basicAuthPass <Password-Goes-Here>
```

### Testing in minikube

Once it's running you can register the broker with your application platform, if using the minikube functional testing image:

```bash
cat <<EOF > "./broker-resource.yaml"
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: aws-service-broker
spec:
  url: "http://172.17.0.2:3199/"
EOF
kubectl create -f ./broker-resource.yaml 
```

check that url and port in the spec matches your environment.

Now you should be able to use svcat to interrogate the broker and provision/bind resources

```bash
# List brokers and their status
svcat get brokers

# List services offered
svcat get classes

# list service plans
svcat get plans

# Provision polly 
svcat provision test-polly --class polly --plan default

# list provisioned services
svcat get instances

# create binding to polly service
svcat bind test-polly --name test-polly-bind

# list bindings
svcat get bindings

# show credentials in bind
kubectl get secret/test-polly-bind -o yaml

# delete binding
svcat unbind --name test-polly-bind

# delete service
svcat deprovision test-polly
```

## Passing In Credentials Via Parameters

The **aws_access_key**, **aws_secret_key** and **aws_session_token** can
be passed in as parameters to the provision request. If provided, they will
be used in place of the aws service catalog process role.

For example

```
# svcat provision my-instance-name \
	-n my-app \
	--class my-instance-class \
	--plan prd \
	-p VpcId=vpc-123451234512341234,aws_access_key=bacdbcadbcadbcad,aws_secret_key=abcdabcdabcdabcdabcdabcdabcdabcd,aws_session_token=averylongsessssssssssssssssssssiontoken
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
  aws_session_token: averylongsessssssssssssssssssssiontoken
```

## Managing Resources Via Assumed Role

The aws-service-broker has the ability to assume a role for all resources it manages.

This role can be in the same account, or a seperate target account.

To setup the role, assume admin credentials in the account where the role will reside
and create the role for the aws-service-broker to assume.

```
service_broker_account_id=123456654321 # role where the service broker will run, will be the same as the target if in single account

aws cloudformation create-stack \
    --stack-name AwsServiceBrokerWorkerRole \
    --template-body file://setup/aws-service-broker-worker.json \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters ParameterKey=ServiceBrokerAccountId,ParameterValue=$service_broker_account_id
```

To do you this you must ensure that the role the **aws-service-broker** is running allows it to assume the target role.

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
