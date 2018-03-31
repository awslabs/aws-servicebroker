import cfnresponse
import string
import boto3
import random
import traceback


alnum = string.ascii_uppercase + string.ascii_lowercase + string.digits
ec2_client = boto3.client('ec2')


def get_azs(qty):
    azs = ec2_client.describe_availability_zones(Filters=[{'Name': 'state', 'Values': ['available']}])[
        'AvailabilityZones']
    if len(azs) < qty:
        raise Exception('Insufficient availability zones in this region\n')
    azs = [az['ZoneName'] for az in azs]
    return random.sample(azs, qty)


def handler(event, context):
    response_code = cfnresponse.SUCCESS
    response_data = {}
    print(event)
    if event['RequestType'] == 'Create':
        phys_id = ''.join(random.choice(alnum) for _ in range(16))
    else:
        phys_id = event['PhysicalResourceId']
    try:
        if event['RequestType'] in ['Create', 'Update']:
            response_data['AvailabilityZones'] = get_azs(int(event['ResourceProperties']['Qty']))
        cfnresponse.send(event, context, response_code, response_data, phys_id)
    except Exception as e:
        print(str(e))
        traceback.print_exc()
        cfnresponse.send(event, context, cfnresponse.FAILED, response_data, phys_id, str(e))
