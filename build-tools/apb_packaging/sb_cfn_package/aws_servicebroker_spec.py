import os
import yaml
from taskcat.utils import CFNYAMLHandler
from collections import OrderedDict
import re
import copy
import boto3
import botocore
import zipfile
import random
import pip


class AwsServiceBrokerSpecException(Exception):
    def __init__(self, message=None, missing_key=None, incorrect_type=None):
        if missing_key and incorrect_type:
            raise Exception('Cannot specify 2 error types in one exception')
        if not message:
            if missing_key:
                message = "The AWS Service broker specification requires a key at %s" % missing_key
            elif incorrect_type:
                message = "AWS Service broker specification incorrect Type %s" % incorrect_type
        super(AwsServiceBrokerSpecException, self).__init__(message)
        self.errors={'missing_keys': missing_key, 'incorrect_type': incorrect_type }


class AwsServiceBrokerSpec(object):
    """
    Implements the AWS Service Broker Specification
    """
    def __init__(self, service_name, bucket_name=None, key_prefix='', region=None, profile=None, s3acl='private', test=False):
        """
        Load the specification from config files
        """
        package_path = os.path.dirname(os.path.abspath(__file__))
        with open(package_path + "/data/spec.yaml", 'r') as stream:
            self.spec = yaml.load(stream)
        with open(package_path + "/data/parameter_mappings.yaml", 'r') as stream:
            self.parameter_mapping = yaml.load(stream)
        with open(package_path + "/data/spec_mappings.yaml", 'r') as stream:
            self.spec_mappings = yaml.load(stream)
        with open(package_path + "/data/inject_parameters.yaml", 'r') as stream:
            self.inject_parameters = yaml.load(stream)
        self.bucket_name = bucket_name
        self.key_prefix = key_prefix
        if not self.key_prefix.endswith('/') and len(self.key_prefix) > 0:
            self.key_prefix = self.key_prefix + '/'
        self.region = region
        self.profile = profile
        self.s3acl = s3acl
        self.service_name = service_name
        self.test = test

    def lint(self, input_spec):
        """
        validate input against the specification
        :param input_spec:
        :return:
        """
        # TODO: implement

    def get_mapping(self, key, key_type='cfn'):
        """
        translates a key from the AWS Service Broker spec into it's APB equivalent

        :param key:
        :param key_type:
        :return:
        """
        if key_type not in ["cfn", "apb"]:
            raise AwsServiceBrokerSpecException("mapping key_type must be either 'cfn' or 'apb'")
        if not key:
            return None
        if key_type == 'cfn':
            key_value = self.spec_mappings[key]
        else:
            key_value = [m for m in self.spec_mappings.keys() if self.spec_mappings[m] == key][0]
        if key_value == 'None':
            key_value = None
        return key_value

    def build_abp_spec(self, cfn_spec, template, path, build_path=None):
        self.build_path = build_path
        self.template = template
        apb_spec = OrderedDict()
        for key in cfn_spec.keys():
            apb_key = self.get_mapping(key)
            if apb_key and type(cfn_spec[key]) in [float, int, str]:
                nested_keys = self._render_dot(apb_key)
                if len(nested_keys) == 1 or nested_keys[0] not in apb_spec.keys():
                    apb_spec = OrderedDict({**apb_spec, **self._build_nested(key, apb_key, cfn_spec)})
                else:
                    apb_spec[nested_keys[0]] = OrderedDict({**apb_spec[nested_keys[0]], **self._build_nested(key, apb_key, cfn_spec)[nested_keys[0]]})
        for key in self.spec.keys():
            apb_key = self.get_mapping(key)
            if apb_key:
                if apb_key not in apb_spec.keys() and self.spec[key]['Required'] and '.' not in apb_key:
                    raise AwsServiceBrokerSpecException('Required field "%s" is missing from spec metadata' % key)
                elif apb_key not in apb_spec.keys() and self.spec[key]['Default'] and '.' not in apb_key:
                    if apb_key == 'description':
                        apb_spec[apb_key] = re.compile('qs-([a-z0-9]){9}').sub('', template['Description']).strip()
                    elif apb_key not in ['plans']:
                        apb_spec[apb_key] = self.spec[key]['Default']
        prescribed_parameters = {}
        for key in cfn_spec.keys():
            apb_key = self.get_mapping(key)
            if apb_key and type(cfn_spec[key]) in [dict, OrderedDict]:
                if key == 'ServicePlans':
                    apb_spec[apb_key] = []
                    for sp in cfn_spec[key].keys():
                        plan, prescribed_parameters[sp] = self.build_plan(sp, cfn_spec[key][sp])
                        apb_spec[apb_key].append(plan)
        bindings = {"CFNOutputs": [], 'IAMUser': False}
        if 'Outputs' in self.template.keys():
            include = 'all'
            if 'Bindings' in cfn_spec.keys():
                if 'CFNOutputs' in cfn_spec['Bindings']:
                    include = cfn_spec['Bindings']['CFNOutputs']
            for op in self.template['Outputs'].keys():
                if include == 'all' or op in include:
                    bindings["CFNOutputs"].append(op)
        if 'Bindings' in cfn_spec.keys():
            if 'IAM' in cfn_spec['Bindings'].keys():
                if 'AddKeypair' in cfn_spec['Bindings']['IAM'].keys():
                    if cfn_spec['Bindings']['IAM']['AddKeypair'] == True:
                        bindings['IAMUser'] = True
                        if 'Policies' not in cfn_spec['Bindings']['IAM'].keys():
                            cfn_spec['Bindings']['IAM']['Policies'] = []
                        self._inject_iam(cfn_spec['Bindings']['IAM']['Policies'])
        self._inject_utils()
        self._build_functions(path)
        self._upload_template()
        if self.test:
            self.delete_asset_bucket()
        return {"apb_spec": apb_spec, "prescribed_parameters": prescribed_parameters, "bindings": bindings, "template": self.template}

    def _upload_template(self):
        for k in list(self.template.keys()):
            if not self.template[k]:
                self.template.pop(k)
        tpl = CFNYAMLHandler.ordered_safe_dump(self.template, default_flow_style=False)
        key = os.path.join(self.key_prefix, 'templates/%s/template.yaml' % self.service_name)
        self.s3_client.put_object(Body=tpl, Bucket=self.bucket_name, Key=key, ACL=self.s3acl)
        return self.bucket_name, key

    def _build_functions(self, path):
        path = os.path.join(path, 'functions')
        if os.path.isdir(path):
            self._inject_copy_zips()
            for func in [name for name in os.listdir(path) if os.path.isdir(os.path.join(path, name))]:
                self._publish_lambda_zip(os.path.join(path, func), func)
                self.template['Resources']['AWSSBInjectedCopyZips']['Properties']['Objects'].append(
                    func + '/lambda_function.zip')


    def _inject_iam(self, policies=None):
        with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/create_keypair/template.snippet", 'r') as stream:
            snippet = CFNYAMLHandler.ordered_safe_load(stream)
        with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/create_keypair/lambda_function.py", 'r') as stream:
            function_code = stream.read()
        snippet['Resources']['AWSSBInjectedIAMUserLambda']['Properties']['Code']['ZipFile'] = function_code
        policy_template = snippet['Resources'].pop('AWSSBInjectedIAMUserPolicy')
        policy_arns = []
        if policies:
            pnum = 0
            for policy in policies:
                if type(policy) in [dict, OrderedDict]:
                    pnum += 1
                    pname = 'AWSSBInjectedIAMUserPolicy%s' % str(pnum)
                    p = copy.deepcopy(policy_template)
                    p['Properties']['PolicyName'] = pname
                    p['Properties']['PolicyDocument'] = policy['PolicyDocument']
                    snippet['Resources'][pname] = p
                elif policy.startswith('arn:aws:iam'):
                    policy_arns.append(policy)
            if policy_arns:
                snippet['Resources']['AWSSBInjectedIAMUser']['Properties']['ManagedPolicyArns'] = policy_arns
            else:
                snippet['Resources']['AWSSBInjectedIAMUser'].pop('Properties')
        if 'Resources' not in self.template:
            self.template['Resources'] = {}
        if 'Outputs' not in self.template:
            self.template['Outputs'] = {}
        self.template['Resources'] = OrderedDict({**self.template['Resources'], **snippet['Resources']})
        self.template['Outputs'] = OrderedDict({**self.template['Outputs'], **snippet['Outputs']})

    def _inject_copy_zips(self):
        self._make_asset_bucket()
        if 'AWSSBInjectedCopyZips' not in self.template['Resources'].keys():
            with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/copy_zips/template.snippet", 'r') as stream:
                snippet = CFNYAMLHandler.ordered_safe_load(stream)
            with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/copy_zips/lambda_function.py", 'r') as stream:
                    function_code = stream.read()
            snippet['Resources']['AWSSBInjectedCopyZipsLambda']['Properties']['Code']['ZipFile'] = function_code
            p = snippet['Resources']['AWSSBInjectedCopyZipsRole']['Properties']['Policies']
            p[0]['PolicyDocument']['Statement'][0]['Resource'][0] = p[0]['PolicyDocument']['Statement'][0]['Resource'][0].replace(
                '${SourceBucketName}', self.bucket_name
            ).replace('${KeyPrefix}', self.key_prefix)
            p[0]['PolicyDocument']['Statement'][1]['Resource'][0] = p[0]['PolicyDocument']['Statement'][1]['Resource'][
                0].replace(
                '${KeyPrefix}', self.key_prefix
            )
            snippet['Resources']['AWSSBInjectedCopyZips']['Properties']['SourceBucket'] = self.bucket_name
            snippet['Resources']['AWSSBInjectedCopyZips']['Properties']['Prefix'] = self.key_prefix + 'functions/'
            self.template['Resources'] = OrderedDict({**self.template['Resources'], **snippet['Resources']})

    def _inject_utils(self):
        self._make_asset_bucket()
        for util in [
            ['CidrBlocks', 'get_cidrs', 'GetCidrs', 'CidrBlocks', 'AutoCidrs'],
            ['NumberOfAvailabilityZones', 'get_azs', 'GetAzs', 'AvailabilityZones', 'AutoAzs']
        ]:
            if util[0] in self.template['Parameters']:
                if self.template['Parameters'][util[3]]['Default'] == 'Auto':
                    with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/%s/template.snippet" % util[1], 'r') as stream:
                        snippet = CFNYAMLHandler.ordered_safe_load(stream)
                    if not os.path.isfile(os.path.dirname(os.path.abspath(__file__)) + "/functions/%s/requirements.txt" % util[1]):
                        with open(os.path.dirname(os.path.abspath(__file__)) + "/functions/%s/lambda_function.py" % util[1], 'r') as stream:
                            function_code = stream.read()
                        snippet['Resources']['AWSSBInjected%sLambda' % util[2]]['Properties']['Code']['ZipFile'] = function_code
                    else:
                        self._inject_copy_zips()
                        bucket, key = self._publish_lambda_zip(os.path.dirname(os.path.abspath(__file__)) + "/functions/%s/" % util[1], util[1])
                        snippet['Resources']['AWSSBInjected%sLambda' % util[2]]['Properties']['Code']['S3Bucket'] = '!Ref AWSSBInjectedLambdaZipsBucket'
                        snippet['Resources']['AWSSBInjected%sLambda' % util[2]]['Properties']['Code']['S3Key'] = key
                        snippet['Resources']['AWSSBInjected%sLambda' % util[2]]['Properties']['Handler'] = 'lambda_function.handler'
                        snippet['Resources']['AWSSBInjected%sLambda' % util[2]]['Properties']['Code'].pop('ZipFile')
                        self.template['Resources']['AWSSBInjectedCopyZips']['Properties']['Objects'].append(util[1] + '/lambda_function.zip')
                    temp_template = CFNYAMLHandler.ordered_safe_dump(self.template, default_flow_style=False).replace(
                        "!Ref %s" % util[3],
                        "!If [ %s, !GetAtt AWSSBInjected%s.%s, !Ref %s ]" % (util[4], util[2], util[3], util[3])
                    )
                    self.template = CFNYAMLHandler.ordered_safe_load(temp_template)
                    self.template['Resources'] = OrderedDict({**self.template['Resources'], **snippet['Resources']})
                    self.template['Conditions'] = OrderedDict({**self.template['Conditions'], **snippet['Conditions']})

    def _publish_lambda_zip(self, func_path, util_name):
        if self.build_path:
            os.makedirs(self.build_path, exist_ok=True)
            tmp_zip = os.path.join(self.build_path, '%s/functions/%s/lambda_function.zip' % (self.service_name, util_name))
            os.makedirs(os.path.join(self.build_path, '%s/functions/%s' % (self.service_name, util_name)), exist_ok=True)
        else:
            tmp_zip = '/tmp/%s-lambda_function.zip' % random.randrange(10000000, 99999999)
        os.chdir(func_path)
        if os.path.isfile("requirements.txt"):
            with open("requirements.txt") as f:
                for line in f:
                    pip.main(['install', '-U', '-t', '.', line])
        self._zipdir(func_path, tmp_zip)
        key = self.key_prefix + 'functions/' + util_name + '/lambda_function.zip'
        self.s3_client.upload_file(tmp_zip, self.bucket_name, key, ExtraArgs={"ACL": self.s3acl})
        if not self.build_path:
            os.remove(tmp_zip)
        return self.bucket_name, key

    def _zipdir(self, path, output):
        ziph = zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED)
        for root, dirs, files in os.walk(path):
            for file in files:
                ziph.write(os.path.join(root, file), os.path.join(root, file).replace(path, ''))

    def _make_asset_bucket(self):
        if self.profile:
            self.boto3_session = boto3.session.Session(profile_name=self.profile)
        else:
            self.boto3_session = boto3.session.Session()
            if not self.boto3_session.region_name and not self.region:
                self.region = 'us-east-1'
            elif not self.region:
                self.region = self.boto3_session.region_name
        self.s3_client = self.boto3_session.client('s3', region_name=self.region)
        if not self.bucket_name:
            identity = self.boto3_session.client('sts', region_name=self.region).get_caller_identity()
            account_id = identity['Account']
            self.bucket_name = 'awsservicebroker-assets-%s' % account_id
        try:
            self.s3_client.head_bucket(Bucket=self.bucket_name)
        except botocore.exceptions.ClientError as e:
            if e.response['Error']['Code'] != '404':
                raise
            else:
                self.s3_client.create_bucket(Bucket=self.bucket_name)
        return

    def delete_asset_bucket(self):
        s3objects = self.s3_client.list_objects_v2(Bucket=self.bucket_name)
        if 'Contents' in s3objects.keys():
            self.s3_client.delete_objects(Bucket=self.bucket_name, Delete={'Objects': [{'Key': key['Key']} for key in s3objects['Contents']]})
        self.s3_client.delete_bucket(Bucket=self.bucket_name)
    def _build_nested(self, key, apb_key, cfn_spec):
        output_dict = {}
        nested_keys = self._render_dot(apb_key)
        if len(nested_keys) == 1:
            output_dict[nested_keys[0]] = cfn_spec[key]
        else:
            if nested_keys[0] not in output_dict.keys():
                output_dict[nested_keys[0]] = {}
            output_dict[nested_keys[0]][nested_keys[1]] = cfn_spec[key]
        return output_dict

    def _render_dot(self, item):
        return item.split('.')

    def build_plan_parameters(self, plan_params):
        output_parameters = []
        default_parameters = self.inject_parameters.copy()
        cfn_map = {"Parameters": {}, "Interface": {}}
        for i in self.parameter_mapping.keys():
            if self.parameter_mapping[i].startswith('CFN::Parameters'):
                cfn_map["Parameters"][self.parameter_mapping[i].split('.')[1]] = i
            elif self.parameter_mapping[i].startswith('CFN::Interface::'):
                cfn_map["Interface"][self.parameter_mapping[i].replace('CFN::Interface::','')] = i
        if 'Parameters' not in self.template:
            self.template['Parameters'] = {}
        for p in self.template['Parameters'].keys():
            if p not in plan_params:
                param = self.template['Parameters'][p]
                apb_param = OrderedDict({'name': p})
                for e in param.keys():
                    if e in cfn_map["Parameters"].keys():
                        apb_param[cfn_map["Parameters"][e]] = param[e]
                if "AWS::CloudFormation::Interface" in self.template['Metadata'].keys():
                    if "ParameterLabels" in self.template['Metadata']['AWS::CloudFormation::Interface']:
                        if p in self.template['Metadata']['AWS::CloudFormation::Interface']['ParameterLabels']:
                            apb_param['title'] = self.template['Metadata']['AWS::CloudFormation::Interface']['ParameterLabels'][p]['default']
                    if "ParameterGroups" in self.template['Metadata']['AWS::CloudFormation::Interface']:
                        for group in self.template['Metadata']['AWS::CloudFormation::Interface']['ParameterGroups']:
                            if p in group['Parameters']:
                                apb_param['display_group'] = group['Label']['default']
                apb_param['required'] = True
                if 'Default' in self.template['Parameters'][p].keys():
                    if self.template['Parameters'][p]['Default'] == '':
                        apb_param.pop('default')
                        apb_param['required'] = False
                if 'NoEcho' in self.template['Parameters'][p].keys():
                    apb_param['display_type'] = 'password'
                if 'AllowedValues' in self.template['Parameters'][p].keys():
                    apb_param['type'] = 'enum'
                    apb_param['enum'] = self.template['Parameters'][p]['AllowedValues']
                elif 'Type' in self.template['Parameters'][p].keys():
                    if self.template['Parameters'][p]['Type'] == 'Number':
                        apb_param['type'] = 'int'
                    else:
                        apb_param['type'] = 'string'
                if 'default' in apb_param.keys():
                    if type(apb_param['default']) == str:
                        if apb_param['default'].isnumeric():
                            try:
                                apb_param['default'] = int(apb_param['default'])
                            except Exception:
                                pass
                output_parameters.append(apb_param)
        merged_params = []
        for injected in default_parameters:
            if injected['name'] not in [p['name'] for p in output_parameters]:
                merged_params.append(injected)
        merged_params+=output_parameters
        for p in plan_params.keys():
            for o in list(output_parameters):
                if o['name'] == p:
                    output_parameters.remove(o)
                    if type(plan_params[p]) not in [int, float]:
                        if plan_params[p].lower() == 'default':
                            plan_params[p] = o['default']
        return [merged_params, plan_params]

    def build_plan(self, name, plan_input):
        if not re.search(r'^[-a-z0-9]+$', name):
            raise AwsServiceBrokerSpecException('plan name "%s" is invalid, plannames can only consist of lower case letters, numbers and dashes (-)' % name)
        plan_output = {"name": name}
        plan_output['metadata'] = {}
        for i in plan_input.keys():
            if type(plan_input[i]) in [float, int, str]:
                if i in ['LongDescription', 'DisplayName', 'Cost']:
                    mapped = i[0].lower() + i[1:]
                    meta = {}
                    meta[mapped] = plan_input[i]
                    plan_output['metadata'] = {**plan_output['metadata'], **meta}
                else:
                    plan_output = OrderedDict({**plan_output, **self._build_nested(i, self.get_mapping("ServicePlans[].%s" % i), plan_input)['plans']})
        if 'free' not in plan_output.keys():
            plan_output['free'] = False
        if 'ParameterValues' not in plan_input.keys():
            plan_input['ParameterValues'] = {}
        plan_output['parameters'], prescribed_params = self.build_plan_parameters(plan_input['ParameterValues'])
        return plan_output, prescribed_params
