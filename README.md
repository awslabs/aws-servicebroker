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
 
create an IAM role with the CloudFormation service in the trust policy and attach the AdministratorAccess managed 
policy, the arn of this role must be provided to the broker with -roleArn

The user or role that the broker runs as requires at least the following policy 
(will scope this down further before public release)
 
```json
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Action": [
            "cloudformation:*",
            "s3:*",
            "dynamodb:*",
            "iam:PassRole",
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
    -s3Key pcf/templates \
    -s3Region us-west-2 \
    -roleArn arn:aws:iam::1231231234:role/aws-service-broker-cloudformation \
    -port 3199 \
    -tableName awssb
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
