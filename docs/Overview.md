# AWS Service Broker Overview
The AWS Service Broker is a Service Broker that provides access to AWS services within the Service Catalog on RedHat OpenShift Container Platform. The broker implementation is based on the [Ansible Service Broker](https://github.com/openshift/ansible-service-broker/blob/master/docs/introduction.md).

![Architecture](/images/architecture.png)

## Requirements:
**OpenShift** (Origin or Enterprise) >= 3.7  
**[Kubernetes Service Catalog](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/design.md)** >= 0.1.0  
**[Ansible](https://github.com/ansible/ansible)** >= 2.4.0  
**[boto](https://github.com/boto/boto)** >= 2.48  
**[boto3](https://github.com/boto/boto3)** >= 1.4.7  
**IAM Service Role** - this is the IAM role that CloudFormation will use to create the resources in the target AWS account, it must have permissions for the CloudFormation service to be able to assume the role, and for the management of any AWS services created by the AWS Serivce Broker.  
**IAM User** - with at least the following permissions:
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
                "arn:aws:iam::421940136121:role/aws-service-broker-cloudformation"
            ],
            "Effect": "Allow"
        }
    ]
}
```
***NOTE:*** The resource ARN for the "iam:PassRole" permission must match the ARN of the service role created for CloudFormation stack actions.
