from os import environ
import traceback
import json
import boto3
from io import BytesIO

SECRETS = [
    'S3_BUCKET_NAME',
    'S3_REGION',
    'S3_BUCKET_ARN',
    'S3_AWS_ACCESS_KEY_ID',
    'S3_AWS_SECRET_ACCESS_KEY'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_bucket_resource():
    return boto3.resource(
        "s3",
        aws_access_key_id=environ['S3_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['S3_AWS_SECRET_ACCESS_KEY'],
        region_name=environ['S3_REGION']
    ).Bucket(environ['S3_BUCKET_NAME'])


def method_wrapper(method, **kwargs):
    try:
        check_secrets()
        kwargs['bucket'] = create_boto3_bucket_resource()
        return {"success": True, 'response': method(**kwargs)}
    except Exception as e:
        tb = traceback.format_exc()
        return {"success": False, "error":  "%s %s\n\n%s" % (str(e.__class__), str(e), tb)}


def get_list(**kwargs):
    return [i.key for i in kwargs['bucket'].objects.all()]


def get_item(item_id, **kwargs):
    data = BytesIO()
    kwargs['bucket'].download_fileobj(item_id, data)
    data.seek(0)
    return {"item_id": item_id, "content": data.read().decode()}


def delete_item(item_id, **kwargs):
    kwargs['bucket'].delete_objects(Delete={
        'Objects': [
            {
                'Key': item_id
            },
        ]
    })
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

    kwargs['bucket'].put_object(Key=item_id, Body=data['content'].encode())
    return
