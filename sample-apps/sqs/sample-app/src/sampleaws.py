from os import environ
import traceback
import json
import boto3
import botocore

SECRETS = [
    'SQS_QUEUE_ARN',
    'SQS_QUEUE_NAME',
    'SQS_QUEUE_URL',
    'SQS_AWS_ACCESS_KEY_ID',
    'SQS_AWS_SECRET_ACCESS_KEY',
    'SQS_REGION'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_client():
    return boto3.client(
        "sqs",
        aws_access_key_id=environ['SQS_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['SQS_AWS_SECRET_ACCESS_KEY'],
        region_name=environ['SQS_REGION']
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
    messages = kwargs['client'].receive_message(
        QueueUrl=environ['SQS_QUEUE_URL'],
        MaxNumberOfMessages=10,
        VisibilityTimeout=0,
        WaitTimeSeconds=1,
        MessageAttributeNames=['item_id']
    )
    if 'Messages' not in messages.keys():
        return []
    return [ [m['MessageAttributes']['item_id']['StringValue'], m['ReceiptHandle']] for m in messages['Messages'] ]


def get_item(item_id, **kwargs):
    item = kwargs['client'].receive_message(
        QueueUrl=environ['SQS_QUEUE_URL'],
        MaxNumberOfMessages=1,
        VisibilityTimeout=0,
        WaitTimeSeconds=1,
        MessageAttributeNames=['item_id']
    )['Messages'][0]
    return {
        'content': item['Body'],
        'item_id': item['MessageAttributes']['item_id']['StringValue']
    }


def delete_item(item_id, **kwargs):
    handles = [i[1] for i in get_list(client=kwargs['client']) if i[0] == item_id]
    for h in handles:
        try:
            kwargs['client'].delete_message(
                QueueUrl=environ['SQS_QUEUE_URL'],
                ReceiptHandle=h
            )
        except botocore.exceptions.ClientError as e:
            if e.response['Error']['Message'].endswith('The receipt handle has expired.'):
                pass
            else:
                raise
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

    kwargs['client'].send_message(
        QueueUrl=environ['SQS_QUEUE_URL'],
        MessageBody=data['content'],
        DelaySeconds=0,
        MessageAttributes={
            'item_id': {
                'StringValue': item_id,
                'DataType': 'String'
            }
        },
        MessageGroupId='testgroup'
    )
    return
