#!/usr/bin/env bash
#
# Script to start the broker run by Dockerfile
#
aws-servicebroker \
    -logtostderr \
    -brokerId=${BROKER_ID:=awsservicebroker} \
    -enableBasicAuth=true \
    -basicAuthUser=${SECURITY_USER_NAME:=admin} \
    -basicAuthPass=${SECURITY_USER_PASSWORD} \
    -insecure=${INSECURE:=false} \
    -port=${PORT:=3199} \
    -region=${TABLE_REGION:=us-east-1} \
    -s3Bucket=${S3_BUCKET:=awsservicebroker} \
    -s3Key=${S3_KEY:=templates/latest/} \
    -s3Region=${S3_REGION:=us-east-1} \
    -tableName=${TABLE_NAME:=aws-service-broker} \
    -templateFilter=${TEMPLATE_FILTER:=-main.yaml} \
    -tlsCert=${TLS_CERT} \
    -tlsKey=${TLS_KEY} \
    -prescribeOverrides=${PRESCRIBE:=true} \
    -v=${VERBOSITY:=5}
