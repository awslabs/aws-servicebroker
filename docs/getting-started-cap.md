# Getting Started Guide - SUSE Cloud Application Platform

This guide uses helm, for documentation on installing the helm and tiller see [https://docs.helm.sh/using_helm/#install-helm](https://docs.helm.sh/using_helm/#install-helm)

### Prerequisites
Deploying and using the AWS Service Broker requires the following:

1. Cloud Application Platform has been successfully deployed on Amazon EKS as described in the deployment guide (see Chapter 10, Deploying SUSE Cloud Application Platform on Amazon Elastic Kubernetes Service (EKS)) here: https://documentation.suse.com/suse-cap/1/single-html/cap-guides/#cha-cap-depl-eks
2. The AWS Command Line Interface (CLI)
3. The OpenSSL command line tool
4. Ensure the user or role running the broker has an IAM policy set as specified in the IAM section of the AWS Service Broker documentation on Github.

### Setup
Create the required DynamoDB table where the AWS service broker will store its data. This example creates a table named awssb:

aws dynamodb create-table \
		--attribute-definitions \
			AttributeName=id,AttributeType=S \
			AttributeName=userid,AttributeType=S \
			AttributeName=type,AttributeType=S \
		--key-schema \
			AttributeName=id,KeyType=HASH \
			AttributeName=userid,KeyType=RANGE \
		--global-secondary-indexes \
			'IndexName=type-userid-index,KeySchema=[{AttributeName=type,KeyType=HASH},{AttributeName=userid,KeyType=RANGE}],Projection={ProjectionType=INCLUDE,NonKeyAttributes=[id,userid,type,locked]},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}' \
		--provisioned-throughput \
			ReadCapacityUnits=5,WriteCapacityUnits=5 \
		--region ${AWS_REGION} --table-name awssb
Wait until the table has been created. When it is ready, the TableStatus will change to ACTIVE. Check the status using the describe-table command:

aws dynamodb describe-table --table-name awssb
(For more information about the describe-table command, see https://docs.aws.amazon.com/cli/latest/reference/dynamodb/describe-table.html.)

Set a name for the Kubernetes namespace you will install the service broker to. This name will also be used in the service broker URL:

BROKER_NAMESPACE=aws-sb
Create a server certificate for the service broker:

Create and use a separate directory to avoid conflicts with other CA files:

mkdir /tmp/aws-service-broker-certificates && cd $_
Get the CA certificate:

kubectl get secret --namespace scf --output jsonpath='{.items[*].data.internal-ca-cert}' | base64 -di > ca.pem
Get the CA private key:

kubectl get secret --namespace scf --output jsonpath='{.items[*].data.internal-ca-cert-key}' | base64 -di > ca.key
Create a signing request. Replace BROKER_NAMESPACE with the namespace assigned in Step 3:

openssl req -newkey rsa:4096 -keyout tls.key.encrypted -out tls.req -days 365 \
  -passout pass:1234 \
  -subj '/CN=aws-servicebroker.'${BROKER_NAMESPACE} -batch \
  -subj '/CN=aws-servicebroker-aws-servicebroker.aws-sb.svc.cluster.local' -batch
  </dev/null
Decrypt the generated broker private key:

openssl rsa -in tls.key.encrypted -passin pass:1234 -out tls.key
Sign the request with the CA certificate:

openssl x509 -req -CA ca.pem -CAkey ca.key -CAcreateserial -in tls.req -out tls.pem
Install the AWS service broker as documented at https://github.com/awslabs/aws-servicebroker/blob/master/docs/getting-started-k8s.md. Skip the installation of the Kubernetes Service Catalog. While installing the AWS Service Broker, make sure to update the Helm chart version (the version as of this writing is 1.0.0-beta.3). For the broker install, pass in a value indicating the Cluster Service Broker should not be installed (for example --set deployClusterServiceBroker=false). Ensure an account and role with adequate IAM rights is chosen (see Section 10.12.1, “Prerequisites”:

helm install aws-sb/aws-servicebroker \
	     --name aws-servicebroker \
	     --namespace $BROKER_NAMESPACE \
	     --version 1.0.2 \
	     --set aws.secretkey=$AWS_ACCESS_KEY \
	     --set aws.accesskeyid=$AWS_KEY_ID \
	     --set deployClusterServiceBroker=false \
	     --set tls.cert="$(base64 -w0 tls.pem)" \
	     --set tls.key="$(base64 -w0 tls.key)" \
	     --set-string aws.targetaccountid=$AWS_TARGET_ACCOUNT_ID \
	     --set aws.targetrolename=$AWS_TARGET_ROLE_NAME \
	     --set aws.tablename=awssb \
	     --set aws.vpcid=$VPC_ID \
	     --set aws.region=$AWS_REGION \
	     --set authenticate=false
To find the values of aws.targetaccoundid, aws.targetrolename, and vpcId run the following command.

aws eks describe-cluster --name $CLUSTER_NAME
For aws.targetaccoundid and aws.targetrolename, examine the cluster.roleArn field. For vpcId, refer to the cluster.resourcesVpcConfig.vpcId field.

Log into your Cloud Application Platform deployment. Select an organization and space to work with, creating them if needed:

cf api --skip-ssl-validation https://api.example.com
cf login -u admin -p password
cf create-org org
cf create-space space
cf target -o org -s space
Create a service broker in scf. Note the name of the service broker should be the same as the one specified for the --name flag in the helm install step (for example aws-servicebroker. Note that the username and password parameters are only used as dummy values to pass to the cf command:

cf create-service-broker aws-servicebroker username password https://aws-servicebroker-aws-servicebroker.aws-sb.svc.cluster.local
Verify the service broker has been registered:

cf service-brokers
List the available service plans:

cf service-access
Enable access to a service. This example uses the -p to enable access to a specific service plan. See https://github.com/awslabs/aws-servicebroker/blob/master/templates/rdsmysql/template.yaml for information about all available services and their associated plans:

cf enable-service-access rdsmysql -p custom
Create a service instance. As an example, a custom MySQL instance can be created as:

cf create-service rdsmysql custom mysql-instance-name -c '{
  "AccessCidr": "192.0.2.24/32",
  "BackupRetentionPeriod": 0,
  "MasterUsername": "master",
  "DBInstanceClass": "db.t2.micro",
  "EngineVersion": "5.7.17",
  "PubliclyAccessible": "true",
  "region": "$AWS_REGION",
  "StorageEncrypted": "false",
  "VpcId": "$VPC_ID",
  "target_account_id": "$AWS_TARGET_ACCOUNT_ID",
  "target_role_name": "$AWS_TARGET_ROLE_NAME"
}'

### Cleanup
When the AWS Service Broker and its services are no longer required, perform the following steps:

Unbind any applications using any service instances then delete the service instance:

cf unbind-service my_app mysql-instance-name
cf delete-service mysql-instance-name
Delete the service broker in scf:

cf delete-service-broker aws-servicebroker
Delete the deployed Helm chart and the namespace:

helm delete --purge aws-servicebroker
kubectl delete namespace ${BROKER_NAMESPACE}
The manually created DynamoDB table will need to be deleted as well:

aws dynamodb delete-table --table-name awssb --region ${AWS_REGION}
