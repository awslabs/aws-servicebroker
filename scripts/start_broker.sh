#!/usr/bin/env bash
#
# Script to start the broker run by Dockerfile
#
aws-servicebroker \
    -insecure \
    -alsologtostderr \
    -region ${AWS_DEFAULT_REGION:=us-west-2} \
    -s3Bucket ${S3_BUCKET:=awsservicebrokeralpha} \
    -s3Key ${BUCKET_PREFIX:=pcf/templates} \
    -s3Region ${BUCKET_REGION:=us-west-2} \
    -port ${PORT:=3199} \
    -tableName ${DYNAMO_TABLE:=awssb} \
    -enableBasicAuth \
    -basicAuthUser ${SECURITY_USER_NAME:=admin} \
    -basicAuthPass ${SECURITY_USER_PASSWORD} \
    -v=${VERBOSE_LEVEL:=4}
