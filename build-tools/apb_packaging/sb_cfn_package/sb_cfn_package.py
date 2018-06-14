import argparse
import logging
import yaml
import json
from taskcat.utils import CFNYAMLHandler
import os
from random import random
from base64 import b64encode
import shutil
import subprocess
import re
import jinja2
try:
    from aws_servicebroker_spec import AwsServiceBrokerSpec
except:
    from sb_cfn_package.aws_servicebroker_spec import AwsServiceBrokerSpec


def cli():
    if len(logging.getLogger().handlers) == 0:
        logging.getLogger().addHandler(logging.StreamHandler())
    logging.getLogger().handlers[0].setFormatter(logging.Formatter('%(levelname)s: %(message)s'))
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '-l',
        '--loglevel',
        default='error',
        help='Set loglevel. Allowed values are debug, info, warning, error, critical. Default is warning'
    )
    parser.add_argument(
        '-n',
        '--name',
        help='name of AWS service'
    )
    parser.add_argument(
        "-s",
        "--service-spec-path",
        default=None,
        help='Path to the service specification to use for the build, if none is provided, the spec will be exctracted from the CloudFormation template'
    )
    parser.add_argument(
        "-t",
        "--docker-image-tag",
        default=None,
        help='tag to use for the docker image'
    )
    parser.add_argument(
        "-a",
        "--s3-acl",
        default='private',
        help='acl to use for objects uploaded to S3, default is private'
    )
    parser.add_argument(
        "-b",
        "--s3-bucket",
        default=None,
        help='bucket to use for artifacts, will autogenerate a new bucket by default'
    )
    parser.add_argument(
        "-p",
        "--profile",
        default=None,
        help='aws credential profile to use'
    ),
    parser.add_argument(
        "-c",
        "--ci",
        default=None,
        help='Path to the place build output, if not specified a random directory will be created in /tmp'
    )
    parser.add_argument(
        "templatepath",
        help='Path to the CloudFormation template to use for the build'
    )
    args = parser.parse_args()
    template_path = os.path.abspath(args.templatepath)
    loglevel = getattr(logging, args.loglevel.upper())
    logging.getLogger().setLevel(loglevel)
    logging.info('Set loglevel to %s' % args.loglevel.upper())
    logging.debug("Passed arguments: {} ".format(args.__dict__))
    build_path = None
    if args.ci:
        build_path = os.path.abspath(args.ci)
        try:
            shutil.rmtree(build_path + "/%s" % args.name)
        except FileNotFoundError:
            pass
    sb_pack = SbCfnPackage(template_path=template_path, service_spec_path=args.service_spec_path)
    artifacts = sb_pack.build_artifacts(args.name, args.s3_acl, args.s3_bucket, args.profile, build_path=build_path)
    results = sb_pack.create_apb_skeleton(artifacts['apb_spec'], artifacts['prescribed_parameters'],
                                          artifacts['bindings'], artifacts['template'], args.name, build_path=build_path)
    os.chdir(os.path.join(results, 'apb'))
    tag = args.docker_image_tag or artifacts['apb_spec']['name']
    results = subprocess.run(["apb", "build", "--tag", tag], stdout=subprocess.PIPE)
    print(results.stdout.decode("utf-8"))
    if results.returncode != 0:
        if results.stderr:
            print(results.stderr.decode("utf-8"))
        raise Exception('apb build failed')
    if '/' in tag:
        results = subprocess.run(["docker", "push", tag], stdout=subprocess.PIPE)
        for l in results.stdout.decode("utf-8").split('\n'):
            if not l.endswith(': Preparing') and not l.endswith(': Waiting'):
                print(l)
        if results.returncode != 0:
            if results.stderr:
                print(results.stderr.decode("utf-8"))
            raise Exception('docker push failed')
    if args.ci:
        os.makedirs('./ci')
        shutil.copy(os.path.join(os.path.dirname(template_path), 'ci/config.yml'), './ci/config.yml')


class SbCfnPackage(object):
    """
    Main class to handle all of the packaging operations required to turn a CloudFormation template into an APB
    """
    def __init__(self, template_path=None, service_spec_path=None):
        """
        Initialise the class, optionally providing paths for the template and a seperate service spec, if
        service_spec_path is not specified then we'll look for it in the template Metadata.

        :param template_path:
        :param service_spec_path:
        """
        self.template = {}
        self.service_spec = {}
        if template_path:
            self.template_path = os.path.dirname(template_path)
            with open(template_path, 'r') as stream:
                self.template = CFNYAMLHandler.ordered_safe_load(stream)
            if not service_spec_path:
                self.service_spec = self.template['Metadata']['AWS::ServiceBroker::Specification']
        if service_spec_path:
            with open(service_spec_path, 'r') as stream:
                self.service_spec = yaml.load(stream)
        if not self.service_spec:
            raise Exception("cannot continue without either a ['Metadata']['AWS::ServiceBroker::Specification'] section in the template, or a path to a seperate spec using service_spec_path")

    def build_artifacts(self, service_name, s3acl='private', bucket=None, profile=None, test=False, build_path=None):
        """
        Builds artifacts required to create an apb using the specification in the cloudformation template metadata

        :return:
        """
        return AwsServiceBrokerSpec(service_name=service_name, bucket_name=bucket, profile=profile, s3acl=s3acl, test=test).build_abp_spec(self.service_spec, self.template, self.template_path, build_path=build_path)

    def create_apb_skeleton(self, apb_spec, prescribed_parameters, bindings, template, service_name, build_path=None):
        if build_path:
            os.makedirs(build_path, exist_ok=True)
            tmpname = os.path.join(build_path, "%s" % service_name)
            os.makedirs(os.path.join(build_path, "%s" % service_name), exist_ok=True)
        else:
            tmpname = '/tmp/AWSSB-' + str(b64encode(bytes(str(random()), 'utf8'))).replace("b'", '').replace("'", '').replace('=', '')
            os.makedirs(tmpname)
        print("build path: %s" % tmpname)
        shutil.copytree(os.path.dirname(os.path.abspath(__file__)) + '/data/apb_template/', tmpname + '/apb')
        for dname, dirs, files in os.walk(tmpname):
            for fname in files:
                fpath = os.path.join(dname, fname)
                if not fname.endswith('.zip'):
                    with open(fpath) as f:
                        s = f.read()
                    s = s.replace("${SERVICE_NAME}", service_name).replace("${SERVICE_NAME_UPPER}", service_name.upper()).replace('${CREATE_IAM_USER}', str(bindings['IAMUser']))
                    with open(fpath, "w") as f:
                        f.write(s)
        for plan in prescribed_parameters.keys():
            prescribed_parameters[plan]['params_string'] = "{{ namespace }}::{{ _apb_plan_id }}::{{ _apb_service_class_id }}::{{ _apb_service_instance_id }}"
            prescribed_parameters[plan]['params_hash'] = "{{ params_string | checksum }}"
            with open(tmpname + '/apb/roles/aws-provision-apb/vars/%s.yml' % plan, "w") as f:
                f.write(CFNYAMLHandler.ordered_safe_dump(prescribed_parameters[plan], default_flow_style=False))
            shutil.copy(tmpname + '/apb/roles/aws-provision-apb/vars/%s.yml' % plan, tmpname + '/apb/roles/aws-deprovision-apb/vars/%s.yml' % plan)
        with open(tmpname + '/apb/apb.yml', "w") as f:
            f.write(CFNYAMLHandler.ordered_safe_dump(apb_spec, default_flow_style=False))
        with open(tmpname + '/apb/roles/aws-provision-apb/tasks/main.yml') as f:
            main_provision_task = yaml.load(f)
        create_user = False
        try:
            create_user = template['Metadata']['AWS::ServiceBroker::Specification']['Bindings']['IAM']['AddKeypair']
        except KeyError as e:
            pass
        full_bindings = []
        try:
            add_iam = bool(template['Metadata']['AWS::ServiceBroker::Specification']['Bindings']['IAM']['AddKeypair'])
        except KeyError:
            add_iam = False
        if add_iam:
            full_bindings.append({
                "name": apb_spec['name'].upper() + '_AWS_ACCESS_KEY_ID',
                "description": 'AWS IAM Access Key ID, your application must use this for authenticating runtime calls to the %s service' % apb_spec['metadata']['displayName']
            })
            full_bindings.append({
                "name": apb_spec['name'].upper() + '_AWS_SECRET_ACCESS_KEY',
                "description": 'AWS IAM Secret Access Key, your application must use this for authenticating runtime calls to the %s service' % apb_spec['metadata']['displayName']
            })
        for t in main_provision_task:
            if 'name' in t.keys():
                if t['name'] == 'Encode bind credentials':
                    if not create_user:
                        aws_key_id = '%s_AWS_ACCESS_KEY_ID' % service_name
                        aws_key = '%s_AWS_SECRET_ACCESS_KEY' % service_name
                        t['asb_encode_binding']['fields'].pop(aws_key_id.upper())
                        t['asb_encode_binding']['fields'].pop(aws_key.upper())

                    for b in bindings['CFNOutputs']:
                        t['asb_encode_binding']['fields'][camel_convert(b).upper()] = "{{ cfn.stack_outputs.%s }}" % b
                        description = ""
                        if "Description" in template['Outputs'][b].keys():
                            description = template['Outputs'][b]["Description"]
                        full_bindings.append({"name": camel_convert(b).upper(), "description": description})
            elif 'block' in t.keys():
                for it in t['block']:
                    if it['name'] == 'Create Resources':
                        if 'Parameters' in template.keys():
                            for p in template['Parameters'].keys():
                                default = ""
                                if 'Default' in template['Parameters'][p].keys():
                                    default = template['Parameters'][p]['Default']
                                it['cloudformation']['template_parameters'][p] = '{{ %s | default("%s") | string }}' % (p, default)
        with open(tmpname + '/apb/roles/aws-provision-apb/tasks/main.yml', 'w') as f:
            f.write(CFNYAMLHandler.ordered_safe_dump(main_provision_task, default_flow_style=False))
        with open(tmpname + '/template.yaml', 'w') as f:
            f.write(CFNYAMLHandler.ordered_safe_dump(template, default_flow_style=False))
        render_documentation(apb_spec, template, prescribed_parameters, tmpname, full_bindings, add_iam)
        return tmpname


def render_documentation(apb, template, prescribed_params, tmp_path, bindings, add_iam):
    dir_path = os.path.dirname(os.path.realpath(__file__))
    abs_path = os.path.join(dir_path, 'data/serviceclass_documentation_template.md.j2')
    path, filename = os.path.split(abs_path)
    lengths = {}
    for plan in apb['plans']:
        lengths[plan['name']] = {"required": 0, "prescribed": 0, "optional": 0, "generic": 0}
        if 'parameters' in plan.keys():
            lengths[plan['name']]["required"] = len([p for p in plan['parameters'] if 'default' not in p.keys() and p['name'] not in ['aws_access_key', 'aws_secret_key', 'aws_cloudformation_role_arn', 'region', 'SBArtifactS3Bucket', 'SBArtifactS3KeyPrefix', 'VpcId']])
            lengths[plan['name']]["optional"] = len([p for p in plan['parameters'] if 'default' in p.keys() and p['name'] not in ['aws_access_key', 'aws_secret_key', 'aws_cloudformation_role_arn', 'region', 'SBArtifactS3Bucket', 'SBArtifactS3KeyPrefix', 'VpcId']])
            lengths[plan['name']]["generic"] = len([p for p in plan['parameters'] if p['name'] in ['aws_access_key', 'aws_secret_key', 'aws_cloudformation_role_arn', 'region', 'SBArtifactS3Bucket', 'SBArtifactS3KeyPrefix', 'VpcId']])
            lengths[plan['name']]["prescribed"] = len([p for p in prescribed_params[plan['name']] if p not in ["params_string", "params_hash"]])
    if add_iam:
        iam_policy = json.dumps(template['Metadata']['AWS::ServiceBroker::Specification']['Bindings']['IAM']['Policies'], sort_keys=True, indent=4, separators=(',', ': '))
    result = jinja2.Environment(loader=jinja2.FileSystemLoader(path or './')).get_template(filename).render(
        {"apb": apb, "template": template, "prescribed_params": prescribed_params, "lengths": lengths, "bindings": bindings, "add_iam": add_iam, 'iam_policy': iam_policy}
    )

    with open(os.path.join(tmp_path, 'README.md'), 'w') as rendered_file:
        rendered_file.write(result)
    return

def camel_convert(name):
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()
