#!/bin/bash

ACCESSKEYID=$(echo -n $1 | base64)
SECRETKEY=$(echo -n $2 | base64)
oc new-project aws-sb
CA=`oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | grep -v '^$' | tail -n 1`
oc process -f aws-servicebroker.yaml --param-file=parameters.env -p BROKER_CA_CERT=$CA -p ACCESSKEYID=${ACCESSKEYID} -p SECRETKEY=${SECRETKEY} | oc apply -f -
