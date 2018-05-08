from os import environ
import traceback
import json
import boto3


SECRETS = [
    'ROUTE53_AWS_SECRET_ACCESS_KEY',
    'ROUTE53_AWS_ACCESS_KEY_ID',
    'ROUTE53_HOSTED_ZONE_NAME',
    'ROUTE53_HOSTED_ZONE_ID'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_client():
    return boto3.client(
        "route53",
        aws_access_key_id=environ['ROUTE53_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['ROUTE53_AWS_SECRET_ACCESS_KEY']
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
    response = kwargs['client'].list_resource_record_sets(
        HostedZoneId=environ['ROUTE53_HOSTED_ZONE_ID']
    )['ResourceRecordSets']
    return [ i['Name'] for i in response if i['Type'] == "TXT" ]


def get_item(item_id, **kwargs):
    item_name = "%s.%s" % (item_id, environ['ROUTE53_HOSTED_ZONE_NAME'])
    response = kwargs['client'].list_resource_record_sets(
        HostedZoneId=environ['ROUTE53_HOSTED_ZONE_ID'],
        StartRecordName=item_name
    )['ResourceRecordSets']
    try:
        item = [ i for i in response if i['Type'] == "TXT" and i['Name'] == "%s." % item_name ][0]
    except:
        raise Exception("Failed to get item from response: %s" % str([response]))
    item['item_id'] = item_id
    return item


def delete_item(item_id, **kwargs):
    item = get_item(item_id, client=kwargs['client'])
    kwargs['client'].change_resource_record_sets(
        HostedZoneId=environ['ROUTE53_HOSTED_ZONE_ID'],
        ChangeBatch={'Changes': [{
            'Action': 'DELETE',
            'ResourceRecordSet': {
                'Name': item['Name'],
                'Type': item['Type'],
                'TTL': item['TTL'],
                'ResourceRecords': item['ResourceRecords']
            }
        }]}
    )
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

    kwargs['client'].change_resource_record_sets(
        HostedZoneId=environ['ROUTE53_HOSTED_ZONE_ID'],
        ChangeBatch={'Changes': [{
            'Action': 'UPSERT',
            'ResourceRecordSet': {
                'Name': "%s.%s" % (item_id, environ['ROUTE53_HOSTED_ZONE_NAME']),
                'Type': 'TXT',
                'TTL': 1337,
                'ResourceRecords': [
                    {
                        'Value': '"%s"' % data['content']
                    }
                ]
            }
        }]}
    )
    return
