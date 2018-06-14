# build-tools

Tools for building AWS Ansible Playbook Bundles (APBs).

## Building a single AWS APB
1. Make desired changes to base data dir, which contains the template APB source files.
```
$ tree build-tools/apb_packaging/sb_cfn_package/data

data/
├── apb_template
│   [...]
│   ├── Dockerfile
│   ├── entrypoint.sh
│   ├── playbooks
│   │   ├── deprovision.yml
│   │   └── provision.yml
│   └── roles
│       ├── aws-deprovision-apb
│       [...]
│       └── aws-provision-apb
│       [...]
├── inject_parameters.yaml
├── parameter_mappings.yaml
├── spec_mappings.yaml
└── spec.yaml

```
2. Make desired changes to CloudFormation templates
```
$ tree aws-servicebroker/templates/

templates/
├── dynamodb
│   ├── ci
│   └── template.yaml
├── elasticache
│   ├── ci
│   └── template.yaml
├── emr
│   ├── ci
│   └── template.yaml
[...]
```

3. Install the build tools 
```
cd build-tools/apb_packaging/ && \
python setup.py install
```

4. Start a build (SQS APB used as an example).
```
cd ../../templates/sqs/ && \
sb_cfn_package -t DOCKERHUB_REPO_NAME/sqs-apb -n sqs -b my_s3_artifact_bucket --ci '../../build/' ./template.yaml`

# Options provided:
# -t, Docker image tag
# -n, AWS service name
# -b, S3 bucket used for artifact storage (must have write access)
# --ci, CI directory
```

5. Configure the AWS Broker with the new provided S3 artifact bucket name if changed.

