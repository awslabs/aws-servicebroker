# build-tools

Tools for building AWS Ansible Playbook Bundles (APBs).

## Building a single AWS APB
1. Make any desired changes to base data in `build-tools/apb_packaging/sb_cfn_package/data`
1. Install the build tools `cd build-tools/apb_packaging/ && python setup.py install`
1. Create an S3 bucket to store AWS APB artifacts in. E.g. my_s3_artifact_bucket
1. Start a build with `cd ../../templates/s3/ && sb_cfn_package -t DOCKERHUB_REPO_NAME/s3-apb -n s3 -b my_s3_artifact_bucket --ci '../../build/' ./template.yaml`
1. Be sure to configure your broker with the new S3ArtifactBucket name.

