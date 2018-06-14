# AWS Service Broker - Amazon RDS for MariaDB Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png" width="108"> <p align="center">MariaDB is a popular open source relational database created by the original developers of MySQL. Amazon RDS makes it easy to set up, operate, and scale MariaDB deployments in the cloud. With Amazon RDS, you can deploy scalable MariaDB databases in minutes with cost-efficient and resizable hardware capacity. Amazon RDS frees you up to focus on your application by managing time-consuming database administration tasks including backups, software patching, monitoring, scaling and replication.
https://aws.amazon.com/documentation/rds/</p>

Table of contents
=================

* [Parameters](#parameters)
  * [production](#param-production)
  * [dev](#param-dev)
  * [custom](#param-custom)
* [Bind Credentials](#bind-credentials)
* [Examples](#kubernetes-openshift-examples)
  * [production](#example-production)
  * [dev](#example-dev)
  * [custom](#example-custom)

<a id="parameters" />

# Parameters

<a id="param-production" />

## production

Creates an Amazon RDS for MariaDB database optimised for production use

Pricing: https://aws.amazon.com/rds/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
AllocatedStorageAndIops|Storage/IOPS to allocate|100 GB / 1,000 IOPS|100 GB / 1,000 IOPS, 300 GB / 3,000 IOPS, 600 GB / 6,000 IOPS, 1,000 GB / 10,000 IOPS, 1,500 GB / 15,000 IOPS, 2,000 GB / 20,000 IOPS, 3,000 GB / 30,000 IOPS, 4,000 GB / 40,000 IOPS, 6,000 GB / 60,000 IOPS
DBInstanceClass|The compute and memory capacity of the DB instance.|db.m4.xlarge|db.m1.small, db.m1.medium, db.m1.large, db.m1.xlarge, db.m2.xlarge, db.m2.2xlarge, db.m2.4xlarge, db.m3.medium, db.m3.large, db.m3.xlarge, db.m3.2xlarge, db.m4.large, db.m4.xlarge, db.m4.2xlarge, db.m4.4xlarge, db.m4.10xlarge, db.r3.large, db.r3.xlarge, db.r3.2xlarge, db.r3.4xlarge, db.r3.8xlarge, db.t2.micro, db.t2.small, db.t2.medium, db.t2.large
EngineVersion|The version number of the database engine to use.|10.1.23|10.0.17, 10.0.24, 10.0.28, 10.0.31, 10.1.14, 10.1.19, 10.1.23
MysqlServerTimezone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false

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
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|true
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto"|Auto
PortNumber|The port number for the database server to listen on|13306
StorageEncrypted|Indicates whether the DB instance is encrypted.|true
StorageType|Specifies the storage type to be associated with the DB instance.|io1
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|true
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|35
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to true.|false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
MasterUsername|Master database Username|master
MasterUserPassword|Master user database Password, if left at default a 32 character password will be generated|Auto
<a id="param-dev" />

## dev

Creates an Amazon RDS for MariaDB database optimised for dev/test use

Pricing: https://aws.amazon.com/rds/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
AllocatedStorageAndIops|Storage/IOPS to allocate|100 GB / 1,000 IOPS|100 GB / 1,000 IOPS, 300 GB / 3,000 IOPS, 600 GB / 6,000 IOPS, 1,000 GB / 10,000 IOPS, 1,500 GB / 15,000 IOPS, 2,000 GB / 20,000 IOPS, 3,000 GB / 30,000 IOPS, 4,000 GB / 40,000 IOPS, 6,000 GB / 60,000 IOPS
DBInstanceClass|The compute and memory capacity of the DB instance.|db.m4.xlarge|db.m1.small, db.m1.medium, db.m1.large, db.m1.xlarge, db.m2.xlarge, db.m2.2xlarge, db.m2.4xlarge, db.m3.medium, db.m3.large, db.m3.xlarge, db.m3.2xlarge, db.m4.large, db.m4.xlarge, db.m4.2xlarge, db.m4.4xlarge, db.m4.10xlarge, db.r3.large, db.r3.xlarge, db.r3.2xlarge, db.r3.4xlarge, db.r3.8xlarge, db.t2.micro, db.t2.small, db.t2.medium, db.t2.large
EngineVersion|The version number of the database engine to use.|10.1.23|10.0.17, 10.0.24, 10.0.28, 10.0.31, 10.1.14, 10.1.19, 10.1.23
MysqlServerTimezone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false

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
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|false
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|28
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto"|Auto
PortNumber|The port number for the database server to listen on|13306
StorageEncrypted|Indicates whether the DB instance is encrypted.|true
StorageType|Specifies the storage type to be associated with the DB instance.|io1
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|false
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|0
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|60
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to true.|false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
MasterUsername|Master database Username|master
MasterUserPassword|Master user database Password, if left at default a 32 character password will be generated|Auto
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|04:00-06:00
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Sat
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|07:00
<a id="param-custom" />

## custom

Creates an Amazon RDS for MariaDB database with custom configuration

Pricing: https://aws.amazon.com/rds/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string
MasterUsername|Master database Username|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
AllocatedStorageAndIops|Storage/IOPS to allocate|100 GB / 1,000 IOPS|100 GB / 1,000 IOPS, 300 GB / 3,000 IOPS, 600 GB / 6,000 IOPS, 1,000 GB / 10,000 IOPS, 1,500 GB / 15,000 IOPS, 2,000 GB / 20,000 IOPS, 3,000 GB / 30,000 IOPS, 4,000 GB / 40,000 IOPS, 6,000 GB / 60,000 IOPS
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to true.|false|true, false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true|true, false
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|35|
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27|
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|true|true, false
DBInstanceClass|The compute and memory capacity of the DB instance.|db.m4.xlarge|db.m1.small, db.m1.medium, db.m1.large, db.m1.xlarge, db.m2.xlarge, db.m2.2xlarge, db.m2.4xlarge, db.m3.medium, db.m3.large, db.m3.xlarge, db.m3.2xlarge, db.m4.large, db.m4.xlarge, db.m4.2xlarge, db.m4.4xlarge, db.m4.10xlarge, db.r3.large, db.r3.xlarge, db.r3.2xlarge, db.r3.4xlarge, db.r3.8xlarge, db.t2.micro, db.t2.small, db.t2.medium, db.t2.large
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto"|Auto|
EngineVersion|The version number of the database engine to use.|10.1.23|10.0.17, 10.0.24, 10.0.28, 10.0.31, 10.1.14, 10.1.19, 10.1.23
MasterUserPassword|Master user database Password, if left at default a 32 character password will be generated|Auto|
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1|0, 1, 5, 10, 15, 30, 60
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|true|true, false
MysqlServerTimezone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2|2, 3, 4, 5
PortNumber|The port number for the database server to listen on|10001|
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false
StorageEncrypted|Indicates whether the DB instance is encrypted.|true|true, false
StorageType|Specifies the storage type to be associated with the DB instance.|io1|io1, gp2, standard

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
MASTER_USERNAME|
MASTER_PASSWORD|
PORT|
DB_NAME|

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
  name: rdsmariadb-production-minimal-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rdsmariadb-production-complete-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    AllocatedStorageAndIops: 100 GB / 1,000 IOPS # OPTIONAL
    DBInstanceClass: db.m4.xlarge # OPTIONAL
    EngineVersion: 10.1.23 # OPTIONAL
    MysqlServerTimezone: UTC # OPTIONAL
    PreferredBackupWindow: 00:00-02:00 # OPTIONAL
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
```
<a id="example-dev" />

## dev

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rdsmariadb-dev-minimal-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rdsmariadb-dev-complete-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    AllocatedStorageAndIops: 100 GB / 1,000 IOPS # OPTIONAL
    DBInstanceClass: db.m4.xlarge # OPTIONAL
    EngineVersion: 10.1.23 # OPTIONAL
    MysqlServerTimezone: UTC # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rdsmariadb-custom-minimal-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    MasterUsername: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rdsmariadb-custom-complete-example
spec:
  clusterServiceClassExternalName: dh-rdsmariadb
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    MasterUsername: [VALUE] # REQUIRED
    AllocatedStorageAndIops: 100 GB / 1,000 IOPS # OPTIONAL
    AllowMajorVersionUpgrade: false # OPTIONAL
    AutoMinorVersionUpgrade: true # OPTIONAL
    AvailabilityZones: Auto # OPTIONAL
    BackupRetentionPeriod: 35 # OPTIONAL
    CidrBlocks: Auto # OPTIONAL
    CidrSize: 27 # OPTIONAL
    CopyTagsToSnapshot: true # OPTIONAL
    DBInstanceClass: db.m4.xlarge # OPTIONAL
    DBName: Auto # OPTIONAL
    EngineVersion: 10.1.23 # OPTIONAL
    MasterUserPassword: Auto # OPTIONAL
    MonitoringInterval: 1 # OPTIONAL
    MultiAZ: true # OPTIONAL
    MysqlServerTimezone: UTC # OPTIONAL
    NumberOfAvailabilityZones: 2 # OPTIONAL
    PortNumber: 10001 # OPTIONAL
    PreferredBackupWindow: 00:00-02:00 # OPTIONAL
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
    StorageEncrypted: true # OPTIONAL
    StorageType: io1 # OPTIONAL
```

***NOTE: This documentation is auto-generated using available metadata in the ServiceClass and CloudFormation Template. Please do not PR changes to this file, if a change is needed, update the source metadata and ci will re-generate documentation on merge.***