#!/bin/bash
#
# Build and push aws-service-broker to ECR repository
#
name=$1
version=$2

region=us-west-2
path=$GOPATH/src/github.com/awslabs/aws-service-broker

function help {
    echo "USAGE: $0 NAME VERSION"
}

if [ "$name" == "" ]; then
    help
    exit 1
fi

if [ "$version" == "" ]; then
    help
    exit 1
fi

set -e

cd $path

account_id=`aws sts get-caller-identity |jq -r .Account`
url=$account_id.dkr.ecr.$region.amazonaws.com

`aws ecr get-login --no-include-email --region $region`
docker build . -t $name:$version
docker tag $name:$version $url/$name:$version
docker push $url/$name:$version

cd -
