# Getting Started Guide - Kubernetes

This guide uses helm, for documentation on installing the helm client see [https://docs.helm.sh/using_helm/#install-helm](https://docs.helm.sh/using_helm/#install-helm)


### Installing Kubernetes Service Catalog

```bash
# Install helm and tiller into the cluster
helm init
# Wait until tiller is ready before moving onuntil kubectl get pods -n kube-system -l name=tiller | grep 1/1; do sleep 1; done

kubectl create clusterrolebinding tiller-cluster-admin \
    --clusterrole=cluster-admin \
    --serviceaccount=kube-system:default
# Adds the chart repository for the service catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
# Installs the service catalog
helm install svc-cat/catalog --name catalog --namespace catalog
```

### Installing the AWS Service Broker

```bash
# Add the service broker chart repository
helm repo add aws-sb https://awsservicebroker.s3.amazonaws.com/charts

# Show the available variables for the chart
helm inspect aws-sb/aws-servicebroker

# Minimal broker install, assuming defaults above. Add flags to set credentials, region, etc
helm install aws-sb/aws-servicebroker -n aws-servicebroker -ns aws-sb --devel
```
