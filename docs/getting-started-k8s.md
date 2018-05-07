# Getting Started Guide - Kubernetes

[Blog post detailing installation and usage on Kubernetes](https://aws.amazon.com/blogs/opensource/provision-aws-services-kubernetes-aws-service-broker/)

## Installing the AWS Service Broker in Kubernetes

*Note:* The use of the AWS Service Broker in Kubernetes is still in early beta, bugs and possible update related breaking changes may manifest. Use of the AWS Service Broker in Kubernetes is not recommended for production at this time.

### Prerequisites

### Kubernetes v1.9

Testing on V1.9 was done using the [Quick Start for Kubernetes by Heptio on AWS](https://aws.amazon.com/quickstart/architecture/heptio-kubernetes/). Though not tested other kubernetes versions may work.

kubectl must be installed on a linux or osx machine with access to the kubernetes cluster, using one of the master nodes to execute the scripts is a simple way of ensuring the appropriate network access.
kubectl also needs to have a current-context setup with a user who has cluster-admin rights.

### Kubernetes Service Catalog

A simple way to install the Service Catalog is to use helm, for documentation on installing the helm client see https://docs.helm.sh/using_helm/#install-helm:


```bash
# Install helm and tiller into the cluster
helm init
# Wait until tiller is ready before moving onuntil kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done

kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
# Adds the chart repository for the service catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
# Installs the service catalog
helm install svc-cat/catalog --name catalog --namespace catalog
```

### IAM Roles/Users

The AWS Service Broker packages all services into CloudFormation templates that are executed by the broker. The broker can use a role if the cluster is running on EC2, and the pods on the cluster have access to the ec2 metadata endpoint (169.254.169.254), or an IAM user and static keypair can be created for the broker to use. Either way this entity requires the following IAM policy:


```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "cloudformation:CancelUpdateStack",
                "cloudformation:ContinueUpdateRollback",
                "cloudformation:CreateStack",
                "cloudformation:CreateUploadBucket",
                "cloudformation:DeleteStack",
                "cloudformation:DescribeAccountLimits",
                "cloudformation:DescribeStackEvents",
                "cloudformation:DescribeStackResource",
                "cloudformation:DescribeStackResources",
                "cloudformation:DescribeStacks",
                "cloudformation:GetStackPolicy",
                "cloudformation:ListStackResources",
                "cloudformation:ListStacks",
                "cloudformation:SetStackPolicy",
                "cloudformation:UpdateStack",
                "iam:AddUserToGroup",
                "iam:AttachUserPolicy",
                "iam:CreateAccessKey",
                "iam:CreatePolicy",
                "iam:CreatePolicyVersion",
                "iam:CreateUser",
                "iam:DeleteAccessKey",
                "iam:DeletePolicy",
                "iam:DeletePolicyVersion",
                "iam:DeleteRole",
                "iam:DeleteUser",
                "iam:DeleteUserPolicy",
                "iam:DetachUserPolicy",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:GetUser",
                "iam:GetUserPolicy",
                "iam:ListAccessKeys",
                "iam:ListGroups",
                "iam:ListGroupsForUser",
                "iam:ListInstanceProfiles",
                "iam:ListPolicies",
                "iam:ListPolicyVersions",
                "iam:ListRoles",
                "iam:ListUserPolicies",
                "iam:ListUsers",
                "iam:PutUserPolicy",
                "iam:RemoveUserFromGroup",
                "iam:UpdateUser",
                "ec2:DescribeVpcs",
                "ec2:DescribeSubnets",
                "ec2:DescribeAvailabilityZones"
            ],
            "Resource": [
                "*"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "iam:PassRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/AWSServiceBrokerCFNRole"
            ],
            "Effect": "Allow"
        },
        {
            "Action": [
                "ssm:GetParameters"
            ],
            "Resource": [
                "arn:aws:ssm:*:*:parameter/asb-access-key-id-*",
                "arn:aws:ssm:*:*:parameter/asb-secret-access-key-*"
            ],
            "Effect": "Allow"
        }
    ]
}
```

*Note:* the ec2 and iam permissions (with the exception of PassRole) will not be needed in the near future, so the permissions will be able to be scoped down, but for the time being are a requirement to launch many of the services.
A [CloudFormation service role](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-iam-servicerole.html) is also required, an example of a broad policy to enable all broker functionality is included below:


```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "cloudformation:*",
                "iam:*",
                "kms:*",
                "ssm:*",
                "ec2:*",
                "lambda:*",
                "athena:*",
                "dynamodb:*",
                "elasticache:*",
                "emr:*",
                "rds:*",
                "redshift:*",
                "route53:*",
                "s3:*",
                "sns:*",
                "sqs:*"
            ],
            "Resource": [
                "*"
            ],
            "Effect": "Allow"
        }
    ]
}
```

*Note:* the first policy expects the cloudformation service role to be named AWSServiceBrokerCFNRole, if you name it something else you will also need to update this policy

### Python

Python 2.7 or 3.5+ is required by the setup scripts.
In addition the jinja2 python package is required, it can be easily installed with pip:


```bash
pip install jinja2
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
