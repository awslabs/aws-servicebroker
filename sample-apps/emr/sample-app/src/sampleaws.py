from os import environ
import traceback
import json
import boto3


SECRETS = [
    'EMR_DATA_BUCKET',
    'EMR_APPLICATIONS',
    'EMR_REGION',
    'EMR_MASTER_PUBLIC_DNS',
    'EMR_AWS_ACCESS_KEY_ID',
    'EMR_AWS_SECRET_ACCESS_KEY',
    'EMR_CLUSTER_ID'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_boto3_client():
    return boto3.client(
        "emr",
        aws_access_key_id=environ['EMR_AWS_ACCESS_KEY_ID'],
        aws_secret_access_key=environ['EMR_AWS_SECRET_ACCESS_KEY'],
        region_name=environ['EMR_REGION']
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
    ### EMR only allows cancel for steps in pending state, so lets only list cancellable(deletable) items
    return [[i['Name'], i['Id']] for i in kwargs['client'].list_steps(
        ClusterId=environ['EMR_CLUSTER_ID'],
        StepStates=['PENDING']
    )['Steps']]


def get_item(item_id, **kwargs):
    id = [ i[1] for i in get_list(client=kwargs['client']) if i[0] == item_id ][0]
    item = kwargs['client'].describe_step(ClusterId=environ['EMR_CLUSTER_ID'], StepId=id)['Step']
    item['item_id'] = item_id
    return item


def delete_item(item_id, **kwargs):
    ### EMR only provides a cancel option, no delete, so this tries to cancel a step
    id = [ i[1] for i in get_list(client=kwargs['client']) if i[0] == item_id ][0]
    kwargs['client'].cancel_steps(ClusterId=environ['EMR_CLUSTER_ID'], StepIds=[id])
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
    if len([ i[1] for i in get_list(client=kwargs['client']) if i[0] == item_id ]) > 0:
        delete_item(item_id, client=kwargs['client'])
    kwargs['client'].add_job_flow_steps(
        JobFlowId=environ['EMR_CLUSTER_ID'],
        Steps=[{
            'Name': item_id,
            'ActionOnFailure': 'CONTINUE',
            'HadoopJarStep': {
                'Jar': 'command-runner.jar',
                'Args': [
                    "spark-submit",
                    "--executor-memory",
                    "1g",
                    "--class",
                    "org.apache.spark.examples.SparkPi",
                    "/usr/lib/spark/lib/spark-examples.jar",
                    "10"
                ]
            }
        }]
    )
    return
