# AWS Service Broker - Amazon DocumentDB (with MongoDB compatibility)

<img align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src=https://s3.amazonaws.com/awsservicebroker/icons/AmazonRDS_LARGE.png width="108"><p align="center">Amazon DocumentDB (with MongoDB compatibility) is a fast, scalable, highly available, and fully managed document database service that supports MongoDB workloads.</p>&nbsp;

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

Creates an Amazon DocumentDB (with MongoDB compatibility) optimised for production use.  
Pricing: https://aws.amazon.com/documentdb/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
DBInstanceClass|Database Instance Class|db.r5.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which Cluster maintenance will be performed|Sun|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the Cluster maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the Cluster maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create Cluster instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Cluster instance into|

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available.|Auto
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|35
CidrBlocks|comma seperated list of CIDR blocks to place Cluster into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
DBPort|TCP/IP Port for the Database Instance|27017
DBUsername|Database master username|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|3
NumberofReplicas|Number of Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|2
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
StorageEncrypted|Indicates whether the DB instance is encrypted.|true

<a id = "param-dev"></a>

Creates an Amazon DocumentDB (with MongoDB compatibility) optimised for dev/test use.  
Pricing: https://aws.amazon.com/documentdb/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
DBInstanceClass|Database Instance Class|db.r5.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which Cluster maintenance will be performed|Sun|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the Cluster maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the Cluster maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create Cluster instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Cluster instance into|

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available.|Auto
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|0
CidrBlocks|comma seperated list of CIDR blocks to place Cluster into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
DBPort|TCP/IP Port for the Database Instance|27017
DBUsername|Database master username|master
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
NumberofReplicas|Number of Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|0
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true
StorageEncrypted|Indicates whether the DB instance is encrypted.|true
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00
PreferredMaintenanceWindowDay|The day of the week which Cluster maintenance will be performed|Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the Cluster maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the Cluster maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00

<a id = "param-custom"></a>

Creates an Amazon DocumentDB (with MongoDB compatibility) with custom configuration.  
Pricing: https://aws.amazon.com/documentdb/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|
MasterUsername|Master database Username|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database||
AvailabilityZones|list of availability zones to use, must be the same quantity as specified. Leave as Auto for stack to determine AZ names available.|Auto|
BackupRetentionPeriod|The number of days during which automatic DB snapshots are retained. Min is 1 and Max value is 35.|35|
CidrBlocks|comma seperated list of CIDR blocks to place Cluster into, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27|
DBPort|TCP/IP Port for the Database Instance|27017|
DBUsername|Database master username|master|
DBPassword|Master user database Password, if left at default a 32 character password will be generated|Auto|
DBEngineVersion|Select Database Engine Version|3.6.0|3.6.0
DBInstanceClass|Database Instance Class|db.r5.large|db.r4.16xlarge, db.r4.8xlarge, db.r4.4xlarge, db.r4.2xlarge, db.r4.xlarge, db.r4.large, db.r5.24xlarge, db.r5.12xlarge, db.r5.4xlarge, db.r5.2xlarge, db.r5.xlarge, db.r5.large
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2|2, 3, 4, 5
NumberofReplicas|Number of Replicas to deploy in addition to the Primary. If selecting 2 replicas, 3 are required in Number Of Availability Zones.|0|0, 1, 2
VpcId|The ID of the VPC to launch the Cluster instance into||
AutoMinorVersionUpgrade|Indicates that minor engine upgrades are applied automatically to the DB instance during the maintenance window.|true|true, false
StorageEncrypted|Indicates whether the DB instance is encrypted.|true|true, false
PreferredBackupWindow|The daily time range in UTC during which automated backups are created (if automated backups are enabled). Cannot overlap with PreferredMaintenanceWindowTime|00:00-02:00|00:00-02:00, 01:00-03:00, 02:00-04:00, 03:00-05:00, 04:00-06:00, 05:00-07:00, 06:00-08:00, 07:00-09:00, 08:00-10:00, 09:00-11:00, 10:00-12:00, 11:00-13:00, 12:00-14:00, 13:00-15:00, 14:00-16:00, 15:00-17:00, 16:00-18:00, 17:00-19:00, 18:00-20:00, 19:00-21:00, 20:00-22:00, 21:00-23:00, 22:00-24:00
PreferredMaintenanceWindowDay|The day of the week which Cluster maintenance will be performed|Sun|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the Cluster maintenance window, must be more than PreferredMaintenanceWindowStartTime.|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the Cluster maintenance window, must be less than PreferredMaintenanceWindowEndTime.|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create Cluster instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Cluster instance into|

<a id="bind-credentials"></a>

# Bind Credentials

These are the environment variables that are available to an application on bind.

Name           | Description
-------------- | ---------------
ENDPOINT_ADDRESS|
MASTER_USERNAME|
MASTER_PASSWORD|
PORT|

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add them as additional parameters

<a id ="example-production"></a>

## production

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-production-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: production
  parameters: 
    AccessCidr: [VALUE]
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-production-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: production
  parameters: 
    AccessCidr: [VALUE]
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CidrSize: 27
    DBPort: 27017
    DBUsername: master
    DBPassword: Auto
    DBEngineVersion: 3.6.0
    DBInstanceClass: db.r5.large
    NumberOfAvailabilityZones: 3
    NumberofReplicas: 2  
    AutoMinorVersionUpgrade: true
    StorageEncrypted: true
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Sun
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
```


<a id="example-dev"></a>

## dev

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-dev-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: dev
  parameters: 
    AccessCidr: [VALUE]
  
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-dev-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: dev
  parameters: 
    AccessCidr: [VALUE]
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CidrSize: 27
    DBPort: 27017
    DBUsername: master
    DBPassword: Auto
    DBEngineVersion: 3.6.0
    DBInstanceClass: db.r5.large
    NumberOfAvailabilityZones: 2
    NumberofReplicas: 0    
    AutoMinorVersionUpgrade: true
    StorageEncrypted: true
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Sun
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
```


<a id = "example-custom"></a>

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-custom-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
    MasterUsername: [VALUE] # REQUIRED
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: documentdb-custom-complete-example
spec: 
  clusterServiceClassExternalName: documentdb
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
    AvailabilityZones: Auto
    BackupRetentionPeriod: 35
    CidrBlocks: Auto
    CidrSize: 27
    DBPort: 27017
    DBUsername: master
    DBPassword: Auto
    DBEngineVersion: 3.6.0
    DBInstanceClass: db.r5.large
    NumberOfAvailabilityZones: 2
    NumberofReplicas: 0    
    AutoMinorVersionUpgrade: true
    StorageEncrypted: true
    PreferredBackupWindow: 00:00-02:00
    PreferredMaintenanceWindowDay: Sun
    PreferredMaintenanceWindowEndTime: 06:00
    PreferredMaintenanceWindowStartTime: 04:00
```


