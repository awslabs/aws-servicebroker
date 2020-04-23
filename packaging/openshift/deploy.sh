#!/bin/bash

ACCESSKEYID=$(echo -n $1 | base64)
SECRETKEY=$(echo -n $2 | base64)

# On OpenShift 4.2 the project name has changed to "openshift-service-catalog-apiserver-operator"
oc projects -q | grep -q "^kube-service-catalog$" && proj=kube-service-catalog
oc projects -q | grep -q "^openshift-service-catalog-apiserver-operator$" && proj=openshift-service-catalog-apiserver-operator
[ ! "$proj" ] && echo "Error: Cannot find project" && exit 1

# Fetch the cert 
CA=`oc get secret -n $proj -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | grep -v '^$' | tail -n 1`

# Create the project and the AWS Service Broker
oc new-project aws-sb 
oc process -f aws-servicebroker.yaml --param-file=parameters.env \
	--param BROKER_CA_CERT=$CA \
	--param ACCESSKEYID=${ACCESSKEYID} \
	--param SECRETKEY=${SECRETKEY} | oc apply -f - -n aws-sb

