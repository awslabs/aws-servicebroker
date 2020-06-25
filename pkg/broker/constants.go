package broker

import (
	"time"
)

// CacheTTL TTL for catalog cache record expiry
var CacheTTL = 1 * time.Hour

var nonCfnParams = []string{
	"region",
	"target_role_name",
	"target_account_id",
	"user_tags",
	"admin_tags",
}

var nonCfnParamDefs = map[string]interface{}{
	"target_account_id": map[string]interface{}{
		"description":   "AWS Account ID to provision into",
		"display_group": "AWS Account Information",
		"title":         "AWS Target Account ID",
		"type":          "string",
		"default":       "",
	},
	"target_role_name": map[string]interface{}{
		"description":   "AWS IAM Role name to use for provisioning",
		"display_group": "AWS Account Information",
		"title":         "AWS Target Role Name",
		"type":          "string",
		"default":       "",
	},
	"region": map[string]interface{}{
		"description":   "AWS Region to provision into",
		"display_group": "AWS Account Information",
		"title":         "AWS Region",
		"type":          "string",
		"enum": []string{
			"ap-northeast-1",
			"ap-northeast-2",
			"ap-south-1",
			"ap-southeast-1",
			"ap-southeast-2",
			"ca-central-1",
			"eu-central-1",
			"eu-west-1",
			"eu-west-2",
			"sa-east-1",
			"us-east-1",
			"us-east-2",
			"us-west-1",
			"us-west-2",
		},
	},
	"user_tags": map[string]interface{}{
		"description":   "AWS Resource tags to apply to resources (json formatted [{\"Key\": \"MyTagKey\", \"Value\": \"MyTagValue\"}, ...]",
		"display_group": "AWS Account Information",
		"title":         "AWS Tags",
		"type":          "string",
		"default":       "[]",
	},
	"admin_tags": map[string]interface{}{
		"description":   "AWS Resource tags to apply to resources (json formatted [{\"Key\": \"MyTagKey\", \"Value\": \"MyTagValue\"}, ...]",
		"display_group": "AWS Account Information",
		"title":         "Additional AWS Tags",
		"type":          "string",
		"default":       "[]",
	},
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
	cfnOutputBindLambda      = "BindLambda"
)

const (
	templateIDRegex = `\(qs-[a-z0-9]{9}\)`
)
