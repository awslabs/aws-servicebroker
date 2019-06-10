# AWS Service Broker - Amazon Aurora for PostgreSQL Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png" width="108"> <p align="center">Amazon Aurora is a relational database service that combines the speed and availability of high-end commercial databases with the simplicity and cost-effectiveness of open source databases. The PostgreSQL-compatible edition of Aurora delivers up to 3X the throughput of standard PostgreSQL running on the same hardware, enabling existing PostgreSQL applications and tools to run without requiring modification. The combination of PostgreSQL compatibility with Aurora enterprise database capabilities provides an ideal target for commercial database migrations. https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/CHAP_AuroraOverview.html</p>

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

Creates an Amazon Aurora for PostgreSQL database optimised for production use

Pricing: https://aws.amazon.com/rds/aurora/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
DBInstanceClass|Database Instance Class (only PostgreSQL 10.6 or later support the db.r5 instance classes)|db.r4.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
DBEngineVersion|Select Aurora PostgreSQL Database Engine Version|9.6.8|9.6.8, 9.6.9, 10.4, 10.5, 10.6
RDSTimeZone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Sun|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowStartTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false


### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create Cluster in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the RDS Cluster into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|3
NumberofAuroraReplicas|Number of Aurora Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|2
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available. |Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto".|Auto
DBPort|TCP/IP Port for the Database Instance|5432
StorageEncrypted|Indicates whether the DB instance is encrypted.|true 
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|35
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
DBUsername|Database master username|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto

<a id="param-dev" />

## dev

Creates an Amazon Aurora for PostgreSQL database optimised for dev/test use

Pricing: https://aws.amazon.com/rds/aurora/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
DBInstanceClass|Database Instance Class (only PostgreSQL 10.6 or later support the db.r5 instance classes)|db.r4.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
DBEngineVersion|Select Aurora PostgreSQL Database Engine Version|9.6.8|9.6.8, 9.6.9, 10.4, 10.5, 10.6
RDSTimeZone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Sun|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS Cluster in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the RDS Cluster into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
NumberofAuroraReplicas|Number of Aurora Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|1
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available. |Auto 
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|28
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto 
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto".|Auto 
DBPort|TCP/IP Port for the Database Instance|5432
StorageEncrypted|Indicates whether the DB instance is encrypted.|true 
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|0
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|60
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true 
DBUsername|Database master username|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto 
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Sun
PreferredMaintenanceWindowStartTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00
PreferredMaintenanceWindowEndTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00

<a id="param-custom" />

## custom

Creates an Amazon Aurora for PostgreSQL database with custom configuration

Pricing: https://aws.amazon.com/rds/aurora/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string
DBUsername|Master database Username|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available. |Auto|Auto
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|35|35
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27|27
DBName|The name of the database to create when the DB instance is created, will be autogenerated if set to "Auto".|Auto|Auto
DBPort|TCP/IP Port for the Database Instance|5432|5432
DBUsername|Database master username|master|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto|Auto
DBEngineVersion|Select Aurora PostgreSQL Database Engine Version|9.6.8|9.6.8, 9.6.9, 10.4, 10.5, 10.6
DBInstanceClass|Database Instance Class (only PostgreSQL 10.6 or later support the db.r5 instance classes)|db.r4.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
DeletionProtection|Indicates if the DB cluster should have deletion protection enabled. |false|true, false
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2|2, 3, 4, 5
NumberofAuroraReplicas|Number of Aurora Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|0|0, 1, 2
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1|0, 1, 5, 10, 15, 30, 60
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true|true, false
StorageEncrypted|Indicates whether the DB instance is encrypted.|true|true, false
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowStartTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
RDSTimeZone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC


### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
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
  name: aurorapostgresql-production-minimal-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: aurorapostgresql-production-complete-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED    
    DBInstanceClass: db.r4.large # OPTIONAL
    DBEngineVersion: 9.6.8 # OPTIONAL
    RDSTimeZone: UTC # OPTIONAL
    PreferredMaintenanceWindowDay: Sun # OPTIONAL
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
  name: aurorapostgresql-dev-minimal-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: aurorapostgresql-dev-complete-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED    
    DBInstanceClass: db.m4.xlarge # OPTIONAL
    EngineVersion: 9.6.8 # OPTIONAL
    PostgresServerTimezone: UTC # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: aurorapostgresql-custom-minimal-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    DBUsername: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: aurorapostgresql-custom-complete-example
spec:
  clusterServiceClassExternalName: aurorapostgresql
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    DBUsername: [VALUE] # REQUIRED        
    AutoMinorVersionUpgrade: true # OPTIONAL
    AvailabilityZones: Auto # OPTIONAL
    BackupRetentionPeriod: 35 # OPTIONAL
    CidrBlocks: Auto # OPTIONAL
    CidrSize: 27 # OPTIONAL    
    DBInstanceClass: db.r4.large # OPTIONAL
    DeletionProtection: false #OPTIONAL
    DBName: Auto # OPTIONAL
    EngineVersion: 9.6.8 # OPTIONAL
    DBPassword: Auto # OPTIONAL
    MonitoringInterval: 1 # OPTIONAL
    NumberofAuroraReplicas: 0 # OPTIONAL
    NumberOfAvailabilityZones: 2 # OPTIONAL
    DBPort: 5432 # OPTIONAL
    RDSTimeZone: UTC # OPTIONAL    
    PreferredMaintenanceWindowDay: Sun # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
    StorageEncrypted: true # OPTIONAL    
```

