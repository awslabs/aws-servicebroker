#!/bin/bash -xe

docker login -u ${DH_USER} -p ${DH_PASS}
if [ "$(cat broker)" == "True" ] ; then
    docker pull ${BROKER_SOURCE}@sha256:$(cat broker_image_sha)
    docker image tag ${BROKER_SOURCE}@sha256:$(cat broker_image_sha) ${DH_ORG}/aws-service-broker:latest
    docker push ${DH_ORG}/aws-service-broker:latest
    docker tag ${DH_ORG}/aws-service-broker:latest ${DH_ORG}/aws-service-broker:$(cat ${CODEBUILD_SRC_DIR}/commit_id)
    docker push ${DH_ORG}/aws-service-broker:$(cat ${CODEBUILD_SRC_DIR}/commit_id)
fi
