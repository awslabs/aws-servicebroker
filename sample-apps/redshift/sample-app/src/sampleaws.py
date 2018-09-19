import psycopg2
import psycopg2.extras
from os import environ
import traceback
import json


SECRETS = [
    "REDSHIFT_ENDPOINT_ADDRESS",
    "REDSHIFT_PORT",
    "REDSHIFT_MASTER_USERNAME",
    "REDSHIFT_MASTER_USER_PASSWORD",
    "REDSHIFT_DB_NAME"
]


def check_secrets():
    assert set(SECRETS).issubset(set(environ)), "Required secrets are not present in environment variables"


def check_port():
    assert environ['REDSHIFT_PORT'].isdigit(), "Port number is not a number"
    assert int(environ['REDSHIFT_PORT']) <= 65535, "port number is greater that 65535"


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
        "select count(item_id) as c from sampleapp where item_id = %s", (item_id,)
    )
    if kwargs['db_cursor'].fetchone()['c'] == 0:
        kwargs['db_cursor'].execute(
            "insert into sampleapp (item_id, content) VALUES(%s, %s)",
            (item_id, data['content'],)
        )
    else:
        kwargs['db_cursor'].execute(
            "update sampleapp set content=%s where item_id=%s",
            (data['content'], item_id,)
        )
    kwargs['db_conn'].commit()


def get_db():
    conn = psycopg2.connect(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode='require'" % (
            environ["REDSHIFT_ENDPOINT_ADDRESS"],
            environ["REDSHIFT_PORT"],
            environ["REDSHIFT_MASTER_USERNAME"],
            environ["REDSHIFT_MASTER_USER_PASSWORD"],
            environ["REDSHIFT_DB_NAME"]
        ),
        cursor_factory=psycopg2.extras.RealDictCursor
    )
    return conn, conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)


def init_db(db_conn, db_cursor):
    db_cursor.execute(
        "create table if not exists sampleapp (item_id varchar(256) primary key, content varchar(2048))"
    )
    db_conn.commit()
    return None
