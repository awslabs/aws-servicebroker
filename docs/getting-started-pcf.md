# Getting Started Guide - Pivotal Cloud Foundry

*Note:* The use of the AWS Service Broker in Cloud Foundry is at an alpha stage, bugs and possible update related breaking 
changes may manifest. Use of the AWS Service Broker in Pivotal Cloud Foundry is not recommended for production at this 
time.

### Prerequisites

#### PCF 2.1+

Testing on V2.1 was done using the [Pivotal Cloud Foundry on the AWS Cloud Quick Start](https://aws.amazon.com/quickstart/architecture/pivotal-cloud-foundry/). 
Though not tested, older PCF versions may work.

#### IAM Roles/Users

The AWS Service Broker packages all services into CloudFormation templates that are executed by the broker. The broker 
can use a role if the broker is installed into an EC2 instance with access to the ec2 metadata endpoint 
(169.254.169.254). Alternatively, an IAM user and static keypair can be created for the broker to use. The IAM user/role 
requires the following IAM policy:

***Service User/Role***
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "cloudformation:*",
                "ssm:*",
                "dynamodb:*",
                "s3:*"
            ],
            "Resource": [
                "*"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "iam:PassRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/AWSServiceBrokerCFNRole"
            ],
            "Effect": "Allow"
        }
    ]
}
```

A [CloudFormation service role](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-iam-servicerole.html) 
is also required, an example of a broad policy to enable all current service plans is included below, this can be scoped 
down if only specific services are required:

***CloudFormation Role***
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "cloudformation:*",
                "iam:*",
                "kms:*",
                "ssm:*",
                "ec2:*",
                "lambda:*",
                "athena:*",
                "dynamodb:*",
                "elasticache:*",
                "elasticmapreduce:*",
                "rds:*",
                "redshift:*",
                "route53:*",
                "s3:*",
                "sns:*",
                "sqs:*",
                "polly:*",
                "lex:*",
                "translate:*",
                "rekognition:*",
                "kinesis:*"
            ],
            "Resource": [
                "*"
            ],
            "Effect": "Allow"
        }
    ]
}
```

#### DynamoDB Table

The broker uses a DynamoDB table as a persistent store for service instances and as a distributed cache/lock. To create 
the table the following command can be run using the AWS CLI:

```bash
aws dynamodb create-table --attribute-definitions \
AttributeName=id,AttributeType=S AttributeName=userid,AttributeType=S \
AttributeName=type,AttributeType=S --key-schema AttributeName=id,KeyType=HASH \
AttributeName=userid,KeyType=RANGE --global-secondary-indexes \
'IndexName=type-userid-index,KeySchema=[{AttributeName=type,KeyType=HASH},{AttributeName=userid,KeyType=RANGE}],Projection={ProjectionType=INCLUDE,NonKeyAttributes=[id,userid,type,locked]},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}' \
--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
--region us-east-1 --table-name awssb
```

*Note:* the Service User/Role policy expects the CloudFormation role to be named AWSServiceBrokerCFNRole, if you name it 
something else you will also need to update this policy to reflect the name.

### Installation

* Download the [AWS Service Broker Tile](https://awsservicebrokeralpha.s3.amazonaws.com/pcf/aws-service-broker-latest.pivotal)
* Login to Ops Manager and import the tile
* Complete configuration in the `AWS Service Broker Configuration` section. Take note of the following fields:
  * `AWS Access Key ID` and `AWS Secret Access` - if you are using an ec2 instance role attached to the broker hosts, 
  specify "use-role" as the value for both fields, otherwise specify the credentials for the user created in the 
  prerequisites section of this guide.
  * `AWS Region ` - this is the default region for the broker to deploy services into, and must match the region that the 
  DynamoDB table created in the prerequisisites section of this guide was created in (this will be decoupled in an upcoming update).
  * `Amazon S3 Bucket` - specify `awsservicebroker`
  * `Amazon S3 Key Prefix` - specify `templates/latest/`
  * `Amazon S3 Region` - specify `us-east-1`
  * `Amazon S3 Key Suffix` - specify `-main.yaml`
  * `Amazon DynamoDB table name` - specify the name of the table created in the prerequisites section of this guide, default is `awssb`


