from os import environ
import traceback
import json
from pymemcache.client.hash import HashClient


SECRETS = [
    'ELASTICACHE_ENDPOINT_ADDRESS',
    'ELASTICACHE_PORT_NUMBER'
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables. ENVIRONMENT: %s" % str(environ)


def create_memcached_client():
    return HashClient([(environ['ELASTICACHE_ENDPOINT_ADDRESS'], int(environ['ELASTICACHE_PORT_NUMBER']))])


def method_wrapper(method, **kwargs):
    try:
        check_secrets()
        kwargs['client'] = create_memcached_client()
        return {"success": True, 'response': method(**kwargs)}
    except Exception as e:
        tb = traceback.format_exc()
        return {"success": False, "error":  "%s %s\n\n%s" % (str(e.__class__), str(e), tb)}


def get_list(**kwargs):
    # list is not really something you do with elasticache, so faking this with a get for 'test' key
    if kwargs['client'].get('test'):
        return ['test']
    else:
        return []


def get_item(item_id, **kwargs):
    item = kwargs['client'].get(item_id)
    if type(item) == bytes:
        item = item.decode()
    try:
        item = json.loads(item)
        item['item_id'] = item_id
    except:
        item = {'item_id': item_id, 'content': item}
    return item


def delete_item(item_id, **kwargs):
    kwargs['client'].delete(item_id)
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

    if kwargs['client'].get(item_id):
        kwargs['client'].replace(item_id, data['content'])
    else:
        kwargs['client'].add(item_id, data['content'])
    return
