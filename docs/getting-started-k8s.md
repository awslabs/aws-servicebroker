# Getting Started Guide - Kubernetes

[Blog post detailing installation and usage on Kubernetes](https://aws.amazon.com/blogs/opensource/provision-aws-services-kubernetes-aws-service-broker/)

## Installing the AWS Service Broker in Kubernetes

*Note:* The use of the AWS Service Broker in Kubernetes is still in early beta, bugs and possible update related breaking changes may manifest. Use of the AWS Service Broker in Kubernetes is not recommended for production at this time.

### Prerequisites

### Kubernetes Service Catalog

A simple way to install the Service Catalog is to use helm, for documentation on installing the helm client see https://docs.helm.sh/using_helm/#install-helm:


```bash
# Install helm and tiller into the cluster
helm init
# Wait until tiller is ready before moving onuntil kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done

kubectl create clusterrolebinding tiller-cluster-admin \
    --clusterrole=cluster-admin \
    --serviceaccount=kube-system:default
# Adds the chart repository for the service catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
# Installs the service catalog
helm install svc-cat/catalog --name catalog --namespace catalog
```

### Installation

* ssh into host with access to the target kubernetes cluster and ensure all prerequisites are in place
* download deployment files

```bash
mkdir aws_service_broker/
cd aws_service_broker/
wget https://github.com/awslabs/aws-servicebroker/raw/master/scripts/k8s_deployment/install_aws_service_broker.sh
wget https://github.com/awslabs/aws-servicebroker/raw/master/scripts/k8s_deployment/k8s-aws-service-broker.yaml.j2
wget https://github.com/awslabs/aws-servicebroker/raw/master/scripts/k8s_deployment/k8s-template.py
wget https://github.com/awslabs/aws-servicebroker/raw/master/scripts/k8s_deployment/k8s-variables.yaml
```

* edit the k8s-variables.yaml config file and ensure that the AWS Account settings are all set correctly for your account

```yaml
# AWS Account Settings
# If your cluster is not on ec2 with a suitable IAM role enabled then you can supply IAM user credentials here,
# if use-role is specified the broker will attempt to use the EC2 instance profile associated with the node the pod is
# running on. required IAM permissions are documented here:
# https://github.com/awslabs/aws-servicebroker-documentation/blob/master/Overview.md
aws_access_key_id: use-role
aws_secret_access_key: use-role

# A CloudFormation IAM service role is required for the broker to launch AWS services, this role must have all the
# permissions required to manage the resources supported by the broker, as well as the ability to manage IAM users,
# vpc resources and kms keys
aws_cloudformation_role_arn:

# specify an aws region that resources should be launched into
region: us-west-2

# specify a VPC ID, resources that are inside VPC will launch into this VPC, it is recommended you use the same VPC
# that the cluster is in if your cluster is running in ec2
vpc_id:
```

* execute the installer

```bash
chmod +x install_aws_service_broker.sh
./install_aws_service_broker.sh
```
