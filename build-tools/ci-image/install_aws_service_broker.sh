#!/bin/bash

OSX=$(uname -s|grep -c Darwin)

function create-broker-resource {
    mkdir -p ./awssb-cert
    openssl req -nodes -x509 -newkey rsa:4096 -keyout ./awssb-cert/key.pem -out ./awssb-cert/cert.pem -days 365 -subj "/CN=awssb.aws-service-broker.svc"
    if [ "$OSX" == "1" ] ; then
        broker_ca_cert=$(cat ./awssb-cert/cert.pem | base64 -b 0)
    else
        broker_ca_cert=$(cat ./awssb-cert/cert.pem | base64 -w 0)
    fi
    kubectl create secret tls awssb-tls --cert="./awssb-cert/cert.pem" --key="./awssb-cert/key.pem" -n aws-service-broker
    client_token=$(kubectl get sa awsservicebroker-client -o yaml | grep -w awsservicebroker-client-token | grep -o 'awsservicebroker-client-token.*$')
    broker_auth='{ "bearer": { "secretRef": { "kind": "Secret", "namespace": "aws-service-broker", "name": "REPLACE_TOKEN_STRING" } } }'

    cat <<EOF > "./broker-resource.yaml"
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: aws-service-broker
spec:
  url: "https://awssb.aws-service-broker.svc:1338/aws-service-broker/"
  authInfo:
    ${broker_auth}
  caBundle: "${broker_ca_cert}"
EOF

    if [ "$OSX" == "1" ] ; then
        sed -i "" 's/REPLACE_TOKEN_STRING/'"$client_token"'/g' ./broker-resource.yaml
    else
        sed -i 's/REPLACE_TOKEN_STRING/'"$client_token"'/g' ./broker-resource.yaml
    fi

    c=0
    while [ "$(kubectl get pods | grep -c 1/1)" != "2" ] ; do if [ $c -gt 30 ]; then echo "failed waiting for AWS Service Broker Pod to come up..."; exit 1 ; fi; sleep 5 ; c=$((c+1)) ; done

    kubectl create -f ./broker-resource.yaml -n aws-service-broker
}

function aws-service-broker {

    kubectl create ns aws-service-broker

    context=$(kubectl config current-context)
    cluster=$(kubectl config get-contexts $context --no-headers | awk '{ print $3 }')

    kubectl config set-context $context --cluster=$cluster --namespace=aws-service-broker
    j2 ./k8s-aws-service-broker.yaml.j2 > ./k8s-aws-service-broker.yaml
    kubectl create -f "./k8s-aws-service-broker.yaml"

    create-broker-resource
}

echo "========================================================================"
echo "                       AWS_SERVICE_BROKER_ON_K8s                        "
echo "========================================================================"
echo ""
echo "PREREQUISITES:"
echo " - running kubernetes cluster with service-catalog installed."
echo " - kubectl with a cluster and context configured"
echo " - python 2.7 or 3.5+"
echo " - python jinja2 module installed in python path"
echo ""
echo "========================================================================"
echo ""

echo "Starting the AWS Service Broker"
aws-service-broker
