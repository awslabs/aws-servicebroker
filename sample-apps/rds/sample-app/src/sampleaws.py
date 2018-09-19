import pymysql
from os import environ
import traceback
import json


SECRETS = [
    "RDS_MASTER_USERNAME",
    "RDS_PORT",
    "RDS_ENDPOINT_ADDRESS",
    "RDS_DB_NAME",
    "RDS_MASTER_USER_PASSWORD",
    "RDS_ENGINE"
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables"


def check_port():
    assert environ['RDS_PORT'].isdigit(), "Port number is not a number"
    assert int(environ['RDS_PORT']) <= 65535, "port number is greater that 65535"


def method_wrapper(method, **kwargs):
    try:
        check_secrets()
        check_port()
        kwargs['db_conn'], kwargs['db_cursor'] = get_db()
        init_db(kwargs['db_conn'], kwargs['db_cursor'])
        return {"success": True, 'response': method(**kwargs)}
    except Exception as e:
        traceback.print_exc()
        return {"success": False, "error":  "%s %s" % (str(e.__class__), str(e))}


def get_list(**kwargs):
    kwargs['db_cursor'].execute("select item_id from sampleapp")
    return [i['item_id'] for i in kwargs['db_cursor'].fetchall()]


def get_item(item_id, **kwargs):
    kwargs['db_cursor'].execute("select * from sampleapp where item_id = %s", (item_id,))
    return kwargs['db_cursor'].fetchone()


def delete_item(item_id, **kwargs):
    kwargs['db_cursor'].execute("delete from sampleapp where item_id = %s", (item_id,))
    kwargs['db_conn'].commit()


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
    kwargs['db_cursor'].execute(
        "insert into sampleapp (item_id, content) VALUES(%s, %s) on duplicate key update content=%s",
        (item_id, data['content'], data['content'],)
    )
    kwargs['db_conn'].commit()


def get_db():
    if environ["RDS_ENGINE"] == 'mysql':
        conn = pymysql.connect(
            host=environ["RDS_ENDPOINT_ADDRESS"],
            port=int(environ["RDS_PORT"]),
            user=environ["RDS_MASTER_USERNAME"],
            passwd=environ["RDS_MASTER_USER_PASSWORD"],
            db=environ["RDS_DB_NAME"]
        )
        return conn, conn.cursor(pymysql.cursors.DictCursor)
    else:
        raise Exception("Database engine %s is unsupported" % environ["RDS_ENGINE"])


def init_db(db_conn, db_cursor):
    db_cursor.execute(
        "create table if not exists sampleapp (item_id varchar(256) primary key, content varchar(2048))"
    )
    db_conn.commit()
    return None
