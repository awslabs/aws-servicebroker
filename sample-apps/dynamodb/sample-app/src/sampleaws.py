from os import environ
import traceback
import json
import boto3


SECRETS = [
    'TABLE_NAME',
    'TABLE_ARN',
    'DYNAMODB_AWS_ACCESS_KEY_ID',
    'DYNAMODB_AWS_SECRET_ACCESS_KEY'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_table_resource():
    return boto3.resource(
        "dynamodb",
        aws_access_key_id=environ['DYNAMODB_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['DYNAMODB_AWS_SECRET_ACCESS_KEY'],
        region_name=environ['TABLE_ARN'].split(':')[3]
    ).Table(environ['TABLE_NAME'])


def method_wrapper(method, **kwargs):
    try:
        check_secrets()
        kwargs['table'] = create_boto3_table_resource()
        if 'range_id' not in kwargs.keys() and 'item_id' in kwargs.keys():
            kwargs['range_id'] = kwargs['item_id']
        return {"success": True, 'response': method(**kwargs)}
    except Exception as e:
        tb = traceback.format_exc()
        return {"success": False, "error":  "%s %s\n\n%s" % (str(e.__class__), str(e), tb)}


def get_list(**kwargs):
    return [[i['test'], i['test2']] for i in kwargs['table'].scan()['Items']]


def get_item(item_id, **kwargs):
    item = kwargs['table'].get_item(Key={
        'test': item_id,
        'test2': kwargs['range_id']
    })['Item']
    item['item_id'] = item_id
    return item


def delete_item(item_id, **kwargs):
    kwargs['table'].delete_item(Key={
        'test': item_id,
        'test2': kwargs['range_id']
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

    kwargs['table'].put_item(Item={
        environ['DYNAMODB_HASH_ATTRIBUTE_NAME']: item_id,
        environ['DYNAMODB_RANGE_ATTRIBUTE_NAME']: kwargs['range_id'],
        "content": data['content']
    })
    return
