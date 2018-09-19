# AWS Service Broker - Amazon Redshift Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Database_AmazonRedshift_LARGE.png" width="108"> <p align="center">Amazon Redshift is a fast, fully managed, petabyte-scale data warehouse service that makes it simple and cost-effective to efficiently analyze all your data using your existing business intelligence tools. It is optimized for datasets ranging from a few hundred gigabytes to a petabyte or more and costs less than $1,000 per terabyte per year, a tenth the cost of most traditional data warehousing solutions.
https://aws.amazon.com/documentation/redshift/</p>

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

Creates an Amazon Redshift database optimised for production use

Pricing: https://aws.amazon.com/redshift/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
NodeType|The node type that is provisioned for this cluster.|dc1.large|dc1.large, dc1.8xlarge, ds1.xlarge, ds1.8xlarge, ds2.xlarge, ds2.8xlarge
NumberOfNodes|The number of compute nodes in the cluster. If you specify multi-node for the ClusterType parameter, you must specify a number greater than 1. You can not specify this parameter for a single-node cluster. min 2 max 32|3|
PubliclyAccessible|Indicates whether the Cluster is an Internet-facing instance.|False|True, False
UseElasticIP|For public accessable clusters which require a static IP, assign a EIP|False|True, False

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
VpcId|The ID of the VPC to launch the RDS instance into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
DBName|The name of the database to create when the DB instance is created.|Auto
MasterUsername|Master Cluster Username|master
MasterUserPassword|Master user Cluster Password|Auto
AllowVersionUpgrade|When a new version of Amazon Redshift is released, tells whether upgrades can be applied to the engine that is running on the cluster. The upgrades are applied during the maintenance window. The default value is True.|False
PortNumber|The port number for the Cluster to listen on|15439
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|35
ClusterType|The type of cluster. Specify single-node or multi-node (default).  Number of nodes must be greater than 1 for multi-node|multi-node
LogBucketName|Must be a valid S3 Bucket in same Region, if no bucket is provided audit logging will not be enabled for this cluster|
StorageEncrypted|Indicates whether the Cluster storage is encrypted.|True
<a id="param-custom" />

## custom

Creates an Amazon Redhift database with a custom configuration

Pricing: https://aws.amazon.com/redshift/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string
MasterUsername|Master Cluster Username|string
LogBucketName|Must be a valid S3 Bucket in same Region, if no bucket is provided audit logging will not be enabled for this cluster|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|3|2, 3, 4, 5
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27|
NodeType|The node type that is provisioned for this cluster.|dc1.large|dc1.large, dc1.8xlarge, ds1.xlarge, ds1.8xlarge, ds2.xlarge, ds2.8xlarge
NumberOfNodes|The number of compute nodes in the cluster. If you specify multi-node for the ClusterType parameter, you must specify a number greater than 1. You can not specify this parameter for a single-node cluster. min 2 max 32|3|
MasterUserPassword|Master user Cluster Password|Auto|
ClusterType|The type of cluster. Specify single-node or multi-node (default).  Number of nodes must be greater than 1 for multi-node|multi-node|single-node, multi-node
AllowVersionUpgrade|When a new version of Amazon Redshift is released, tells whether upgrades can be applied to the engine that is running on the cluster. The upgrades are applied during the maintenance window. The default value is True.|True|True, False
StorageEncrypted|Indicates whether the Cluster storage is encrypted.|True|True, False
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|35|
PortNumber|The port number for the Cluster to listen on|5439|
PubliclyAccessible|Indicates whether the Cluster is an Internet-facing instance.|False|True, False
UseElasticIP|For public accessable clusters which require a static IP, assign a EIP|False|True, False
DBName|The name of the database to create when the DB instance is created.|Auto|

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
VpcId|The ID of the VPC to launch the RDS instance into||

<a id="bind-credentials" />

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
ENDPOINT_ADDRESS|

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
  name: redshift-production-minimal-example
spec:
  clusterServiceClassExternalName: dh-redshift
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: redshift-production-complete-example
spec:
  clusterServiceClassExternalName: dh-redshift
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    NodeType: dc1.large # OPTIONAL
    NumberOfNodes: 3 # OPTIONAL
    PubliclyAccessible: False # OPTIONAL
    UseElasticIP: False # OPTIONAL
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: redshift-custom-minimal-example
spec:
  clusterServiceClassExternalName: dh-redshift
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    MasterUsername: [VALUE] # REQUIRED
    LogBucketName: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: redshift-custom-complete-example
spec:
  clusterServiceClassExternalName: dh-redshift
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    MasterUsername: [VALUE] # REQUIRED
    LogBucketName: [VALUE] # REQUIRED
    NumberOfAvailabilityZones: 3 # OPTIONAL
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    AvailabilityZones: Auto # OPTIONAL
    CidrBlocks: Auto # OPTIONAL
    CidrSize: 27 # OPTIONAL
    NodeType: dc1.large # OPTIONAL
    NumberOfNodes: 3 # OPTIONAL
    MasterUserPassword: Auto # OPTIONAL
    ClusterType: multi-node # OPTIONAL
    AllowVersionUpgrade: True # OPTIONAL
    StorageEncrypted: True # OPTIONAL
    BackupRetentionPeriod: 35 # OPTIONAL
    PortNumber: 5439 # OPTIONAL
    PubliclyAccessible: False # OPTIONAL
    UseElasticIP: False # OPTIONAL
    DBName: Auto # OPTIONAL
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***