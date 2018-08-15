package broker

import (
	"flag"
)

// AddFlags adds defined flags to cli options
func AddFlags(o *Options) {
	flag.StringVar(&o.KeyID, "keyId", "", "AWS IAM User Key ID to use, if left blank will attempt to use a role, if defined secret-key must also be defined.")
	flag.StringVar(&o.SecretKey, "secretKey", "", "AWS IAM User Secret Key to use, if left blank will attempt to use a role, if defined key-id must also be defined.")
	flag.StringVar(&o.Profile, "profile", "", "AWS credential profile to use, mutually exclusive to key-id and secret-key.")
	flag.StringVar(&o.TableName, "tableName", "aws-service-broker", "DynamoDB table to use for persistent data storage.")
	flag.StringVar(&o.Region, "region", "us-east-1", "AWS Region the DynamoDB table and S3 bucket are stored in.")
	flag.StringVar(&o.S3Bucket, "s3Bucket", "awsservicebroker", "S3 bucket name where templates are stored.")
	flag.StringVar(&o.S3Region, "s3Region", "us-east-1", "region S3 bucket is located in.")
	flag.StringVar(&o.S3Key, "s3Key", "templates/", "S3 key where templates are stored.")
	flag.StringVar(&o.TemplateFilter, "templateFilter", "-main.yaml", "only process templates with the defined suffix.")
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog.")
	flag.StringVar(&o.BrokerID, "brokerId", "aws-service-broker", "An ID to use for partitioning broker data in DynamoDb. if multiple brokers are used in the same AWS account, this value must be unique per broker")
	flag.StringVar(&o.RoleArn, "roleArn", "", "CloudFormation service role ARN to use when launching service instances.")
}
