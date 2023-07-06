# Getting Started Guide - Kubernetes

This guide uses helm, for documentation on installing the helm and tiller see [https://docs.helm.sh/using_helm/#install-helm](https://docs.helm.sh/using_helm/#install-helm)


### Installing Kubernetes Service Catalog

```bash
# Adds the chart repository for the service catalog
helm repo add svc-cat https://kubernetes-sigs.github.io/service-catalog
# Installs the service catalog
helm install catalog svc-cat/catalog --namespace catalog
```

### Optional - reduce service catalog max polling interval
The default max interval is 20 minutes, which is too long in most cases, this can be reduced by adding 
`--operation-polling-maximum-backoff-duration=120s` as an additional argument to the controller-manager deployment under 
`spec.template.containers[0].args`.

### Installing the AWS Service Broker

```bash
# Add the service broker chart repository
helm repo add aws-sb https://awsservicebroker.s3.amazonaws.com/charts

# Show the available variables for the chart
helm show values aws-sb/aws-servicebroker
### Note: If setting aws.targetaccountid on the helm cli, do not use --set, use --set-string, see https://github.com/helm/helm/issues/1707 for more info

# Minimal broker install, assuming defaults above. Sets up a ClusterServiceBroker. Add flags to set credentials, region, etc
helm install aws-sb/aws-servicebroker \
  --name aws-servicebroker \
  --namespace aws-sb \
  --set aws.secretkey=REPLACEME \
  --set aws.accesskeyid=REPLACEME
  
  
# Install broker for the specified namespace only
helm install aws-sb/aws-servicebroker --name aws-servicebroker --namespace aws-sb \
  --set deployNamespacedServiceBroker=true
```
