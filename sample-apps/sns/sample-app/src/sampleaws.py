from os import environ
import traceback
import json
import boto3
import subprocess

SECRETS = [
    'SNS_TOPIC_ARN',
    'SNS_AWS_SECRET_ACCESS_KEY',
    'SNS_AWS_ACCESS_KEY_ID',
    'SNS_AWS_REGION'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_client():
    return boto3.client(
        "sns",
        aws_access_key_id=environ['SNS_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['SNS_AWS_SECRET_ACCESS_KEY'],
        region_name=environ['SNS_AWS_REGION']
    )


def method_wrapper(method, **kwargs):
    try:
        check_secrets()
        kwargs['client'] = create_boto3_client()
        return {"success": True, 'response': method(**kwargs)}
    except Exception as e:
        tb = traceback.format_exc()
        return {"success": False, "error":  "%s %s\n\n%s" % (str(e.__class__), str(e), tb)}


def get_list(**kwargs):
    return


def get_item(item_id, **kwargs):
    return


def delete_item(item_id, **kwargs):
    return

def put_item(item_id, **kwargs):
    assert kwargs['content_type'] == 'application/json', "PUT must have a json contentType"
    assert kwargs['data'] != None, "PUT data empty"
    try:
        print(kwargs['data'])
        data = json.loads(kwargs['data'])
    except Exception:
        traceback.print_exc()
        raise Exception("put data is not valid json")
    assert 'content' in data, 'missing key "content" in put data'

    kwargs['client'].publish(TopicArn=environ['SNS_TOPIC_ARN'], Message=kwargs['data'])
    return
