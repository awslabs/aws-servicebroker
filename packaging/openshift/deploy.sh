#!/bin/bash

ACCESSKEYID=$(echo -n $1 | base64)
SECRETKEY=$(echo -n $2 | base64)
oc new-project aws-sb
# On OpenShift 4.x the project name has changed to "openshift-service-catalog-apiserver"
oc project kube-service-catalog || oc project openshift-service-catalog-apiserver || echo "Warning: Cannot find project ..."
CA=`oc get secret                         -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | grep -v '^$' | tail -n 1`
oc process -f aws-servicebroker.yaml --param-file=parameters.env -p BROKER_CA_CERT=$CA -p ACCESSKEYID=${ACCESSKEYID} -p SECRETKEY=${SECRETKEY} | oc apply -f -
