#!/bin/bash
if [[ ! -z "${OPENSHIFT_TARGET}" ]] && [[ ! -z "${OPENSHIFT_TOKEN}" ]]; then
  echo "Got OPENSHIFT token."
  LOGIN_PARAMS="--insecure-skip-tls-verify=true --token=$OPENSHIFT_TOKEN"
else
  echo "Attempting to login with a service account..."
  OPENSHIFT_TARGET=https://kubernetes.default
  LOGIN_PARAMS="--certificate-authority /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
    --token $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
fi

oc login $OPENSHIFT_TARGET $LOGIN_PARAMS