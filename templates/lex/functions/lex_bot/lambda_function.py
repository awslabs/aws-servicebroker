import json
import logging
import threading
import boto3
import cfnresponse
import time


lex_client = boto3.client('lex-models')
s3_client = boto3.client('s3')


def create_bot(intent):
    lex_client.put_bot(**intent)


def delete_bot(name, retries=0):
    try:
        lex_client.delete_bot(name=name)
    except lex_client.exceptions.NotFoundException as e:
        if 'does not exist. Choose another resource.' not in str(e):
            raise
    except lex_client.exceptions.ConflictException as e:
        if 'There is a conflicting operation in progress' in str(e):
            if retries > 10:
                raise
            else:
                retries+=1
                time.sleep(retries*3)
                delete_bot(name, retries)

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
        data = s3_client.get_object(Bucket=event['ResourceProperties']['Bucket'],
                                    Key=event['ResourceProperties']['Key'])['Body'].read()
        try:
            bot = json.loads(data)
        except Exception as e:
            logging.error('Exception: %s' % e, exc_info=True)
            raise Exception('Intent json is malformed')
        if event['RequestType'] != 'Delete':
            create_bot(bot)
            physical_id = bot['name']
        else:
            delete_bot(event['PhysicalResourceId'])
    except Exception as e:
        logging.error('Exception: %s' % e, exc_info=True)
        status = cfnresponse.FAILED
        reason = str(e)
    finally:
        timer.cancel()
        cfnresponse.send(event, context, status, {}, physical_id, reason)
