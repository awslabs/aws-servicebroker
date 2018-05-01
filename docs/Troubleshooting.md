# Troubleshooting

## Secrets created for [APBs](https://github.com/ansibleplaybookbundle/ansible-playbook-bundle) parameters were NOT filtered out

If you have created secrets (e.g. aws_access_key & aws_secret_key) for your APB's to consume, but you still see them as parameters in the AWS APBs, follow the steps bellow to troubleshoot

1. Verify that you've followed the [steps to create secrets](https://github.com/openshift/ansible-service-broker/blob/master/docs/secrets.md)

1. Verifying that the secret exist for APB parameters

    The secret (e.g. awsservicebroker-asb-secret) should exist in the ASB's namespace (e.g. aws-service-broker)
    ```bash
    $ oc get secret awsservicebroker-asb-secret -o yaml
    apiVersion: v1
    data:
    aws_access_key:               <ACCESS-KEY-VALUE>
    aws_cloudformation_role_arn:  <ROLE-ARM-VALUE>
    aws_secret_key:               <SECRET-KEY-VALUE>
    kind: Secret
    metadata:
    creationTimestamp: 2018-01-05T16:56:04Z
    name: awsservicebroker-asb-secret
    namespace: aws-service-broker
    resourceVersion: "1781"
    selfLink: /api/v1/namespaces/aws-service-broker/secrets/awsservicebroker-asb-secret
    uid: 21d2d8dd-ff95-11e7-b7ef-64006a55912e
    type: Opaque
    ```

1. Verify that the secret were processed for the APBs

    Review the ASB pod's logs, and search for messages which says `Filtering secrets from spec ...` and `Param <SECRET> matched`.  For example the below is a partial logs of the ASB which shows that the secret parameter values `aws_access_key, aws_secret_key, aws_cloudformation_role_arn` are being processed for each of the SNS APB's plans.
    ```bash
    ...
    [2018-01-05T16:56:46.905Z] [DEBUG] Filtering secret parameters out of specs...
    [2018-01-05T16:56:46.905Z] [DEBUG] Filtering secrets from spec dh-sns-apb
    [2018-01-05T16:56:46.908Z] [DEBUG] Found secret with name awsservicebroker-asb-secret
    [2018-01-05T16:56:46.908Z] [DEBUG] Found secret keys: [aws_access_key aws_cloudformation_role_arn aws_secret_key]
    [2018-01-05T16:56:46.908Z] [DEBUG] Filtering secrets from plan sns-topicwithsub
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_access_key matched%!(EXTRA string=aws_access_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_secret_key matched%!(EXTRA string=aws_secret_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_cloudformation_role_arn matched%!(EXTRA string=aws_cloudformation_role_arn)
    [2018-01-05T16:56:46.908Z] [DEBUG] Filtering secrets from plan sns-topic
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_access_key matched%!(EXTRA string=aws_access_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_secret_key matched%!(EXTRA string=aws_secret_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_cloudformation_role_arn matched%!(EXTRA string=aws_cloudformation_role_arn)
    [2018-01-05T16:56:46.908Z] [DEBUG] Filtering secrets from plan sns-subscription
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_access_key matched%!(EXTRA string=aws_access_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_secret_key matched%!(EXTRA string=aws_secret_key)
    [2018-01-05T16:56:46.908Z] [DEBUG] Param aws_cloudformation_role_arn matched%!(EXTRA string=aws_cloudformation_role_arn)
    ...
    ```

1. Verify that the parameters are NOT visible in the OpenShift WebUI

    Select any one of the APB's that a secret was created for, and verify that those parameters are NOT visible for a user to enter in the WebUI.

## Service provisioning fails

There are a variety of ways this can manifest:

* provision fails with error
* provision succeeds but the cloudformation stack either is non-existent, or in CREATE_FAILED state
* provision succeeds, but bind secrets are empty

### Monitoring APB Logs

The provisioning of an APB will occur in a temporary namespace/pod which will run the playbook for the APB's provision steps. When the provision is successful, this namespace/pod will be removed, and your APB will be available for use in the project/namespace that you've specified.

However if the provision was not successful, reviewing the logs of the pod that's temporarily launched would be help. Since the namespace/pod is random, it's best to open a terminal and issue a `watch` command to get pods in all of the namespaces before provisioning an APB.

Below is an example output of the command:  `watch 'oc get pods --all-namespaces'`
```bash
Every 2.0s: oc get pods --all-namespaces

NAMESPACE                           NAME                                  READY     STATUS      RESTARTS   AGE
aws-service-broker                  aws-asb-1-fqd89                       1/1       Running     0          1h
aws-service-broker                  aws-asb-etcd-1-mvwrk                  1/1       Running     0          1h
default                             docker-registry-1-6pknr               1/1       Running     0          1h
default                             persistent-volume-setup-jn7fj         0/1       Completed   0          1h
default                             router-1-zwplz                        1/1       Running     0          1h
kube-service-catalog                apiserver-742091420-mmqdg             2/2       Running     0          1h
kube-service-catalog                controller-manager-1159488142-24txt   1/1       Running     2          1h
openshift-template-service-broker   apiserver-7qr86                       1/1       Running     0          1h
```

After provisioning your APB, you will see another namespace/pod appear (e.g. RDS APB was provisioned)
```bash
Every 2.0s: oc get pods --all-namespaces

NAMESPACE                           NAME                                       READY     STATUS      RESTARTS   AGE
aws-service-broker                  aws-asb-1-fqd89                            1/1       Running     0          1h
aws-service-broker                  aws-asb-etcd-1-mvwrk                       1/1       Running     0          1h
default                             docker-registry-1-6pknr                    1/1       Running     0          1h
default                             persistent-volume-setup-jn7fj              0/1       Completed   0          1h
default                             router-1-zwplz                             1/1       Running     0          1h
dh-rds-apb-prov-r6cw6               apb-80d437a3-4a9f-46c0-b9fd-25635882450a   0/1       Error       0          5m
kube-service-catalog                apiserver-742091420-mmqdg                  2/2       Running     0          1h
kube-service-catalog                controller-manager-1159488142-24txt        1/1       Running     2          1h
openshift-template-service-broker   apiserver-7qr86                            1/1       Running     0          1h
```

The above shows that and `Error` occurred in the pod `apb-80d437a3-4a9f-46c0-b9fd-25635882450a` in the namespace `dh-rds-apb-prov-r6cw6`.  Review the logs in that pod to further investigate the error.

### Checking logs for Ansible playbook errors

Check whether the underlying ansible playbook experienced any errors:

```bash
journalctl --no-pager --since "-1 day" _SYSTEMD_UNIT=docker.service _COMM=dockerd-current | grep FAILED
```

### Checking for CloudFormation stack errors

To investigate a CloudFormation stack failure and it's causes, refer to the [Troubleshooting Guide](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/troubleshooting.html) in the CloudFormation Documentation.

## Troubleshooting AWS Service Broker issues

For troubleshooting general AWS Broker issues, review the debugging guide for the [Ansible Service Broker](https://github.com/openshift/ansible-service-broker/blob/master/docs/debugging.md).

### AWS Service broker logs

The AWS Service Broker logs can be viewed by running the following OpenShift command, note that these commands require "oc login" to already be completed.

```bash
oc logs po/$(oc get pods -n aws-service-broker --no-headers | awk '{print $1}') -c aws-asb -n aws-service-broker | less
```

### Kubernetes Service Catalog logs

The Kubernetes Service Catalog invokes the AWS Service Broker and monitors for provisioning status's and available AWS services.

```bash
oc logs $(oc get pods -n kube-service-catalog | grep controller-manager | awk '{print $1}') -n kube-service-catalog | less
```
