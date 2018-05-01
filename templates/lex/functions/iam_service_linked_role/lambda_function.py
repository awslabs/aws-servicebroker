import json
import logging
import threading
import boto3
import cfnresponse


iam_client = boto3.client('iam')


def create_role(service_name, custom_suffix=None):
    try:
        if custom_suffix:
            role = iam_client.create_service_linked_role(AWSServiceName='%s.amazonaws.com' % service_name,
                                                         CustomSuffix=custom_suffix)['Role']
        else:
            role = iam_client.create_service_linked_role(AWSServiceName='%s.amazonaws.com' % service_name)['Role']
    except iam_client.exceptions.InvalidInputException as e:
        if ' has been taken in this account, please try a different suffix.' in str(e):
            role = iam_client.get_role(RoleName=e.response['Error']['Message'].split(' ')[3])['Role']
        else:
            raise
    return role['Arn']


def delete_role(arn):
    try:
        role_name = arn.split('/')[-1]
        iam_client.delete_service_linked_role(RoleName=role_name)
    except iam_client.exceptions.NoSuchEntityException as e:
        if 'Cannot find the service role to delete.' not in str(e):
            raise


def timeout(event, context):
    logging.error('Execution is about to time out, sending failure response to CloudFormation')
    cfnresponse.send(event, context, cfnresponse.FAILED, {}, None)


def handler(event, context):
    timer = threading.Timer((context.get_remaining_time_in_millis() / 1000.00) - 0.5, timeout, args=[event, context])
    timer.start()
    print('Received event: %s' % json.dumps(event))
    status = cfnresponse.SUCCESS
    reason = None
    physical_id = None
    try:
        service_name = event['ResourceProperties']['ServiceName']
        if 'CustomSuffix' in event['ResourceProperties'].keys():
            custom_suffix = event['ResourceProperties']['CustomSuffix']
        else:
            custom_suffix = None
        if event['RequestType'] != 'Delete':
            physical_id = create_role(service_name, custom_suffix)
        else:
            physical_id = event['PhysicalResourceId']
            delete_role(physical_id)
    except Exception as e:
        logging.error('Exception: %s' % e, exc_info=True)
        status = cfnresponse.FAILED
        reason = str(e)
    finally:
        timer.cancel()
        cfnresponse.send(event, context, status, {'Arn': physical_id}, physical_id, reason)
