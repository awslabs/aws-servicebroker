from bottle import get, put, delete, post, run, request, static_file, template, view, debug
from sampleaws import method_wrapper, get_list, get_item, put_item, delete_item
import traceback
import json
from os import environ
from botocore.vendored import requests

# this file should be able to be used largely as is for any aws apb, code the aws api calls specific to the service in
# the sampleaws.py file


@get('/api/')
def list_method():
    return method_wrapper(get_list)


@get('/api/<item_id>')
def describe_method(item_id="test", protocol='email'):
    return method_wrapper(get_item, item_id=item_id, protocol=protocol)


@delete('/api/<item_id>')
def delete_method(item_id="test"):
    return method_wrapper(delete_item, item_id=item_id)


@put('/api/<item_id>')
def put_method(item_id="test", data=None, content_type=None):
    if not data:
        data = request.body.read().decode('utf-8')
    if not content_type:
        content_type = request.content_type
    return method_wrapper(
        put_item,
        item_id=item_id,
        data=data,
        content_type=content_type
    )

@post('/api/')
def post_method(data=None, content_type=None):
    if not data:
        data = request.body.read().decode('utf-8')
    if not content_type:
        content_type = request.content_type
    if 'x-amz-sns-message-type' not in request.headers.keys():
        raise Exception('missing headers')
    if request.headers['x-amz-sns-message-type'] != 'SubscriptionConfirmation':
        return
    url = json.loads(data)['SubscribeURL']
    requests.get(url)
    return

@get('/test')
def test_method():
    # TODO:
    # * replace hard coded assertions with configurables on response values that may differ from service to service
    # * make function code less ugly and repetitive
    success = True
    response = {
        "put": {"success": False, "data": ["skipped"]}
    }
    try:
        response['put'] = put_method(
            item_id="http://%s/api/" % environ['HOSTNAME'],
            content_type="application/json",
            data='{"content": {"protocol": "http"}}'
        )
        assert response['put']["success"], response['put']["error"]
        response['put']['data'] = ["Successfully put item"]
    except Exception as e:
        tb = traceback.format_exc()
        success = False
        response['put']["success"] = False
        response['put']["data"] = ["%s %s\n\n%s" % (str(e.__class__), str(e), tb)]
    return {"success": success, "data": response}


@get('/')
@view('index')
def index():
    return test_method()


# For Static files
@get("/static/css/<filename:re:.*\.css>")
def css(filename):
    return static_file(filename, root="static/css")


@get("/static/font/<filename:re:.*\.(eot|otf|svg|ttf|woff|woff2?)>")
def font(filename):
    return static_file(filename, root="static/font")


@get("/static/img/<filename:re:.*\.(jpg|png|gif|ico|svg)>")
def img(filename):
    return static_file(filename, root="static/img")


@get("/static/js/<filename:re:.*\.js>")
def js(filename):
    return static_file(filename, root="static/js")


if __name__ == '__main__':
    run(host='0.0.0.0', port=8080, debug=True, reloader=True)
