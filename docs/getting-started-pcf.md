# Getting Started Guide - Pivotal Cloud Foundry


### Compatibility

Testing on V2.1/2.2/2.4 was done using the [Pivotal Cloud Foundry on the AWS Cloud Quick Start](https://aws.amazon.com/quickstart/architecture/pivotal-cloud-foundry/). 
Though not tested, older PCF versions may work.

### Installation

* Download the latest tile from the [releases page](https://github.com/awslabs/aws-servicebroker/releases)
* Login to Ops Manager and import the tile
* Complete configuration in the `AWS Service Broker Configuration` section. Take note of the following fields:
  * `Broker ID` - An ID to use for partitioning broker data in DynamoDb. if multiple brokers are used in the same AWS account, this value must be unique per broker. This is a customer selected string. 
  * `AWS Access Key ID` and `AWS Secret Access` (_**REQUIRED**_) -  Specify the credentials for the user created in the prerequisites section of this guide. If you are using an ec2 instance role attached to the broker hosts, leave these fields blank. 
  * `Target AWS Account ID` and `Target IAM Role Name` - if you would like to provision into a different account, or use a 
  different role for provisioning, populate these with the account and role details. The role specified must allow the 
  broker user/role to assume it
  * `AWS Region ` - this is the default region for the broker to deploy services into, and must match the region that the 
  DynamoDB table created in the prerequisisites section of this guide was created in (this will be decoupled in an upcoming update).
  * `Amazon S3 Bucket` - specify `awsservicebroker`
  * `Amazon S3 Key Prefix` - specify `templates/latest/`
  * `Amazon S3 Region` - specify `us-east-1`
  * `Amazon S3 Key Suffix` - specify `-main.yaml`
  * `Amazon DynamoDB table name` - specify the name of the table created in the prerequisites section of this guide, default is `awssb`

![Service Broker Install](images/SBinstall01.gif)


