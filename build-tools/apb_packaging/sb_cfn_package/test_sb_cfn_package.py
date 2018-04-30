import unittest
import os
from collections import OrderedDict
import datetime

try:
    from sb_cfn_package import SbCfnPackage, cli
except:
    from sb_cfn_package.sb_cfn_package import SbCfnPackage, cli
try:
    from aws_servicebroker_spec import AwsServiceBrokerSpec, AwsServiceBrokerSpecException
except:
    from sb_cfn_package.aws_servicebroker_spec import AwsServiceBrokerSpec, AwsServiceBrokerSpecException

class Tests(unittest.TestCase):

    def test_build_artifacts(self):
        sb_pack = SbCfnPackage(os.path.dirname(os.path.abspath(__file__)) + "/../sample.yaml")
        outp = sb_pack.build_artifacts("test")
        outp['template']['Resources']['AWSSBInjectedCopyZips']['Properties']['SourceBucket'] = ''
        outp['template']['Resources']['AWSSBInjectedCopyZipsRole']['Properties']['Policies'][0]['PolicyDocument']['Statement'][0]['Resource'][0] = ''
        self.maxDiff = None
        self.assertNotEqual(None, outp)

    def test_AwsServiceBrokerSpecException_multiple_errors(self):
        error = None
        try:
            raise AwsServiceBrokerSpecException(missing_key='some_key', incorrect_type=str)
        except Exception as e:
            error = e
        self.assertEqual(Exception, type(error))
        self.assertEqual(str(error), 'Cannot specify 2 error types in one exception')

    def test_AwsServiceBrokerSpecException_missing_key(self):
        error = None
        try:
            raise AwsServiceBrokerSpecException(missing_key='some_key')
        except Exception as e:
            error = e
        self.assertEqual(AwsServiceBrokerSpecException, type(error))
        self.assertEqual(str(error), 'The AWS Service broker specification requires a key at some_key')

    def test_AwsServiceBrokerSpecException_incorrect_type(self):
        error = None
        try:
            raise AwsServiceBrokerSpecException(incorrect_type='some_type')
        except Exception as e:
            error = e
        self.assertEqual(AwsServiceBrokerSpecException, type(error))
        self.assertEqual(str(error), 'AWS Service broker specification incorrect Type some_type')

    def test_incorrect_mapping_type(self):
        sb_spec = AwsServiceBrokerSpec('test')
        self.assertRaises(AwsServiceBrokerSpecException, sb_spec.get_mapping, key=None, key_type='garbage')

    def test_empty_list_mapping(self):
        sb_spec = AwsServiceBrokerSpec('test')
        self.assertEqual(None, sb_spec.get_mapping(key=[]))

    def test_reverse_mapping(self):
        sb_spec = AwsServiceBrokerSpec('test')
        self.assertEqual('Version', sb_spec.get_mapping(key='version', key_type='apb'))

    def test_no_iam_user_policy(self):
        sb_pack = SbCfnPackage(os.path.dirname(os.path.abspath(__file__)) + "/../sample.yaml")
        sb_pack.service_spec['Bindings']['IAM'].pop('Policies')
        outp = sb_pack.build_artifacts('test')['template']['Resources']['AWSSBInjectedIAMUserCreator']['Properties']
        self.assertEqual('PolicyArns' not in outp.keys(), True)

    def test_create_apb_skeleton(self):
        sb_pack = SbCfnPackage(os.path.dirname(os.path.abspath(__file__)) + "/../sample.yaml")
        artifacts = sb_pack.build_artifacts('test', test=True)
        results = sb_pack.create_apb_skeleton(artifacts['apb_spec'], artifacts['prescribed_parameters'], artifacts['bindings'], artifacts['template'], 'SQS')
        self.assertEqual(results.startswith('/tmp/AWSSB-'), True)

if __name__ == '__main__':
    unittest.main()
