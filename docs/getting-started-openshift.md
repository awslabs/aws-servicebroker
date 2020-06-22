# Getting Started Guide - OpenShift

### Installation

You will need to complete the [installation prerequisites](/docs/install_prereqs.md) before proceeding.

The installation scripts require the oc commandline tool to be installed and logged in to the target OpenShift cluster.

```bash
mkdir awssb
cd awssb

### Fetch installation artifacts
wget https://raw.githubusercontent.com/awslabs/aws-servicebroker/release-v1.0.2/packaging/openshift/deploy.sh
wget https://raw.githubusercontent.com/awslabs/aws-servicebroker/release-v1.0.2/packaging/openshift/aws-servicebroker.yaml
wget https://raw.githubusercontent.com/awslabs/aws-servicebroker/release-v1.0.2/packaging/openshift/parameters.env
chmod +x deploy.sh

### Edit parameters.env and update parameters as needed
vi parameters.env

### If you are running on ec2 and have an IAM role setup with the required broker do not pass ACCESS_KEY_ID and SECRET_KEY
./deploy.sh <ACCESS_KEY_ID> <SECRET_KEY>

### check that the broker is running:
oc get pods | grep aws-servicebroker

### check servicebroker logs
oc logs $(oc get pods --no-headers -o name | grep aws-servicebroker)
```
