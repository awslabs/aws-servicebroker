import json
import logging
import threading
import boto3
import cfnresponse


def copy_objects(source_bucket, dest_bucket, prefix, objects):
    s3 = boto3.client('s3')
    for o in objects:
        key = prefix + o
        copy_source = {
            'Bucket': source_bucket,
            'Key': key
        }
        print('copy_source: %s' % copy_source)
        print('dest_bucket = %s' % dest_bucket)
        print('key = %s' % key)
        s3.copy_object(CopySource=copy_source, Bucket=dest_bucket, Key=key)


def delete_objects(bucket, prefix, objects):
    s3 = boto3.client('s3')
    objects = {'Objects': [{'Key': prefix + o} for o in objects]}
    s3.delete_objects(Bucket=bucket, Delete=objects)


def timeout(event, context):
    logging.error('Execution is about to time out, sending failure response to CloudFormation')
    cfnresponse.send(event, context, cfnresponse.FAILED, {}, None)


def handler(event, context):
    timer = threading.Timer((context.get_remaining_time_in_millis() / 1000.00) - 0.5, timeout, args=[event, context])
    timer.start()
    print('Received event: %s' % json.dumps(event))
    status = cfnresponse.SUCCESS
    try:
        source_bucket = event['ResourceProperties']['SourceBucket']
        dest_bucket = event['ResourceProperties']['DestBucket']
        prefix = event['ResourceProperties']['Prefix']
        objects = event['ResourceProperties']['Objects']
        if event['RequestType'] == 'Delete':
            delete_objects(dest_bucket, prefix, objects)
        else:
            copy_objects(source_bucket, dest_bucket, prefix, objects)
    except Exception as e:
        logging.error('Exception: %s' % e, exc_info=True)
        status = cfnresponse.FAILED
    finally:
        timer.cancel()
        cfnresponse.send(event, context, status, {}, None)
