# AWS Service Broker - Amazon RDS for Oracle

<img align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src=https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png width="108"><p align="center">Oracle Database is a relational database management system developed by Oracle. Amazon RDS makes it easy to set up, operate, and scale Oracle Database deployments in the cloud.</p>&nbsp;

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

<a id="parameters"></a>

# Parameters

<a id = "param-production"></a>

Creates an Amazon RDS for Oracle optimised for production use.  
Pricing: https://aws.amazon.com/rds/oracle/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
AllocatedStorageAndIops|Storage/IOPS to allocate|100GB 1000IOPS|100GB 1000IOPS, 300GB 3000IOPS, 600GB 6000IOPS, 1000GB 10000IOPS, 1500GB 15000IOPS, 2000GB 20000IOPS, 3000GB 30000IOPS, 4000GB 40000IOPS, 6000GB 60000IOPS
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to True.|false|true, false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true|true, false
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|true|true, false
DBInstanceClass|The compute and memory capacity of the DB instance.|db.r5.large|db.z1d.large, db.z1d.xlarge, db.z1d.2xlarge, db.z1d.3xlarge, db.z1d.6xlarge, db.z1d.12xlarge, db.m5.large, db.m5.2xlarge, db.m5.4xlarge, db.m5.12xlarge, db.m5.24xlarge, db.r5.large, db.r5.xlarge, db.r5.2xlarge, db.r5.4xlarge, db.r5.12xlarge, db.r5.24xlarge, db.x1e.xlarge, db.x1e.4xlarge, db.x1e.8xlarge, db.x1e.16xlarge, db.x1e.32xlarge, db.x1.16xlarge, db.x1.32xlarge, db.r4.large, db.r4.xlarge, db.r4.2xlarge, db.r4.4xlarge, db.r4.8xlarge, db.r4.16xlarge, db.r3.large, db.r3.xlarge, db.r3.2xlarge, db.r3.4xlarge, db.r3.8xlarge, db.m3.medium, db.m3.large, db.m3.xlarge, db.m3.2xlarge, db.m4.large, db.m4.xlarge, db.m4.2xlarge, db.m4.4xlarge, db.m4.10xlarge, db.m4.16xlarge, db.t3.small, db.t3.medium, db.t3.large, db.t3.xlarge, db.t3.2xlarge, db.t2.micro, db.t2.small, db.t2.medium, db.t2.large, db.t2.2xlarge
Engine|Database Engine|Enterprise-Edition-EE-Bring-Your-Own-License|Enterprise-Edition-EE-Bring-Your-Own-License, Standard-Edition-One-SE1-Bring-Your-Own-License, Standard-Edition-Two-SE2-Bring-Your-Own-License, Standard-Edition-SE-Bring-Your-Own-License, Standard-Edition-Two-SE2-License-Included, Standard-Edition-One-SE1-License-Included
EngineVersion|The version number of the database engine to use.|12.1.0.2.v16|12.2.0.1.ru-2019-04.rur-2019-04.r1, 12.2.0.1.ru-2019-01.rur-2019-01.r1, 12.2.0.1.ru-2018-10.rur-2018-10.r1, 12.1.0.2.v16, 12.1.0.2.v15, 12.1.0.2.v14, 12.1.0.2.v13, 12.1.0.2.v12, 12.1.0.2.v11, 12.1.0.2.v10, 11.2.0.4.v20, 11.2.0.4.v19, 11.2.0.4.v18, 11.2.0.4.v17, 11.2.0.4.v16, 11.2.0.4.v15, 11.2.0.4.v14, 11.2.0.4.v13, 11.2.0.4.v12, 11.2.0.4.v11, 11.2.0.4.v10
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1|0, 1, 5, 10, 15, 30, 60
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|true|true, false
ServerTimezone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
CharacterSetName|The character set being used by the database|AL32UTF8|AL32UTF8, AR8ISO8859P6, AR8MSWIN1256, BLT8ISO8859P13, BLT8MSWIN1257, CL8ISO8859P5, CL8MSWIN1251, EE8ISO8859P2, EL8ISO8859P7, EE8MSWIN1250, EL8MSWIN1253, IW8ISO8859P8, IW8MSWIN1255, JA16EUC, JA16EUCTILDE, JA16SJIS, JA16SJISTILDE, KO16MSWIN949, NE8ISO8859P10, NEE8ISO8859P4, TH8TISASCII, TR8MSWIN1254, US7ASCII, UTF8, VN8MSWIN1258, WE8ISO8859P1, WE8ISO8859P15, WE8ISO8859P9, WE8MSWIN1252, ZHS16GBK, ZHT16HKSCS, ZHT16MSWIN950, ZHT32EUC
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2|1, 2, 3, 4, 5
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false
ReadReplica|Number of Read Replicas to create. Only available on  Oracle Enterprise Edition (EE) engine with version 12.1 or higher|0|0, 1, 2
StorageEncrypted|Indicates whether the DB instance is encrypted.|true|true, false
StorageType|Specifies the storage type to be associated with the DB instance.|io1|io1, gp2, standard

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create RDS instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the RDS instance into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to True.|false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Setting 0 disables automatic snapshots, maximum value is 35|35
CidrBlocks|comma seperated list of CIDR blocks to place RDS into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|true
Engine|Database Engine|Enterprise-Edition-EE-Bring-Your-Own-License
EngineVersion|The version number of the database engine to use.|12.1.0.2.v16
MasterUserPassword|Master user database Password, if left at default a 32 character password will be generated|Auto
MasterUsername|Master database Username|master
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|true
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|3
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
PortNumber|The port number for the database server to listen on|1521
ReadReplica|Number of Read Replicas to create. Only available on  Oracle Enterprise Edition (EE) engine with version 12.1 or higher|1
StorageEncrypted|Indicates whether the DB instance is encrypted.|true
StorageType|Specifies the storage type to be associated with the DB instance.|io1

<a id = "param-dev"></a>

Creates an Amazon RDS for Oracle optimised for dev/test use.  
Pricing: https://aws.amazon.com/rds/oracle/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string
MasterUsername|Master database Username|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
AllocatedStorageAndIops|Storage/IOPS to allocate|100GB 1000IOPS|100GB 1000IOPS, 300GB 3000IOPS, 600GB 6000IOPS, 1000GB 10000IOPS, 1500GB 15000IOPS, 2000GB 20000IOPS, 3000GB 30000IOPS, 4000GB 40000IOPS, 6000GB 60000IOPS
AllowMajorVersionUpgrade|If you update the EngineVersion property to a version that's different from the DB instance's current major version, set this property to True.|false|true, false
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true|true, false
CopyTagsToSnapshot|Indicates whether to copy all of the user-defined tags from the DB instance to snapshots of the DB instance.|true|true, false
DBInstanceClass|The compute and memory capacity of the DB instance.|db.r5.large|db.z1d.large, db.z1d.xlarge, db.z1d.2xlarge, db.z1d.3xlarge, db.z1d.6xlarge, db.z1d.12xlarge, db.m5.large, db.m5.2xlarge, db.m5.4xlarge, db.m5.12xlarge, db.m5.24xlarge, db.r5.large, db.r5.xlarge, db.r5.2xlarge, db.r5.4xlarge, db.r5.12xlarge, db.r5.24xlarge, db.x1e.xlarge, db.x1e.4xlarge, db.x1e.8xlarge, db.x1e.16xlarge, db.x1e.32xlarge, db.x1.16xlarge, db.x1.32xlarge, db.r4.large, db.r4.xlarge, db.r4.2xlarge, db.r4.4xlarge, db.r4.8xlarge, db.r4.16xlarge, db.r3.large, db.r3.xlarge, db.r3.2xlarge, db.r3.4xlarge, db.r3.8xlarge, db.m3.medium, db.m3.large, db.m3.xlarge, db.m3.2xlarge, db.m4.large, db.m4.xlarge, db.m4.2xlarge, db.m4.4xlarge, db.m4.10xlarge, db.m4.16xlarge, db.t3.small, db.t3.medium, db.t3.large, db.t3.xlarge, db.t3.2xlarge, db.t2.micro, db.t2.small, db.t2.medium, db.t2.large, db.t2.2xlarge
Engine|Database Engine|Enterprise-Edition-EE-Bring-Your-Own-License|Enterprise-Edition-EE-Bring-Your-Own-License, Standard-Edition-One-SE1-Bring-Your-Own-License, Standard-Edition-Two-SE2-Bring-Your-Own-License, Standard-Edition-SE-Bring-Your-Own-License, Standard-Edition-Two-SE2-License-Included, Standard-Edition-One-SE1-License-Included
EngineVersion|The version number of the database engine to use.|12.1.0.2.v16|12.2.0.1.ru-2019-04.rur-2019-04.r1, 12.2.0.1.ru-2019-01.rur-2019-01.r1, 12.2.0.1.ru-2018-10.rur-2018-10.r1, 12.1.0.2.v16, 12.1.0.2.v15, 12.1.0.2.v14, 12.1.0.2.v13, 12.1.0.2.v12, 12.1.0.2.v11, 12.1.0.2.v10, 11.2.0.4.v20, 11.2.0.4.v19, 11.2.0.4.v18, 11.2.0.4.v17, 11.2.0.4.v16, 11.2.0.4.v15, 11.2.0.4.v14, 11.2.0.4.v13, 11.2.0.4.v12, 11.2.0.4.v11, 11.2.0.4.v10
MonitoringInterval|The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance.|1|0, 1, 5, 10, 15, 30, 60
MultiAZ|Specifies if the database instance is a multiple Availability Zone deployment.|true|true, false
ServerTimezone|The default timezone for the database engine to use.|UTC|Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Monrovia, Africa/Nairobi, Africa/Tripoli, Africa/Windhoek, America/Araguaina, America/Asuncion, America/Bogota, America/Caracas, America/Chihuahua, America/Cuiaba, America/Denver, America/Fortaleza, America/Guatemala, America/Halifax, America/Manaus, America/Matamoros, America/Monterrey, America/Montevideo, America/Phoenix, America/Santiago, America/Tijuana, Asia/Amman, Asia/Ashgabat, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Beirut, Asia/Calcutta, Asia/Damascus, Asia/Dhaka, Asia/Irkutsk, Asia/Jerusalem, Asia/Kabul, Asia/Karachi, Asia/Kathmandu, Asia/Krasnoyarsk, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Taipei, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Vladivostok, Asia/Yakutsk, Asia/Yerevan, Atlantic/Azores, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Perth, Australia/Sydney, Canada/Newfoundland, Canada/Saskatchewan, Brazil/East, Europe/Amsterdam, Europe/Athens, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Sarajevo, Pacific/Auckland, Pacific/Fiji, Pacific/Guam, Pacific/Honolulu, Pacific/Samoa, US/Alaska, US/Central, US/Eastern, US/East-Indiana, US/Pacific, UTC
CharacterSetName|The character set being used by the database|AL32UTF8|AL32UTF8, AR8ISO8859P6, AR8MSWIN1256, BLT8ISO8859P13, BLT8MSWIN1257, CL8ISO8859P5, CL8MSWIN1251, EE8ISO8859P2, EL8ISO8859P7, EE8MSWIN1250, EL8MSWIN1253, IW8ISO8859P8, IW8MSWIN1255, JA16EUC, JA16EUCTILDE, JA16SJIS, JA16SJISTILDE, KO16MSWIN949, NE8ISO8859P10, NEE8ISO8859P4, TH8TISASCII, TR8MSWIN1254, US7ASCII, UTF8, VN8MSWIN1258, WE8ISO8859P1, WE8ISO8859P15, WE8ISO8859P9, WE8MSWIN1252, ZHS16GBK, ZHT16HKSCS, ZHT16MSWIN950, ZHT32EUC
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2|1, 2, 3, 4, 5
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which RDS maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the RDS maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the RDS maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PubliclyAccessible|Indicates whether the DB instance is an Internet-facing instance.|false|true, false
ReadReplica|Number of Read Replicas to create. Only available on  Oracle Enterprise Edition (EE) engine with version 12.1 or higher|0|0, 1, 2
StorageEncrypted|Indicates whether the DB instance is encrypted.|true|true, false
StorageType|Specifies the storage type to be associated with the DB instance.|io1|io1, gp2, standard

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create RDS instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the RDS instance into||

<a id="bind-credentials"></a>

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
EndpointAddress|
MasterUsername|
MasterPassword|
Port|
DBName|

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add them as additional parameters

<a id ="example-production"></a>

## production

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-production-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: production
  parameters: 
    AccessCidr: [VALUE]    
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-production-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: production
  parameters: 
    AccessCidr: [VALUE]
    AllocatedStorageAndIops: 100GB 1000IOPS
    AllowMajorVersionUpgrade: false
    AutoMinorVersionUpgrade: true
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CopyTagsToSnapshot: true
    DBInstanceClass: db.r5.large
    DBName: myorcldb
    Engine: Enterprise-Edition-EE-Bring-Your-Own-License
    EngineVersion: 12.1.0.2.v16
    MasterUserPassword: Auto
    MasterUsername: master
    MonitoringInterval: 1
    MultiAZ: true
    ServerTimezone: UTC
    CharacterSetName: AL32UTF8
    NumberOfAvailabilityZones: 3
    CidrSize: 27
    PortNumber: 1521
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Mon
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
    PubliclyAccessible: false
    ReadReplica: 1
    StorageEncrypted: true
    StorageType: io1    
```


<a id="example-dev"></a>

## dev

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-dev-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: dev
  parameters: 
    AccessCidr: [VALUE]    
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-dev-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: dev
  parameters: 
    AccessCidr: [VALUE]
    AllocatedStorageAndIops: 100GB 1000IOPS
    AllowMajorVersionUpgrade: false
    AutoMinorVersionUpgrade: true
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CopyTagsToSnapshot: true
    DBInstanceClass: db.r5.large
    DBName: myorcldb
    Engine: Enterprise-Edition-EE-Bring-Your-Own-License
    EngineVersion: 12.1.0.2.v16
    MasterUserPassword: Auto
    MasterUsername: [VALUE]
    MonitoringInterval: 60
    MultiAZ: false
    ServerTimezone: UTC
    CharacterSetName: AL32UTF8
    NumberOfAvailabilityZones: 2
    CidrSize: 27
    PortNumber: 1521
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Mon
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
    PubliclyAccessible: false
    ReadReplica: 0
    StorageEncrypted: true
    StorageType: io1
```


<a id = "example-custom"></a>

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-custom-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: rdsoracle-custom-complete-example
spec: 
  clusterServiceClassExternalName: rdsoracle
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
    AllocatedStorageAndIops: 100GB 1000IOPS
    AllowMajorVersionUpgrade: false
    AutoMinorVersionUpgrade: true
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CopyTagsToSnapshot: true
    DBInstanceClass: db.r5.large
    DBName: myorcldb
    Engine: Enterprise-Edition-EE-Bring-Your-Own-License
    EngineVersion: 12.1.0.2.v16
    MasterUserPassword: Auto
    MasterUsername: [VALUE]
    MonitoringInterval: 1
    MultiAZ: true
    ServerTimezone: UTC
    CharacterSetName: AL32UTF8
    NumberOfAvailabilityZones: 2
    CidrSize: 27
    PortNumber: 1521
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Mon
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
    PubliclyAccessible: false
    ReadReplica: 0
    StorageEncrypted: true
    StorageType: io1    
```


