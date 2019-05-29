# AWS Service Broker - Amazon Aurora for MySQL Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png" width="108"> <p align="center"> Amazon Aurora is a relational database service that combines the speed and availability of high-end commercial databases with the simplicity and cost-effectiveness of open source databases. The MySQL-compatible edition of Aurora delivers up to 5X the throughput of standard MySQL running on the same hardware, and enables existing MySQL applications and tools to run without requiring modification. https://aws.amazon.com/rds/aurora/</p>

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

Creates an Amazon Aurora for MySQL database optimised for production use

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
DBInstanceClass| Database Instance Class (Aurora MySQL supports the db.t3 instance classes for Aurora MySQL 1.15 and higher, and all Aurora MySQL 2.* versions). Not applicable if Engine Mode is Serverless.|db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large, db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r3.8xlarge, db.r3.4xlarge, db.r3.2xlarge, db.r3.xlarge, db.r3.large, db.t3.2xlarge, db.t3.xlarge, db.t3.large, db.t3.medium, db.t3.small, db.t3.micro, db.t2.medium, db.t2.small
DBEngineVersion|Select Aurora Database Engine Version|Aurora-MySQL5.6.10a|Aurora-MySQL5.6.10a, Aurora-MySQL5.6-1.19.0, Aurora-MySQL5.7.12, Aurora-MySQL5.7-2.03.2, Aurora-MySQL5.7-2.03.3, Aurora-MySQL5.7-2.03.4, Aurora-MySQL5.7-2.03.4.2, Aurora-MySQL5.7-2.04.0, Aurora-MySQL5.7-2.04.1, Aurora-MySQL5.7-2.04.1.2  
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
DBPort|TCP/IP Port for the Database Instance|3306
StorageEncrypted|Indicates whether the DB instance is encrypted.|true 
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|35
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
DBUsername|Database master username|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto

<a id="param-dev" />

## dev

Creates an Amazon Aurora for MySQL database optimised for dev/test use

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
DBInstanceClass| Database Instance Class (Aurora MySQL supports the db.t3 instance classes for Aurora MySQL 1.15 and higher, and all Aurora MySQL 2.* versions). Not applicable if Engine Mode is Serverless.|db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large, db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r3.8xlarge, db.r3.4xlarge, db.r3.2xlarge, db.r3.xlarge, db.r3.large, db.t3.2xlarge, db.t3.xlarge, db.t3.large, db.t3.medium, db.t3.small, db.t3.micro, db.t2.medium, db.t2.small
DBEngineVersion|Select Aurora Database Engine Version|Aurora-MySQL5.6.10a|Aurora-MySQL5.6.10a, Aurora-MySQL5.6-1.19.0, Aurora-MySQL5.7.12, Aurora-MySQL5.7-2.03.2, Aurora-MySQL5.7-2.03.3, Aurora-MySQL5.7-2.03.4, Aurora-MySQL5.7-2.03.4.2, Aurora-MySQL5.7-2.04.0, Aurora-MySQL5.7-2.04.1, Aurora-MySQL5.7-2.04.1.2  
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
DBPort|TCP/IP Port for the Database Instance|3306
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

Creates an Amazon Aurora for MySQL database with custom configuration

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
DBPort|TCP/IP Port for the Database Instance|3306|3306
DBUsername|Database master username|master|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto|Auto
DBEngineVersion|Select Aurora Database Engine Version|Aurora-MySQL5.6.10a|Aurora-MySQL5.6.10a, Aurora-MySQL5.6-1.19.0, Aurora-MySQL5.7.12, Aurora-MySQL5.7-2.03.2, Aurora-MySQL5.7-2.03.3, Aurora-MySQL5.7-2.03.4, Aurora-MySQL5.7-2.03.4.2, Aurora-MySQL5.7-2.04.0, Aurora-MySQL5.7-2.04.1, Aurora-MySQL5.7-2.04.1.2  
DBInstanceClass| Database Instance Class (Aurora MySQL supports the db.t3 instance classes for Aurora MySQL 1.15 and higher, and all Aurora MySQL 2.* versions). Not applicable if Engine Mode is Serverless.|db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large, db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r3.8xlarge, db.r3.4xlarge, db.r3.2xlarge, db.r3.xlarge, db.r3.large, db.t3.2xlarge, db.t3.xlarge, db.t3.large, db.t3.medium, db.t3.small, db.t3.micro, db.t2.medium, db.t2.small
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
DBEngineMode||provisioned|provisioned, serverless
ServerlessMinCapacityUnit|The minimum capacity for an Aurora DB cluster in serverless DB engine mode. The minimum capacity must be less than or equal to the maximum capacity.|2|2, 4, 8, 16, 32, 64, 128, 256
ServerlessMaxCapacityUnit|The maximum capacity for an Aurora DB cluster in serverless DB engine mode. The maximum capacity must be greater than or equal to the minimum capacity.|64|2, 4, 8, 16, 32, 64, 128, 256
ServerlessAutoPause|Specifies whether to allow or disallow automatic pause for an Aurora DB cluster in serverless DB engine mode. A DB cluster can be paused only when its idle (it has no connections).|true|true, false
ServerlessSecondsUntilAutoPause|The time, in seconds, before an Aurora DB cluster in serverless mode is auto paused. Min = 300, Max = 86400 (24hrs)|300|300



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
  name: auroramysql-production-minimal-example
spec:
  clusterServiceClassExternalName: auroramysql
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: auroramysql-production-complete-example
spec:
  clusterServiceClassExternalName: auroramysql
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED    
    DBInstanceClass: db.r4.large # OPTIONAL
    DBEngineVersion: Aurora-MySQL5.6.10a # OPTIONAL
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
  name: auroramysql-dev-minimal-example
spec:
  clusterServiceClassExternalName: auroramysql
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: auroramysql-dev-complete-example
spec:
  clusterServiceClassExternalName: auroramysql
  clusterServicePlanExternalName: dev
  parameters:
    AccessCidr: [VALUE] # REQUIRED    
    DBInstanceClass: db.m4.xlarge # OPTIONAL
    EngineVersion: Aurora-MySQL5.6.10a # OPTIONAL
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
  name: auroramysql-custom-minimal-example
spec:
  clusterServiceClassExternalName: auroramysql
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
  name: auroramysql-custom-complete-example
spec:
  clusterServiceClassExternalName: auroramysql
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
    EngineVersion: Aurora-MySQL5.6.10a # OPTIONAL
    DBPassword: Auto # OPTIONAL
    MonitoringInterval: 1 # OPTIONAL
    NumberofAuroraReplicas: 0 # OPTIONAL
    NumberOfAvailabilityZones: 2 # OPTIONAL
    DBPort: 3302 # OPTIONAL
    RDSTimeZone: UTC # OPTIONAL    
    PreferredMaintenanceWindowDay: Sun # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PubliclyAccessible: false # OPTIONAL
    StorageEncrypted: true # OPTIONAL    
    DBEngineMode: provisioned # OPTIONAL
    ServerlessMinCapacityUnit: 2 # OPTIONAL
    ServerlessMaxCapacityUnit: 64 # OPTIONAL
    ServerlessAutoPause: true # OPTIONAL
    ServerlessSecondsUntilAutoPause: 300 # OPTIONAL
```

