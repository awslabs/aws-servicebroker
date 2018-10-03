# AWS Service Broker - Amazon ElastiCache for memcached Documentation

<img  align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><img align="right" src="https://s3.amazonaws.com/thp-aws-icons-dev/Database_AmazonElasticCache_LARGE.png" width="108"> <p align="center">Amazon ElastiCache is a web service that makes it easy to set up, manage, and scale distributed in-memory cache environments in the cloud. It provides a high performance, resizeable, and cost-effective in-memory cache, while removing the complexity associated with deploying and managing a distributed cache environment.
https://aws.amazon.com/documentation/elasticache/</p>

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

Creates an Amazon ElastiCache for memcached, optimised for production use

Pricing: https://aws.amazon.com/elasticache/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
PreferredMaintenanceWindowDay|The day of the week which ElastiCache maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the ElastiCache maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the ElastiCache maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
CacheNodeType|The compute and memory capacity of nodes in a cache cluster.|cache.m4.large|cache.t2.small, cache.t2.medium, cache.m3.medium, cache.m3.large, cache.m3.xlarge, cache.m3.2xlarge, cache.m4.medium, cache.m4.large, cache.m4.xlarge, cache.m4.2xlarge, cache.m4.4xlarge, cache.m4.10xlarge, cache.r4.large, cache.r4.xlarge, cache.r4.2xlarge, cache.r4.4xlarge, cache.r4.8xlarge
EngineVersion|Family to be used with cluster or parameter group|1.4.34|1.4.34, 1.4.33, 1.4.24, 1.4.5
NumCacheNodes|The number of cache nodes in the cluster.|3|

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Memcache cluster into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Value
-------------- | --------------- | ---------------
ClusterType|The type of cluster. Specify single-node or multi-node (default).  Number of nodes must be greater than 1 for multi-node|multi-node
AllowVersionUpgrade|Indicates that minor engine upgrades will be applied automatically to the cache cluster during the maintenance window. The default value is true.|False
PortNumber|The port number for the Cluster to listen on|6379
AZMode|Specifies whether the nodes in this Memcached cluster are created in a single Availability Zone or created across multiple Availability Zones in the cluster's region. This parameter is only supported for Memcached cache clusters. If the AZMode and PreferredAvailabilityZones are not specified, ElastiCache assumes single-az mode.|cross-az
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|2
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto
CidrBlocks|comma seperated list of CIDR blocks to place ElastiCache into, must be the same quantity as specified in NumberOfAvailabilityZones. If auto is specified unused cidr space in the vpc will be used|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|26
<a id="param-custom" />

## custom

Creates an Amazon ElastiCache for memcached with custom configuration

Pricing: https://aws.amazon.com/elasticache/pricing/

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to database|string

### Optional

These parameters can optionally be declared when provisioning

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
NumberOfAvailabilityZones|Quantity of subnets to use, if selecting more than 2 the region this stack is in must have at least that many Availability Zones|3|2, 3, 4, 5
PreferredMaintenanceWindowDay|The day of the week which ElastiCache maintenance will be performed|Mon|Mon, Tue, Wed, Thu, Fri, Sat, Sun
PreferredMaintenanceWindowStartTime|The weekly start time in UTC for the ElastiCache maintenance window, must be less than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|04:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
PreferredMaintenanceWindowEndTime|The weekly end time in UTC for the ElastiCache maintenance window, must be more than PreferredMaintenanceWindowEndTime and cannot overlap with PreferredBackupWindow|06:00|00:00, 01:00, 02:00, 03:00, 04:00, 05:00, 06:00, 07:00, 08:00, 09:00, 10:00, 11:00, 12:00, 13:00, 14:00, 15:00, 16:00, 17:00, 18:00, 19:00, 20:00, 21:00, 22:00
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NumberOfAvailabilityZones|Auto|
CidrBlocks|comma seperated list of CIDR blocks to place ElastiCache into, must be the same quantity as specified in NumberOfAvailabilityZones. If auto is specified unused cidr space in the vpc will be used|Auto|
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|26|
CacheNodeType|The compute and memory capacity of nodes in a cache cluster.|cache.m4.large|cache.t2.small, cache.t2.medium, cache.m3.medium, cache.m3.large, cache.m3.xlarge, cache.m3.2xlarge, cache.m4.medium, cache.m4.large, cache.m4.xlarge, cache.m4.2xlarge, cache.m4.4xlarge, cache.m4.10xlarge, cache.r4.large, cache.r4.xlarge, cache.r4.2xlarge, cache.r4.4xlarge, cache.r4.8xlarge
EngineVersion|Family to be used with cluster or parameter group|1.4.34|1.4.34, 1.4.33, 1.4.24, 1.4.5
NumCacheNodes|The number of cache nodes in the cluster.|3|
ClusterType|The type of cluster. Specify single-node or multi-node (default).  Number of nodes must be greater than 1 for multi-node|multi-node|single-node, multi-node
AllowVersionUpgrade|Indicates that minor engine upgrades will be applied automatically to the cache cluster during the maintenance window. The default value is true.|True|True, False
PortNumber|The port number for the Cluster to listen on|5439|
AZMode|Specifies whether the nodes in this Memcached cluster are created in a single Availability Zone or created across multiple Availability Zones in the cluster's region. This parameter is only supported for Memcached cache clusters. If the AZMode and PreferredAvailabilityZones are not specified, ElastiCache assumes single-az mode.|cross-az|single-az, cross-az

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name           | Description     | Default         | Accepted Values
-------------- | --------------- | --------------- | ---------------
target_account_id|AWS Account ID to provision into (optional)||
target_role_name|IAM Role name to provision with (optional), must be used in combination with target_account_id||
region|AWS Region to create RDS instance in.|us-west-2|ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Memcache cluster into||

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
  name: elasticache-production-minimal-example
spec:
  clusterServiceClassExternalName: elasticache
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: elasticache-production-complete-example
spec:
  clusterServiceClassExternalName: elasticache
  clusterServicePlanExternalName: production
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    CacheNodeType: cache.m4.large # OPTIONAL
    EngineVersion: 1.4.34 # OPTIONAL
    NumCacheNodes: 3 # OPTIONAL
```
<a id="example-custom" />

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: elasticache-custom-minimal-example
spec:
  clusterServiceClassExternalName: elasticache
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
```

### Complete
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: elasticache-custom-complete-example
spec:
  clusterServiceClassExternalName: elasticache
  clusterServicePlanExternalName: custom
  parameters:
    AccessCidr: [VALUE] # REQUIRED
    NumberOfAvailabilityZones: 3 # OPTIONAL
    PreferredMaintenanceWindowDay: Mon # OPTIONAL
    PreferredMaintenanceWindowStartTime: 04:00 # OPTIONAL
    PreferredMaintenanceWindowEndTime: 06:00 # OPTIONAL
    AvailabilityZones: Auto # OPTIONAL
    CidrBlocks: Auto # OPTIONAL
    CidrSize: 26 # OPTIONAL
    CacheNodeType: cache.m4.large # OPTIONAL
    EngineVersion: 1.4.34 # OPTIONAL
    NumCacheNodes: 3 # OPTIONAL
    ClusterType: multi-node # OPTIONAL
    AllowVersionUpgrade: True # OPTIONAL
    PortNumber: 5439 # OPTIONAL
    AZMode: cross-az # OPTIONAL
```

