import cfnresponse
import random
import string
import boto3
import traceback

alnum = string.ascii_uppercase + string.ascii_lowercase + string.digits
iam_client = boto3.client('iam')
ssm_client = boto3.client('ssm')


def handler(event, context):
    response_code = cfnresponse.SUCCESS
    response_data = {}
    if event['RequestType'] == 'Create':
        phys_id = ''.join(random.choice(alnum) for _ in range(16))
    else:
        phys_id = event['PhysicalResourceId']
    response_data['asb-access-key-id'] = 'asb-access-key-id-%s' % phys_id
    response_data['asb-secret-access-key'] = 'asb-secret-access-key-%s' % phys_id
    try:
        username = event['ResourceProperties']['Username']
        if event['RequestType'] == 'Create':
            response = iam_client.create_access_key(UserName=username)
            aws_access_key_id = response['AccessKey']['AccessKeyId']
            secret_access_key = response['AccessKey']['SecretAccessKey']
            ssm_client.put_parameter(Name=response_data['asb-access-key-id'], Value=aws_access_key_id, Type='SecureString')
            ssm_client.put_parameter(Name=response_data['asb-secret-access-key'], Value=secret_access_key, Type='SecureString')
        elif event['RequestType'] == 'Update':
            print('Update operation unsupported')
            response_code = cfnresponse.FAILED
        elif event['RequestType'] == 'Delete':
            for access_key in iam_client.list_access_keys(UserName=username)['AccessKeyMetadata']:
                iam_client.delete_access_key(UserName=username, AccessKeyId=access_key['AccessKeyId'])
            ssm_client.delete_parameters(Names=[response_data['asb-access-key-id'], response_data['asb-secret-access-key']])
        cfnresponse.send(event, context, response_code, response_data, phys_id)
    except Exception as e:
        print(str(e))
        traceback.print_exc()
        cfnresponse.send(event, context, cfnresponse.FAILED, response_data, phys_id)
