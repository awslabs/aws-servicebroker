package broker

import "time"

// CacheTTL TTL for catalog cache record expiry
var CacheTTL = 1 * time.Hour

var nonCfnParams = []string{
	"aws_cloudformation_role_arn",
	"aws_access_key",
	"aws_secret_key",
	"region",
	"SBArtifactS3Bucket",
	"SBArtifactS3KeyPrefix",
	"target_role_name",
	"target_account_id",
}

const (
	bindParamRoleName = "RoleName"
	bindParamScope    = "Scope"
)

const (
	cfnOutputPolicyArnPrefix = "PolicyArn"
	cfnOutputSSMValuePrefix  = "ssm:"
	cfnOutputUserKeyID       = "UserKeyId"
	cfnOutputUserSecretKey   = "UserSecretKey"
)
