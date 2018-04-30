from setuptools import setup

setup_options = dict(
    name="sb_cfn_package",
    version="1.0",
    description="Build AWS Service Broker services out of CloudFormation templates",
    url="https://github.com/awslabs/aws-servicebroker",
    author="AWS Hybrid Team",
    license="Apache 2.0",
    packages=['sb_cfn_package'],
    zip_safe=False,
    extras_require={
        ':python_version>="3.5"': ['argparse>=1.1', 'taskcat>=2018.305.233813', 'pip', 'pyyaml']
    },

    classifiers=(
        'Intended Audience :: Developers',
        'Intended Audience :: System Administrators',
        'Natural Language :: English',
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3.5'
    ),
    scripts=['bin/sb_cfn_package'],
    include_package_data=True
)

setup(**setup_options)
