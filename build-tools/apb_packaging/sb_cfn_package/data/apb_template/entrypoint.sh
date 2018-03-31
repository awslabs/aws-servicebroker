#!/bin/bash
ACTION=$1
USER_ID=$(id -u)
shift
playbooks=/opt/apb/actions

if [ ${USER_UID} != ${USER_ID} ]; then
  sed "s@${USER_NAME}:x:\${USER_ID}:@${USER_NAME}:x:${USER_ID}:@g" ${BASE_DIR}/etc/passwd.template > /etc/passwd
fi
oc-login.sh

if [[ -e "$playbooks/$ACTION.yaml" ]]; then
  ansible-playbook $playbooks/$ACTION.yaml $@
elif [[ -e "$playbooks/$ACTION.yml" ]]; then
  ansible-playbook $playbooks/$ACTION.yml $@
else
  echo "'$ACTION' NOT IMPLEMENTED" # TODO
fi