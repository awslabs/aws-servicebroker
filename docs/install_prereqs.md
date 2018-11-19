### DynamoDB Table

Table can be created with the following aws cli command:

```bash
aws dynamodb create-table --attribute-definitions \
AttributeName=id,AttributeType=S AttributeName=userid,AttributeType=S \
AttributeName=type,AttributeType=S --key-schema AttributeName=id,KeyType=HASH \
AttributeName=userid,KeyType=RANGE --global-secondary-indexes \
'IndexName=type-userid-index,KeySchema=[{AttributeName=type,KeyType=HASH},{AttributeName=userid,KeyType=RANGE}],Projection={ProjectionType=INCLUDE,NonKeyAttributes=[id,userid,type,locked]},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}' \
--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
--region us-east-1 --table-name awssb
```

You can customize the table name as needed and pass in your table name using –tableName

### IAM 
 
By default the broker will use the same credentials for provisioning ServiceInstances and for broker operations like 
fetching the catalog and reading/writing metadata to DynamoDB.

The user or role that the broker runs as requires the following policy:
 
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
         "s3:GetObject",
         "s3:ListBucket"
      ],
      "Resource": [
         "arn:aws:s3:::awsservicebroker/templates/*",
         "arn:aws:s3:::awsservicebroker"
      ],
      "Effect": "Allow"
    },
    {
      "Action": [
        "dynamodb:PutItem",
        "dynamodb:GetItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "arn:aws:dynamodb:<REGION>:<ACCOUNT_ID>:table/<TABLE_NAME>",
      "Effect": "Allow"
    },
    {
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters"
      ],
      "Resource": [
          "arn:aws:ssm:<REGION>:<ACCOUNT_ID>:parameter/asb-*",
          "arn:aws:ssm:<REGION>:<ACCOUNT_ID>:parameter/Asb*",
      ],
      "Effect": "Allow"
    }
  ]
}
```

| **NOTE:** replace the `<REGION>`, `<ACCOUNT_ID>` and `<TABLE_NAME>` placeholders in the above json before creating the policy
 
The role/user used for provisioning requires additional permissions for provisioning, binding and deprovisioning ServiceInstances. By default this is the same user/role as the broker role, so can be added to that, or can be applied to a separate role, see [Managing Resources Via Assumed Role](/docs/README.md#managing-resources-via-assumed-role).

```json
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "SsmForSecretBindings",
        "Action": "ssm:PutParameter",
        "Resource": "arn:aws:ssm:<REGION>:<ACCOUNT_ID>:parameter/asb-*",
        "Effect": "Allow"
      },
      {
        "Sid": "AllowCfnToGetTemplates",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::awsservicebroker/templates/*",
        "Effect": "Allow"
      },
      {
         "Sid": "CloudFormation",
         "Action": [
            "cloudformation:CreateStack",
            "cloudformation:DeleteStack",
            "cloudformation:DescribeStacks",
            "cloudformation:UpdateStack",
            "cloudformation:CancelUpdateStack"
         ],
         "Resource": [
            "arn:aws:cloudformation:<REGION>:<ACCOUNT_ID>:stack/aws-service-broker-*/*"
         ],
         "Effect": "Allow"
      },
     {
        "Sid": "ServiceClassPermissions",
        "Action": [
           "athena:*",
           "dynamodb:*",
           "kms:*",
           "elasticache:*",
           "elasticmapreduce:*",
           "kinesis:*",
           "rds:*",
           "redshift:*",
           "route53:*",
           "s3:*",
           "sns:*",
           "sqs:*",
           "ec2:*",
           "iam:*",
           "lambda:*"
        ],
        "Resource": [
           "*"
        ],
        "Effect": "Allow"
     }
   ]
}
```

| **NOTE:** replace the `<REGION>` and `<ACCOUNT_ID>` placeholders in the above json before creating the policy

If a custom catalog is published, this policy may need to be adapted.
