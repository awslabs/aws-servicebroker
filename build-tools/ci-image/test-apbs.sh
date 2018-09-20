#!/bin/bash -xe

ssh -i ~/.ssh/helioscibot.pem -o "StrictHostKeyChecking=no" ec2-user@$(echo ${OC_HOST} | awk -F ":" '{print $1}') "docker login -u ${DH_USER} -p ${DH_PASS} docker.io"
while [ "$(test_apb.py checklock)" == "-1" ] ; do
    sleep 20
done

oc login --insecure-skip-tls-verify ${OC_HOST} -u admin -p ${OC_PASS}

if [ "$(cat broker)" == "True" ] ; then
    APBS=$(ls -1 templates/)
    ssh -i ~/.ssh/helioscibot.pem -o "StrictHostKeyChecking=no" ec2-user@$(echo ${OC_HOST} | awk -F ":" '{print $1}') "docker pull docker.io/${DH_ORG}/aws-service-broker:latest"
    oc login -u admin -p ${OC_PASS} ${OC_HOST}
    oc project aws-service-broker
    oc rollout latest aws-asb
    oc rollout status dc/aws-asb
    retries=10
    while true; do
        retries=$((retries - 1))
        apb relist --broker-name aws-service-broker
        pod=$(oc get pods --no-headers | grep -v 'etcd' | grep -v 'deploy' | awk '{print $1}')
        if [ "$(oc logs ${pod} -c aws-asb | grep -c 'EXTRA string=aws_access_key')" != "0" ] ; then
            break
        fi
        if [ ${retries} -eq 0 ]; then
            echo "service broker failed to map secrets"
            exit 1
        fi
        sleep 120
    done
else
    APBS=$(cat apbs)
fi

if [ "$(echo $APBS)" == "" ] ; then
    echo "No APB's were changed, so there's nothing to test"
    exit 0
fi

FAILED=0
for APB_NAME in $(echo $APBS); do
    echo "Testing APB: ${APB_NAME}"
    ssh -i ~/.ssh/helioscibot.pem -o "StrictHostKeyChecking=no" ec2-user@$(echo ${OC_HOST} | awk -F ":" '{print $1}') "docker pull docker.io/${DH_ORG}/${APB_NAME}-apb:latest > /tmp/${APB_NAME}_docker_pull"
    cd ${CODEBUILD_SRC_DIR}/build/${APB_NAME}/apb
    test_apb.py ${APB_NAME} . admin ${OC_PASS} ${OC_HOST} $(cat ${CODEBUILD_SRC_DIR}/commit_id) > /tmp/${APB_NAME} 2>&1 &
    echo "${APB_NAME}" > /tmp/$!
    sleep 30
done

function failed () {
    echo "$1 failed testing"
    cat /tmp/$1
    let "FAILED+=1"
}

IFS=$'\n'
for l in $(jobs -l | grep test_apb.py) ; do
    p=$(echo "$l" | awk '{print $2}')
    name=$(cat /tmp/${p})
    wait ${p} || failed ${name}
done


LOCK_COUNT=$(test_apb.py checklock)
if [ "$LOCK_COUNT" == "0" ] ; then
    test_apb.py resetlock
    ssh -i ~/.ssh/helioscibot.pem -o "StrictHostKeyChecking=no" ec2-user@$(echo ${OC_HOST} | awk -F ":" '{print $1}') "source ~/.bash_profile ; cd ~/catasb/ec2/minimal ; ./reset_environment.sh"  > /tmp/catasb_reset.log 2>&1 || cat /tmp/catasb_reset.log
    test_apb.py resetunlock
fi

if [ "${FAILED}" != "0" ] ; then
    echo "${FAILED} APB(s) failed tests"
    exit 1
fi
