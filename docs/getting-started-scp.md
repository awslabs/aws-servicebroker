This guide helps you integrate AWS services into the Cloud Foundry environment on SAP Cloud Platform by using the AWS Service Broker.

For a high-level overview of how this works, take a look at the following image:

![SCP AWS Service Broker integration](/docs/images/scp-aws-service-broker.png)

## Pre-requisites

* Account on [SAP Cloud Platform, Cloud Foundry environment](https://cloudplatform.sap.com/index.html).
* Account on [AWS](https://aws.amazon.com/free/?awsf.Free%20Tier%20Types=*all&all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc#featured).
* [Cloud Foundry CLI](https://help.sap.com/viewer/e275296cbb1e4d5886fa38a2a2c78c06/Cloud/en-US/4a8eb630c2734c01a25090c51d48102b.html).
* [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

## Configuration in your AWS Environment

1. Sign in to your AWS account and in the AWS console, search for Cloudformation.
3. Choose Create Stack and under Choose a Template, pick Upload a template to Amazon S3.
4. Save the contents of the [this file](https://github.com/awslabs/aws-servicebroker/blob/master/setup/prerequisites.yaml) to your local environment as a yaml file. In the AWS console, click Choose and upload this file from your local environment.
5. Choose Next.

Once the Cloudformation stack creation is complete, the DynamoDB table and the IAM user are displayed as follows:

![AWS Cloudformation Stack](/docs/images/aws-cloudformation.png)

## Deploy the AWS Service Broker to SAP Cloud Platform

1. Download the binary zip, aws-sb-cloudfoundry-app.zip, of the AWS Service Broker [here](https://github.com/awslabs/aws-servicebroker/releases).
2. Unzip the file. Open a terminal and navigate to the AWS Service Broker repository:
```
cd aws-sb-cf-cloudfoundry-app
```
3. Log in to SAP Cloud Platform in the Cloud Foundry environment:
```
cf login
```
4. Adapt the following URL to enter your API endpoint:
https://api.<REGION_TECHNICAL_KEY>.hana.ondemand.com.

Find out your region and the technical key [here](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/350356d1dc314d3199dca15bd2ab9b0e.html#loiof344a57233d34199b2123b9620d0bb41)

5. Enter username and password for your SAP Cloud Platform account.
6. Edit the `aws-sb-cf-cloudfoundry-app/manifest.yml` file and replace the values within the angular brackets:
```yaml
applications:
- name: aws-service-broker
buildpacks:
- binary_buildpack
command: |
export PARAM_OVERRIDE_${BROKER_ID}_all_all_all_region=${AWS_DEFAULT_REGION} ; ./cfnsb --logtostderr --prescribeOverrides=${PRESCRIBE} --v=${VERBOSITY} --brokerId=${BROKER_ID} --enableBasicAuth=true --insecure=${INSECURE} --port=${PORT} --region=${AWS_DEFAULT_REGION} --s3Bucket=${S3_BUCKET} --s3Key=${S3_KEY} --s3Region=${S3_REGION} --tableName=${TABLE_NAME} --templateFilter=${TEMPLATE_FILTER} --tlsCert=${TLS_CERT} --tlsKey=${TLS_KEY}
disk_quota: 1G
env:
# Required
AWS_ACCESS_KEY_ID: <ENTER YOUR AWS ACCOUNT KEY>
AWS_SECRET_ACCESS_KEY: <ENTER YOUR AWS ACCOUNT KEY SECRET> 
SECURITY_USER_PASSWORD: <ENTER ANY SECURITY USER PASSWORD>
# Optional
AWS_DEFAULT_REGION: <ENTER YOUR REGION>
BROKER_ID: awsservicebroker
INSECURE: "true"
PRESCRIBE: "true"
ROOT: $HOME
S3_BUCKET: awsservicebroker
S3_KEY: templates/latest
S3_REGION: us-east-1
SCHEME: https
SECURITY_USER_NAME: awssb
TABLE_NAME: awssb
TEMPLATE_FILTER: -main.yaml
TLS_CERT: ""
TLS_KEY: ""
VERBOSITY: "5"
instances: 1
memory: 1G
```
7. Push the AWS Service Broker to Cloud Foundry
```
cf push
```
You can view the URL of the deployed service that will be used in the next step either in the CF CLI or in the SAP Cloud Platform cockpit.

8. Adapt and use the following command to register the AWS Service Broker:
```
cf create-service-broker aws-service-broker <SECURITY_USER_NAME> <SECURITY_USER_PASSWORD> <URL_OF_THE_SERVICE_BROKER>
```
For <SECURITY_USER_NAME> and <SECURITY_USER_PASSWORD>, use the values you entered into the manifest.yml. You can find the <URL_OF_THE_SERVICE_BROKER> in your Space in the SAP Cloud Platform cockpit under Application Routes.

Note: You must either be assigned the role of a Cloud Foundry administrator or limit the registration to a single Cloud Foundry space by using the --space-scoped flag.

For further details, refer to the [blog](https://blogs.sap.com/2019/06/04/how-to-consume-aws-services-on-sap-cloud-platform/) and the official documentation of [SAP Cloud Platform](https://help.sap.com/viewer/a7e6a78032b6488e98a39f4e9ab3b241/Cloud/en-US).