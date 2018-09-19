#!/bin/bash -xe

cd build-tools/apb_packaging
pip3 install --upgrade pip==9.0.3
python3 setup.py test
python3 setup.py install
cd ../../templates/

docker login -u ${DH_USER} -p ${DH_PASS}
for a in $(cat ../apbs) ; do
    cd ${a}
    sb_cfn_package -t ${DH_ORG}/${a}-apb -n ${a} -b ${DH_ORG} --ci '../../build/' ./template.yaml
    docker tag ${DH_ORG}/${a}-apb:latest ${DH_ORG}/${a}-apb:$(cat ${CODEBUILD_SRC_DIR}/commit_id)
    docker push ${DH_ORG}/${a}-apb:$(cat ${CODEBUILD_SRC_DIR}/commit_id)
    cd ../
done
