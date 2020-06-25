package broker

import (
	"errors"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/golang/glog"
)

// Create AWS Session
func AwsSessionGetter(keyid string, secretkey string, region string, accountId string, profile string, params map[string]string) *session.Session {
	// Check whether the target region has been overridden
	if params["region"] != "" {
		region = params["region"]
	}

	creds := awsCredentialsGetter(keyid, secretkey, profile, params, ec2metadata.New(session.Must(session.NewSession())), sts.New(session.Must(session.NewSession())))
	cfg := aws.NewConfig().WithCredentials(&creds).WithRegion(region)
	currentAccountSession := session.Must(session.NewSession(cfg))
	sess, err := assumeTargetRole(currentAccountSession, params, region, accountId)
	if err != nil {
		panic(err)
	}
	return sess
}

func AwsCfnClientGetter(sess *session.Session) CfnClient {
	return CfnClient{cloudformation.New(sess)}
}

func AwsSsmClientGetter(sess *session.Session) ssmiface.SSMAPI {
	return ssm.New(sess)
}

func AwsS3ClientGetter(sess *session.Session) S3Client {
	return S3Client{s3.New(sess)}
}

func AwsDdbClientGetter(sess *session.Session) *dynamodb.DynamoDB {
	return dynamodb.New(sess)
}

func AwsStsClientGetter(sess *session.Session) *sts.STS {
	return sts.New(sess)
}

func AwsIamClientGetter(sess *session.Session) iamiface.IAMAPI {
	return iam.New(sess)
}

func AwsLambdaClientGetter(sess *session.Session) lambdaiface.LambdaAPI {
	return lambda.New(sess)
}

func GetCallerId(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error) {
	return svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
}

func assumeTargetRole(sess *session.Session, params map[string]string, region string, accountId string) (*session.Session, error) {

	if params["target_role_name"] == "" {
		glog.Infof("Parameter 'target_role_name' not set. Not assuming role.")
		return sess, nil
	}

	// retrieve AWS partition from instance metadata service
	partition, err := ec2metadata.New(sess).GetMetadata("/services/partition")

	if err != nil {
		partition = "aws" // no access to metadata service, defaults to AWS Standard Partition
	}

	targetAccountRoleArn := generateRoleArn(params, accountId, partition)
	glog.Infof("Assuming role arn '%s'.", targetAccountRoleArn)
	credentialsTargetAccount := stscreds.NewCredentials(sess, targetAccountRoleArn)

	sessionTargetAccount := session.Must(session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentialsTargetAccount,
	}))

	return sessionTargetAccount, nil
}

func getObjectBody(s3svc S3Client, bucket, key string) ([]byte, error) {
	obj, err := s3svc.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	if obj.Body == nil {
		return nil, errors.New("s3 object body missing")
	}
	defer obj.Body.Close()
	file, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}
	return file, nil
}
