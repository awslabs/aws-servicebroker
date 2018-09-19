from bottle import get, put, delete, run, request, static_file, template, view, debug
from sampleaws import method_wrapper, get_list, get_item, put_item, delete_item
import traceback
import json
from os import environ

# this file should be able to be used largely as is for any aws apb, code the aws api calls specific to the service in
# the sampleaws.py file


@get('/api/')
def list_method():
    return method_wrapper(get_list)


@get('/api/<item_id>')
def describe_method(item_id="test"):
    return method_wrapper(get_item, item_id=item_id)


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


@get('/test')
def test_method():
    # TODO:
    # * replace hard coded assertions with configurables on response values that may differ from service to service
    # * make function code less ugly and repetitive
    success = True
    response = {
        "cleanup": {"success": False, "data": []},
        "put": {"success": False, "data": ["skipped"]},
        "describe": {"success": False, "data": ["skipped"]},
        "list": {"success": False, "data": ["skipped"]},
        "idempotency": {"success": False, "data": ["skipped"]},
        "delete": {"success": False, "data": ["skipped"]}
    }
    try:
        response['cleanup']["success"] = True
        items = list_method()
        assert items['success'], items['error']
        for item in items['response']:
            response['cleanup']["data"].append("cleaning up item %s" % item)
            delete_method(item_id=item)
        if len(items) == 0:
            response['cleanup']["data"].append("No items to cleanup")
        else:
            response['cleanup']["data"].append("Successfully cleaned up all items")
    except Exception as e:
        tb = traceback.format_exc()
        success = False
        response['cleanup']["success"] = False
        response['cleanup']["data"].append("%s %s\n\n%s" % (str(e.__class__), str(e), tb))
    if success:
        try:
            response['put'] = put_method(
                item_id="test",
                content_type="application/json",
                data='{"content": "test_content"}'
            )
            assert response['put']["success"], response['put']["error"]
            response['put']['data'] = ["Successfully put item"]
        except Exception as e:
            tb = traceback.format_exc()
            success = False
            response['put']["success"] = False
            response['put']["data"] = ["%s %s\n\n%s" % (str(e.__class__), str(e), tb)]
    if success:
        try:
            response['describe'] = describe_method(item_id="test")
            assert response['describe']["success"], response['describe']["error"]
            expected = {
                "item_id": "test",
                "content": "test_content"
            }
            assert response['describe']["response"] == expected, \
                'unexpected response object, expecting: "%s" got: "%s" ' % (
                    json.dumps(expected),
                    json.dumps(response['describe']["response"])
                )
            response['describe']['data'] = ["Successfully described item"]
        except Exception as e:
            tb = traceback.format_exc()
            success = False
            response['describe']["success"] = False
            response['describe']["data"] = ["%s %s\n\n%s" % (str(e.__class__), str(e), tb)]
    if success:
        try:
            response['list'] = list_method()
            assert response['list']["success"], response['list']["error"]
            expected = ["test"]
            assert response['list']["response"] == expected, \
                'unexpected response object, expecting: "%s" got: "%s" ' % (
                    json.dumps(expected),
                    json.dumps(response['list']["response"])
                )
            response['list']['data'] = ["Successfully listed items"]
        except Exception as e:
            tb = traceback.format_exc()
            success = False
            response['list']["success"] = False
            response['list']["data"] = ["%s %s\n\n%s" % (str(e.__class__), str(e), tb)]
    if success:
        try:
            response['idempotency']["success"] = True
            response['idempotency']["data"] = []
            put_method(
                item_id="test",
                content_type="application/json",
                data='{"content": "test_content_idempotent"}'
            )
            response['idempotency']['data'].append("Put an additional item with duplicate item_id")
            describe = describe_method(item_id="test")
            assert describe["success"], describe["error"]
            expected = {
                "item_id": "test",
                "content": "test_content_idempotent"
            }
            assert describe["response"] == expected, \
                'unexpected describe response object, expecting: "%s" got: "%s" ' % (
                    json.dumps(expected),
                    json.dumps(describe["response"])
                )
            response['idempotency']['data'].append("Verified describe has updated content")
            list_items = list_method()
            assert list_items["success"], list_items["error"]
            expected = ["test"]
            assert list_items["response"] == expected, \
                'unexpected list response object, expecting: "%s" got: "%s" ' % (
                    json.dumps(expected),
                    json.dumps(list_items["response"])
                )
            response['idempotency']['data'].append("Verified list does not contain duplicates")
            response['idempotency']['data'].append("Successfully tested idempotency")
        except Exception as e:
            tb = traceback.format_exc()
            success = False
            response['idempotency']["success"] = False
            response['idempotency']["data"].append("%s %s\n\n%s" % (str(e.__class__), str(e), tb))
    if success:
        try:
            response['delete'] = delete_method(item_id="test")
            assert response['delete']["success"], response['delete']["error"]
            response['delete']['data'] = ["Successfully deleted item"]
        except Exception as e:
            tb = traceback.format_exc()
            success = False
            response['delete']["success"] = False
            response['delete']["data"] = ["%s %s\n\n%s" % (str(e.__class__), str(e), tb)]
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
