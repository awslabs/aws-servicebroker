# AWS Service Broker - Amazon Elasticsearch

<img align="left" src="https://s3.amazonaws.com/awsservicebroker/icons/aws-service-broker.png" width="120"><p align="center">Amazon Elasticsearch Service is a fully managed service that makes it easy for you to deploy, secure, and operate Elasticsearch at scale with zero down time. The service offers open-source Elasticsearch APIs, managed Kibana, and integrations with Logstash and other AWS Services, enabling you to securely ingest data from any source and search, analyze, and visualize it in real time. Amazon Elasticsearch Service lets you pay only for what you use – there are no upfront costs or usage requirements. With Amazon Elasticsearch Service, you get the ELK stack you need, without the operational overhead.</p>&nbsp;

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
## production parameters

Creates an Amazon Elasticsearch optimised for production use.  
Pricing: https://aws.amazon.com/elasticsearch-service/pricing/?nc=sn&loc=3

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to ElasticSearch Cluster|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
EngineVersion|Version of Elasticsearch|6.7|6.7, 6.5, 6.4, 6.3, 6.2, 6.0, 5.6, 5.5, 5.3, 5.1, 2.3, 1.5
NodeAvailabilityZones|Number of AZs. For two AZs, you must choose instances in multiples of two. For three AZs, we recommend instances in multiples of three for equal distribution across the Availability Zones.|2|1, 2
NodeInstanceType|The node type to be provisioned for the Elasticsearch cluster|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
DedicatedMasterInstanceType|Master Instance Type|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
DedicatedMasterInstanceCount|The number of dedicated master nodes (instances) to use in the ES domain (set to 0 to disable dedicated master nodes).|3|0, 3, 5
StorageType|Specifies the storage type to be associated to the Data Nodes|io1|io1, gp2, standard
AllocatedStorageAndIops|Storage/IOPS to allocate. Total cluster size will be (EBS volume size x Instance count).|100GB 1000IOPS|100GB 1000IOPS, 300GB 3000IOPS, 600GB 6000IOPS, 1000GB 10000IOPS
PreferredSnapshotTime|The hour in UTC during which the service takes an automated daily snapshot of the indices in the Amazon ES domain|0|0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23
NodeToNodeEncryption|Specifies whether node-to-node encryption is enabled.|true|true, false
EncryptionAtRest|Whether the domain should encrypt data at rest, and if so, the AWS Key Management Service (KMS) key to use. Can only be used to create a new domain, not update an existing one.|true|true, false

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Cluster into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NodeAvailabilityZones|Auto
CidrBlocks|comma seperated list of CIDR blocks to place into, must be the same quantity as specified in NodeAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
EngineVersion|Version of Elasticsearch|6.7
ESDomainName|A name for the Amazon ES domain, will be autogenerated if set to "Auto".|Auto
NodeAvailabilityZones|Number of AZs. For two AZs, you must choose instances in multiples of two. For three AZs, we recommend instances in multiples of three for equal distribution across the Availability Zones.|2
NodeInstanceType|The node type to be provisioned for the Elasticsearch cluster|r5.large.elasticsearch
NodeInstanceCount|Number of Data Nodes for the ES Domain|2
DedicatedMasterInstanceType|Master Instance Type|r5.large.elasticsearch
DedicatedMasterInstanceCount|The number of dedicated master nodes (instances) to use in the ES domain (set to 0 to disable dedicated master nodes).|3
StorageType|Specifies the storage type to be associated to the Data Nodes|io1
AllocatedStorageAndIops|Storage/IOPS to allocate. Total cluster size will be (EBS volume size x Instance count).|100GB 1000IOPS
PreferredSnapshotTime|The hour in UTC during which the service takes an automated daily snapshot of the indices in the Amazon ES domain|0
NodeToNodeEncryption|Specifies whether node-to-node encryption is enabled.|true
EncryptionAtRest|Whether the domain should encrypt data at rest, and if so, the AWS Key Management Service (KMS) key to use. Can only be used to create a new domain, not update an existing one.|true

<a id = "param-dev"></a>
## Dev Parameters

Creates an Amazon Elasticsearch optimised for dev/test use.  
Pricing: https://aws.amazon.com/elasticsearch-service/pricing/?nc=sn&loc=3

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to ElasticSearch Cluster|string

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
EngineVersion|Version of Elasticsearch|6.7|6.7, 6.5, 6.4, 6.3, 6.2, 6.0, 5.6, 5.5, 5.3, 5.1, 2.3, 1.5
NodeAvailabilityZones|Number of AZs. For two AZs, you must choose instances in multiples of two. For three AZs, we recommend instances in multiples of three for equal distribution across the Availability Zones.|1|1, 2
NodeInstanceType|The node type to be provisioned for the Elasticsearch cluster|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
DedicatedMasterInstanceType|Master Instance Type|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
DedicatedMasterInstanceCount|The number of dedicated master nodes (instances) to use in the ES domain (set to 0 to disable dedicated master nodes).|0|0, 3, 5
StorageType|Specifies the storage type to be associated to the Data Nodes|io1|io1, gp2, standard
AllocatedStorageAndIops|Storage/IOPS to allocate. Total cluster size will be (EBS volume size x Instance count).|100GB 1000IOPS|100GB 1000IOPS, 300GB 3000IOPS, 600GB 6000IOPS, 1000GB 10000IOPS
PreferredSnapshotTime|The hour in UTC during which the service takes an automated daily snapshot of the indices in the Amazon ES domain|0|0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23
NodeToNodeEncryption|Specifies whether node-to-node encryption is enabled.|true|true, false
EncryptionAtRest|Whether the domain should encrypt data at rest, and if so, the AWS Key Management Service (KMS) key to use. Can only be used to create a new domain, not update an existing one.|true|true, false

### Generic

These parameters are required, but generic or require privileged access to the underlying AWS account, we recommend they are configured with a broker secret, see [broker documentation](/docs/) for details.

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
target_account_id | AWS Account ID to provision into(optional) ||
target_role_name | IAM Role name to provision with(optional), must be used in combination with target_account_id ||
region | AWS Region to create instance in.| us-west-2 | ap-northeast-1, ap-northeast-2, ap-south-1, ap-southeast-1, ap-southeast-2, ca-central-1, eu-central-1, eu-west-1, eu-west-2, sa-east-1, us-east-1, us-east-2, us-west-1, us-west-2
VpcId|The ID of the VPC to launch the Cluster into||

### Prescribed

These are parameters that are prescribed by the plan and are not configurable, should adjusting any of these be required please choose a plan that makes them available.

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NodeAvailabilityZones|Auto
CidrBlocks|comma seperated list of CIDR blocks to place into, must be the same quantity as specified in NodeAvailabilityZones|Auto
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27
EngineVersion|Version of Elasticsearch|6.7
ESDomainName|A name for the Amazon ES domain, will be autogenerated if set to "Auto".|Auto
NodeAvailabilityZones|Number of AZs. For two AZs, you must choose instances in multiples of two. For three AZs, we recommend instances in multiples of three for equal distribution across the Availability Zones.|2
NodeInstanceType|The node type to be provisioned for the Elasticsearch cluster|r5.large.elasticsearch
NodeInstanceCount|Number of Data Nodes for the ES Domain|2
DedicatedMasterInstanceType|Master Instance Type|r5.large.elasticsearch
DedicatedMasterInstanceCount|The number of dedicated master nodes (instances) to use in the ES domain (set to 0 to disable dedicated master nodes).|3
StorageType|Specifies the storage type to be associated to the Data Nodes|io1
AllocatedStorageAndIops|Storage/IOPS to allocate. Total cluster size will be (EBS volume size x Instance count).|100GB 1000IOPS
PreferredSnapshotTime|The hour in UTC during which the service takes an automated daily snapshot of the indices in the Amazon ES domain|0
NodeToNodeEncryption|Specifies whether node-to-node encryption is enabled.|true
EncryptionAtRest|Whether the domain should encrypt data at rest, and if so, the AWS Key Management Service (KMS) key to use. Can only be used to create a new domain, not update an existing one.|true

<a id = "param-custom"></a>
## Custom parameters

Creates an Amazon Elasticsearch with custom configuration.  
Pricing: https://aws.amazon.com/elasticsearch-service/pricing/?nc=sn&loc=3

### Required

At a minimum these parameters must be declared when provisioning an instance of this service

Name           | Description     | Accepted Values
-------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to ElasticSearch Cluster|

### Optional

These parameters can optionally be declared when provisioning

Name            | Description     | Default         | Accepted Values
--------------- | --------------- | --------------- | ---------------
AccessCidr|CIDR block to allow to connect to ElasticSearch Cluster||
VpcId|The ID of the VPC to launch the ElasticSearch instance into||
AvailabilityZones|list of availability zones to use, must be the same quantity as specified in NodeAvailabilityZones|Auto|
CidrBlocks|comma seperated list of CIDR blocks to place into, must be the same quantity as specified in NodeAvailabilityZones|Auto|
CidrSize|Size of Cidr block to allocate if CidrBlocks is set to Auto.|27|
EngineVersion|Version of Elasticsearch|6.7|6.7, 6.5, 6.4, 6.3, 6.2, 6.0, 5.6, 5.5, 5.3, 5.1, 2.3, 1.5
ESDomainName|A name for the Amazon ES domain, will be autogenerated if set to "Auto".|Auto|
NodeAvailabilityZones|Number of AZs. For two AZs, you must choose instances in multiples of two. For three AZs, we recommend instances in multiples of three for equal distribution across the Availability Zones.|2|1, 2
NodeInstanceType|The node type to be provisioned for the Elasticsearch cluster|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
NodeInstanceCount|Number of Data Nodes for the ES Domain|2|
DedicatedMasterInstanceType|Master Instance Type|r5.large.elasticsearch|t2.micro.elasticsearch, t2.small.elasticsearch, t2.medium.elasticsearch, m5.large.elasticsearch, m5.xlarge.elasticsearch, m5.2xlarge.elasticsearch, m5.4xlarge.elasticsearch, m5.12xlarge.elasticsearch, m4.large.elasticsearch, m4.xlarge.elasticsearch, m4.2xlarge.elasticsearch, m4.4xlarge.elasticsearch, m4.10xlarge.elasticsearch, c5.large.elasticsearch, c5.xlarge.elasticsearch, c5.2xlarge.elasticsearch, c5.4xlarge.elasticsearch, c5.9xlarge.elasticsearch, c5.18xlarge.elasticsearch, c4.large.elasticsearch, c4.xlarge.elasticsearch, c4.2xlarge.elasticsearch, c4.4xlarge.elasticsearch, c4.8xlarge.elasticsearch, r5.large.elasticsearch, r5.xlarge.elasticsearch, r5.2xlarge.elasticsearch, r5.4xlarge.elasticsearch, r5.12xlarge.elasticsearch, r4.large.elasticsearch, r4.xlarge.elasticsearch, r4.2xlarge.elasticsearch, r4.4xlarge.elasticsearch, r4.8xlarge.elasticsearch, r4.16xlarge.elasticsearch, i3.large.elasticsearch, i3.xlarge.elasticsearch, i3.2xlarge.elasticsearch, i3.4xlarge.elasticsearch, i3.8xlarge.elasticsearch, i3.16xlarge.elasticsearch
DedicatedMasterInstanceCount|The number of dedicated master nodes (instances) to use in the ES domain (set to 0 to disable dedicated master nodes).|3|0, 3, 5
StorageType|Specifies the storage type to be associated to the Data Nodes|io1|io1, gp2, standard
AllocatedStorageAndIops|Storage/IOPS to allocate. Total cluster size will be (EBS volume size x Instance count).|100GB 1000IOPS|100GB 1000IOPS, 300GB 3000IOPS, 600GB 6000IOPS, 1000GB 10000IOPS
PreferredSnapshotTime|The hour in UTC during which the service takes an automated daily snapshot of the indices in the Amazon ES domain|0|0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23
NodeToNodeEncryption|Specifies whether node-to-node encryption is enabled.|true|true, false
EncryptionAtRest|Whether the domain should encrypt data at rest, and if so, the AWS Key Management Service (KMS) key to use. Can only be used to create a new domain, not update an existing one.|true|true, false

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
ELASTICSEARCH_DOMAIN_ARN | The ARN of the Elasticsearch domain.
ELASTICSEARCH_ENDPOINT | The endpoint address of the Elasticsearch cluster.
ES_DOMAIN_NAME | The Elasticsearch domain name.

# Kubernetes/Openshift Examples

***Note:*** Examples do not include generic parameters, if you have not setup defaults for these you will need to add them as additional parameters

<a id ="example-production"></a>

## production

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-production-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: production
  parameters: 
    AccessCidr: [VALUE]    
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-production-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: production
  parameters:    
    AvailabilityZones: Auto
    CidrBlocks: Auto
    CidrSize: 27
    EngineVersion: 6.7
    ESDomainName: Auto
    NodeAvailabilityZones: 2
    NodeInstanceType: r5.large.elasticsearch
    NodeInstanceCount: 2
    DedicatedMasterInstanceType: r5.large.elasticsearch
    DedicatedMasterInstanceCount: 3    
    PreferredSnapshotTime: 0
    NodeToNodeEncryption: true
    EncryptionAtRest: true
```


<a id="example-dev"></a>

## dev

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-dev-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: dev
  parameters: 
    AccessCidr: [VALUE]
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-dev-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: dev
  parameters:   
    CidrSize: 27
    EngineVersion: 6.7
    ESDomainName: Auto
    NodeAvailabilityZones: 1
    NodeInstanceType: r5.large.elasticsearch
    NodeInstanceCount: 1    
    DedicatedMasterInstanceCount: 0    
    PreferredSnapshotTime: 0
    NodeToNodeEncryption: true
    EncryptionAtRest: true
```


<a id = "example-custom"></a>

## custom

### Minimal
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-custom-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
```


### Complete
```yaml
apiVersion: servicecatalog.k8s.io / v1beta1
kind: ServiceInstance
metadata: 
  name: elasticsearch-custom-complete-example
spec: 
  clusterServiceClassExternalName: elasticsearch
  clusterServicePlanExternalName: custom
  parameters: 
    AccessCidr: [VALUE]
    VpcId: [VALUE]
    AvailabilityZones: Auto
    CidrBlocks: Auto
    CidrSize: 27
    EngineVersion: 6.7
    ESDomainName: Auto
    NodeAvailabilityZones: 2
    NodeInstanceType: r5.large.elasticsearch
    NodeInstanceCount: 2
    DedicatedMasterInstanceType: r5.large.elasticsearch
    DedicatedMasterInstanceCount: 3
    StorageType: io1
    AllocatedStorageAndIops: 100GB 1000IOPS
    PreferredSnapshotTime: 0
    NodeToNodeEncryption: true
    EncryptionAtRest: true
```


