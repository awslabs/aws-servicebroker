package broker

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/golang/glog"
)

// Create AWS Session
func AwsSessionGetter(keyid string, secretkey string, region string, accountId string, profile string, params map[string]string) *session.Session {
	creds := AwsCredentialsGetter(keyid, secretkey, profile, params)
	cfg := aws.NewConfig().WithCredentials(&creds).WithRegion(region)
	currentAccountSession := session.Must(session.NewSession(cfg))
	sess, err := assumeTargetRole(currentAccountSession, params, region, accountId)
	if err != nil {
		panic(err)
	}
	return sess
}

func AwsCfnClientGetter(sess *session.Session) *cloudformation.CloudFormation {
	return cloudformation.New(sess)
}

func AwsSsmClientGetter(sess *session.Session) *ssm.SSM {
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

func GetCallerId(svc stsiface.STSAPI) (*sts.GetCallerIdentityOutput, error) {
	return svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
}

func assumeTargetRole(sess *session.Session, params map[string]string, region string, accountId string) (*session.Session, error) {

	if _, ok := params["target_role_name"]; !ok {
		glog.Infof("Parameter 'target_role_name' not set. Using process credentials.")
		return sess, nil
	}

	targetAccountRoleArn := generateRoleArn(params, accountId)
	glog.Infof("Assuming role arn '%s'.", targetAccountRoleArn)
	credentialsTargetAccount := stscreds.NewCredentials(sess, targetAccountRoleArn)

	sessionTargetAccount := session.Must(session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentialsTargetAccount,
	}))

	return sessionTargetAccount, nil
}
