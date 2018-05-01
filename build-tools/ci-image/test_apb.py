#!/usr/bin/env python
from sys import argv, exit
import boto3
import yaml
from openshift import client, config
import subprocess
import shlex
import random
import re
import base64
from time import sleep
from botocore.vendored import requests
from cStringIO import StringIO
import os
import traceback


CLUSTER = os.environ['OCP_CLUSTER_HOSTNAME']
LOCK_DDB_TABLE = os.environ['LOCK_DDB_TABLE']
REGION = os.environ['AWS_REGION']
PROJECT_NAME = os.environ['PROJECT_NAME']
SHORT_HOSTNAME = os.environ['SHORT_HOSTNAME']

ignore_list = ['aws_access_key', 'aws_secret_key', 'aws_cloudformation_role_arn', 'region', 'VpcId']


class AwsApbCi(object):

    def __init__(self, apb_name, path='.', username=None, password=None, hostname=None, ddb_table=None, lock_only=False):
        self.path=path
        if not lock_only:
            self.required = self._get_required_params()
            self.config = self._parse_ci_config()
            self.assert_config()
            if username and password and hostname:
                self.login(username, password, hostname)
            self.oapi = self._get_api_client()
        self.lock_table = None

    def lock(self, reset_cluster=False):
        if not self.lock_table:
            self.lock_table = dynamodb.Table(LOCK_DDB_TABLE)
        if not reset_cluster:
            retries = 0
            while True:
                if self.get_lock() != -1:
                    break
                sleep(10)
                retries += 1
                if retries > 60:
                    raise Exception("Cannot lock, current status -1 for more than 10 minutes")
            self.lock_table.update_item(
                Key={'cluster': CLUSTER},
                UpdateExpression="SET lock_count = lock_count + :r",
                ExpressionAttributeValues = {':r': 1}
            )
        else:
            self.lock_table.update_item(
                Key={'cluster': CLUSTER},
                UpdateExpression="SET lock_count = :r",
                ExpressionAttributeValues = { ':r': -1 },
                ConditionExpression=boto3.dynamodb.conditions.Attr('lock_count').eq(0)
            )

    def unlock(self, reset_cluster=False):
        if not self.lock_table:
            self.lock_table = dynamodb.Table(LOCK_DDB_TABLE)
        if not reset_cluster:
            self.lock_table.update_item(
                Key={'cluster': CLUSTER},
                UpdateExpression="SET lock_count = lock_count - :r",
                ExpressionAttributeValues = { ':r': 1 }
            )
        else:
            self.lock_table.update_item(
                Key={'cluster': CLUSTER},
                UpdateExpression="SET lock_count = :r",
                ExpressionAttributeValues = { ':r': 0 },
                ConditionExpression=boto3.dynamodb.conditions.Attr('lock_count').eq(-1)
            )

    def get_lock(self):
        if not self.lock_table:
            self.lock_table = dynamodb.Table(LOCK_DDB_TABLE)
        return self.lock_table.get_item(
            Key={'cluster': CLUSTER},
            ConsistentRead=True
        )['Item']['lock_count']

    def _get_required_params(self):
        spec = yaml.load(open(self.path + '/apb.yml', 'r'))
        params = {}
        for plan in spec['plans']:
            params[plan['name']] = []
            for param in plan['parameters']:
                #print(param)
                if 'required' in param.keys() and 'default' not in param.keys():
                    if param['required'] and param['name'] not in ignore_list:
                        params[plan['name']].append(param['name'])
        return params

    def _parse_ci_config(self):
        return yaml.load(open(self.path + '/ci/config.yml', 'r'))

    def assert_config(self):
        regions = boto3.session.Session().get_available_regions('cloudformation')
        if 'regions' not in self.config['global'].keys():
            raise AssertionError("regions missing from ci config \"global\" section")
        if not set(self.config['global']['regions']).issubset(set(regions)):
            raise AssertionError("global regions list contain invalid regions, accepted regions are: %s" % regions)
        missing = []
        for p in self.config['plans'].keys():
            if p not in self.required.keys():
                raise AssertionError("ci config declares unkown plan \"%s\"" % p)
            else:
                config_params = set(self.config['plans'][p]['parameters'].keys())
                required_params = set(self.required[p])
                if not required_params.issubset(config_params):
                    missing.append([p, required_params ^ config_params])
        if len(missing) > 0 :
            raise AssertionError("not all required parameters are provided in the ci config: %s" % missing)
        return True

    def _get_api_client(self):
        try:
            config.load_kube_config()
        except IOError as e:
            if e.errno != 2:
                raise
            else:
                print("could not find existing kube config")
                traceback.print_exc()
        oapi = client.OapiApi()
        return oapi

    def create_project(self, project_name):
        return self.oapi.create_project(
            {
                "apiVersion": "v1",
                "kind": "Project",
                "metadata": {
                    "name": project_name
                },
                "annotations": {
                    "openshift.io/display-name": "",
                    "openshift.io/description": ""
                }
            }
        )

    def delete_project(self, project_name):
        return self.oapi.delete_project(project_name)

    def _generate_provision_apb_object(self, apb_name, project_name, plan_name, parameters):
        return {
            "apiVersion": "v1",
            "kind": "Template",
            "metadata": {
                "name": apb_name,
                "namespace": project_name
            },
            "objects": [
                {
                    "apiVersion": "servicecatalog.k8s.io/v1beta1",
                    "kind": "ServiceInstance",
                    "metadata": {
                        "name": apb_name,
                        "namespace": project_name
                    },
                    "spec": {
                        "clusterServiceClassExternalName": "dh-%s" % apb_name,
                        "clusterServicePlanExternalName": plan_name,
                        "parameters": parameters
                    }
                }
            ]
        }

    def _generate_binding_object(self, apb_name, project_name):
        return {
            "apiVersion": "servicecatalog.k8s.io/v1beta1",
            "kind": "ServiceBinding",
            "metadata": {
                "name": "%s-binding" % apb_name,
                "namespace": project_name
            },
            "spec": {
                "instanceRef": {
                    "name": apb_name
                }
            }
        }

    def _oc(self, command, follow=None):
        command = ["oc"] + shlex.split(command)
        if not follow:
            return subprocess.check_output(command)
        else:
            regexr = re.compile(follow)
            outp = b""
            process = subprocess.Popen(command, stdout=subprocess.PIPE)
            while True:
                line = process.stdout.readline()
                outp += line
                if regexr.search(line) or line == '':
                    break
            process.kill()
            return outp

    def provision_apb(self, apb_name, project_name, plan_name, parameters):
        # In a rush and haven't figured out how to make calls to the servicecatalog api using the python module,
        # so I'm wrapping the cli for now
        # todo: use the python module for this
        self._oc("project %s" % project_name)
        create_yaml = yaml.dump(self._generate_provision_apb_object(apb_name, project_name, plan_name, parameters))
        fname = '/tmp/' + str(random.randint(0, 10000000000000000)) + '.yml'
        f = open(fname, 'w')
        f.write(create_yaml)
        f.close()
        processed_yaml = self._oc('process -f ' + fname)
        fname = '/tmp/' + str(random.randint(0, 10000000000000000)) + '.yml'
        f = open(fname, 'w')
        f.write(processed_yaml.decode('utf-8'))
        f.close()
        self._oc('create -f ' + fname)
        return yaml.load(self._oc("get ServiceInstance/%s -o yaml" % apb_name))

    def bind_apb(self, apb_name, project_name):
        # In a rush and haven't figured out how to make calls to the servicecatalog api using the python module,
        # so I'm wrapping the cli for now
        # todo: use the python module for this
        self._oc("project %s" % project_name)
        create_yaml = yaml.dump(self._generate_binding_object(apb_name, project_name))
        fname = '/tmp/' + str(random.randint(0, 10000000000000000)) + '.yml'
        f = open(fname, 'w')
        f.write(create_yaml)
        f.close()
        return self._oc('create -f ' + fname)

    def deprovision_apb(self, apb_name, project_name):
        self._oc("project %s" % project_name)
        return self._oc("delete serviceinstance/%s" % apb_name)

    def get_provision_logs(self, apb_name, project_name, wait=True):
        retries = 0
        logs = ""
        while retries < 60:
            for pod in yaml.load(self._oc("get pods -a --all-namespaces -o yaml"))['items']:
                if 'labels' in pod['metadata'].keys():
                    if 'apb-action' in pod['metadata']['labels'] and 'apb-fqname' in pod['metadata']['labels']:
                        if '"namespace":"%s"' % project_name in pod['spec']['containers'][0]['args'][2]:
                            self._oc("project %s" % pod['metadata']['namespace'])
                            if pod['status']['phase'] != 'Pending':
                                if wait:
                                    logs = self._oc(
                                        "logs -f pod/%s" % pod['metadata']['name'],
                                        r'localhost[ ]*: ok=[0-9]*[ ]*changed=[0-9]*[ ]*unreachable=[0-9]*[ ]*failed=[0-9]*[ ]*'
                                    ).decode('utf-8')
                                else:
                                    logs = self._oc("logs -f pod/%s" % pod['metadata']['name']).decode('utf-8')
                                regexpr = re.compile(r'localhost[ ]*: ok=[0-9]*[ ]*changed=[0-9]*[ ]*unreachable=[0-9]*[ ]*failed=[0-9]*[ ]*')
                                status = 'Pending'
                                if regexpr.search(logs):
                                    if re.compile(r'localhost[ ]*: ok=[0-9]*[ ]*changed=[0-9]*[ ]*unreachable=[0-9]*[ ]*failed=0').search(logs):
                                        status = 'Success'
                                    else:
                                        status = 'Failed'
                                msgs = [i for i in logs.split('\n') if i.startswith('    "msg": "stack_suffix: ')]
                                stack_name = None
                                if len(msgs) == 1:
                                    stack_name = "AWSServiceBroker-%s-%s" % (apb_name, yaml.load(yaml.load(msgs[0])['msg'])['stack_suffix'])
                                print("%s successful: " % apb_name)
                                print(logs)
                                return {'status': status, 'logs': logs, 'stack_name': stack_name}
            retries += 1
            sleep(1)
        print("%s failed: " % apb_name)
        print(logs)
        raise Exception("Cannot find a provision pod for serviceinstance \"%s\" in project \"%s\"" % (apb_name, project_name))

    def get_cfn_stack(self, stack_name, profile=None):
        return cfn_client.describe_stacks(StackName=stack_name)["Stacks"][0]

    def login(self, username, password, hostname):
        self._oc("login --insecure-skip-tls-verify %s -u %s -p %s" % (hostname, username, password))
        self.oapi = self._get_api_client()
        return True


if argv[1] in ['resetlock','resetunlock', 'checklock']:
    oapi = AwsApbCi(argv[1], lock_only=True)
    if len(argv) > 2:
        dynamodb = boto3.session.Session(profile_name=argv[2]).resource('dynamodb', region_name=REGION)
    else:
        dynamodb = boto3.resource('dynamodb', region_name=REGION)
    if argv[1] == 'resetlock':
        oapi.lock(reset_cluster=True)
    elif argv[1] == 'resetunlock':
        oapi.unlock(reset_cluster=True)
    elif argv[1] == 'checklock':
        print(oapi.get_lock())
else:
    apb_name = argv[1]
    print("logging into openshift cluster")
    oapi = AwsApbCi(argv[1], argv[2], argv[3], argv[4], argv[5])

    oapi._oc("project aws-service-broker")
    asb_secrets = yaml.load(oapi._oc("get secret awsservicebroker-asb-secret -o yaml"))
    if 'region' in asb_secrets['data'].keys():
        region = base64.decodestring("%s" % asb_secrets['data']['region'])
    else:
        region = REGION
    print("creating boto3 clients")
    if len(argv) > 7:
        cfn_client = boto3.session.Session(profile_name=argv[7]).client('cloudformation', region_name=region)
        s3_client = boto3.session.Session(profile_name=argv[7]).client('s3', region_name=REGION)
        dynamodb = boto3.session.Session(profile_name=argv[7]).resource('dynamodb', region_name=REGION)
    else:
        cfn_client = boto3.client('cloudformation', region_name=region)
        s3_client = boto3.client('s3', region_name=REGION)
        dynamodb = boto3.resource('dynamodb', region_name=REGION)

    outp = []

    for plan_name in oapi.config['plans'].keys():
        print('starting tests for plan %s' % plan_name)
        oapi.lock()
        plan = oapi.config['plans'][plan_name]
        test_id = random.randint(10000000, 99999999)
        project_name = 'cibot-%s-%s' % (apb_name, test_id)
        success = True
        try:
            print('creating project')
            print(oapi.create_project(project_name))
            print('provisioning apb')
            print(oapi.provision_apb(apb_name, project_name, plan_name, plan['parameters']))
            print('getting provision pod logs')
            results = oapi.get_provision_logs(apb_name, project_name)
            assert results['status'] == 'Success', "provision ansible playbook execution failed. logs: %s" % results['logs']
            stack_name = results['stack_name']
            print('getting cfn stack status')
            stack_details = oapi.get_cfn_stack(stack_name)
            stack_id = stack_details['StackId']
            assert stack_details['StackStatus'] == 'CREATE_COMPLETE', "cloudformation stack %s failed to complete successfully" % stack_name
            print('creating binding')
            print(oapi.bind_apb(apb_name, project_name))
            # system user didn't have the right role to deploy the sample-app for some reason, so adding it
            if 'sample_app' in plan.keys():
                print('provisioning sample app')
                print(oapi._oc("adm policy add-role-to-user system:deployer system:serviceaccount:%s:deployer -n %s" % (project_name, project_name)))
                sa_plan = '%s-sample' % apb_name
                if 'sample_app_plan' in plan.keys():
                    sa_plan = plan['sample_app_plan']
                print(oapi.provision_apb(plan['sample_app'], project_name, sa_plan, {}))
                print('waiting for sample app to deploy')
                sleep(15)
                retries = 0
                while True:
                    if "%s-sample-app" % apb_name in oapi._oc("get dc --no-headers"):
                        break
                    retries += 1
                    if retries > 60:
                        raise Exception('Sample app failed to deploy after 5 minutes')
                    sleep(5)
                print(oapi._oc("project %s" % project_name))
                print(oapi._oc("rollout status dc/%s-sample-app" % apb_name))
                print('adding bind secrets as env vars')
                retries = 0
                while True:
                    try:
                        print(oapi._oc("set env --from=secret/%s-binding dc/%s-sample-app" % (apb_name, apb_name)))
                        break
                    except Exception as e:
                        retries += 1
                        print(str(e))
                        traceback.print_exc()
                        if retries == 60:
                            raise
                        sleep(20)
                print('do deployments to ensure env vars are available in the containers')
                print(oapi._oc("rollout status dc/%s-sample-app" % apb_name))
                print(oapi._oc("rollout latest %s-sample-app" % apb_name))
                print(oapi._oc("rollout status dc/%s-sample-app" % apb_name))
                sample_url = 'http://%s-sample-app-%s.%s' % (apb_name, project_name, SHORT_HOSTNAME)
                print('getting sample app html from %s' % sample_url)
                resp = requests.get(sample_url)
                assert resp.status_code == 200, "http request to sample app failed with code: %s" % str(resp.status_code)
                print("uploading sample-app html report to s3")
                s3_client.upload_fileobj(StringIO(resp.text), PROJECT_NAME, 'testoutput/%s.html' % project_name, ExtraArgs={
                    "ACL": "public-read",
                    "ContentType": "text/html",
                    "ContentDisposition": "inline"
                })
                assert '<pre>\nOverall Success: True\n' in resp.text, "Sample app tests failed."
            print('plan %s tests completed successfully' % plan_name)
        except Exception as e:
            success = False
            print('plan %s failed tests' % plan_name)
            print(e)
            traceback.print_exc()
        finally:
            if success:
                print("starting cleanup for %s" % project_name)
                try:
                    print(oapi.deprovision_apb("%s-sample-app" % apb_name, project_name))
                except Exception as e:
                    print(e)
                    traceback.print_exc()
                try:
                    print(oapi._oc("delete ServiceBinding/%s-binding -n %s" % (apb_name, project_name)))
                except Exception as e:
                    print(e)
                    traceback.print_exc()
                try:
                    print(oapi.deprovision_apb(apb_name, project_name))
                except Exception as e:
                    print(e)
                    traceback.print_exc()
                try:
                    cfn_client.delete_stack(StackName=stack_name)
                except Exception as e:
                    print(e)
                    traceback.print_exc()
                try:
                    print(oapi.delete_project(project_name))
                except Exception as e:
                    print(e)
                    traceback.print_exc()
                oapi.unlock()
            else:
                oapi.unlock()
                exit(1)
            outp.append({
                'plan_name': plan_name,
                'report_url': "https://s3.amazonaws.com/%s/testoutput/%s.html" % (PROJECT_NAME, project_name),
                'stack_id': stack_id
            })
    print("uploading results to s3")
    s3_client.upload_fileobj(StringIO(yaml.dump(outp)), PROJECT_NAME, 'testoutput/%s.yml' % argv[6])
    print('All tests completed successfully')
