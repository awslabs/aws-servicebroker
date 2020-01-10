import cfnresponse
import string
import boto3
import random
import netaddr
import traceback


alnum = string.ascii_uppercase + string.ascii_lowercase + string.digits
ec2_client = boto3.resource('ec2')


def get_cidrs(size, qty, vpc_id):
    # TODO: add locking mechanism to prevent collisions when run concurrently
    vpc = ec2_client.Vpc(vpc_id)
    allocated_cidrs = netaddr.IPSet([s.cidr_block for s in vpc.subnets.all()])
    unused_cidrs = netaddr.IPSet([vpc.cidr_block]) ^ allocated_cidrs
    available_cidrs = []
    for sl in [list(s.subnet(size)) for s in unused_cidrs.iter_cidrs()]:
        available_cidrs = available_cidrs + sl
    if len(available_cidrs) < qty:
        raise Exception("Not enough available space in the vpc\n")
    return [str(s) for s in available_cidrs[-(qty):]]


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
            response_data['CidrBlocks'] = get_cidrs(
                int(event['ResourceProperties']['CidrSize']),
                int(event['ResourceProperties']['Qty']),
                event['ResourceProperties']['VpcId']
            )
        cfnresponse.send(event, context, response_code, response_data, phys_id)
    except Exception as e:
        print(str(e))
        traceback.print_exc()
        cfnresponse.send(event, context, cfnresponse.FAILED, response_data, phys_id, str(e))
