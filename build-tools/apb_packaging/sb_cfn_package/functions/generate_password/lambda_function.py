import cfnresponse
import random
import secrets
import string
import traceback


alnum = string.ascii_uppercase + string.ascii_lowercase + string.digits
alphabet = string.ascii_lowercase
upperalphabet = alphabet.upper()
digits = string.digits
special_characters = "-;:,./<>?[]{}_=+~!@%^()"
full_alphabet = alphabet + special_characters


def generate_password(pw_len):
    pwlist = []
    for i in range(1):
        pwlist.append(secrets.choice(alnum))
        pwlist.append(secrets.choice(upperalphabet))
        pwlist.append(secrets.choice(digits))
        pwlist.append(secrets.choice(special_characters))
    for i in range(pw_len - len(pwlist)):
        pwlist.append(secrets.choice(full_alphabet))
    random.shuffle(pwlist)
    return "".join(pwlist)


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
            if 'Length' in event['ResourceProperties']:
                pw_len = int(event['ResourceProperties']['Length'])
            else:
                pw_len = 16
            response_data['MasterUserPassword'] = generate_password(pw_len)
        cfnresponse.send(event, context, response_code, response_data, phys_id)
    except Exception as e:
        print(str(e))
        traceback.print_exc()
        cfnresponse.send(event, context, cfnresponse.FAILED, response_data, phys_id, str(e))
