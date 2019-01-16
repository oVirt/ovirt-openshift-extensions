#!/bin/sh -ex

function cleanup() {
  podman umount $cid
  podman stop $cid
  podman rm $cid
}

trap cleanup 0

cid=$(podman run -d --entrypoint sleep openshift/origin-ansible:v3.11 1d)

openshift_ansible_dir=$(podman mount --notruncate ${cid})

#files="integ.ini vars.yaml install_okd.yaml install_extensions.yaml setup_dns.yaml deploy_ovirt_storage_driver.yaml flex_deployer_job.yaml"
files="integ.ini vars.yaml install_okd.yaml setup_dns.yaml"
tree $openshift_ansible_dir 

cp -v $files $openshift_ansible_dir/usr/share/ansible/openshift-ansible
tree  $openshift_ansible_dir
cd $openshift_ansible_dir/usr/share/ansible/openshift-ansible

export ANSIBLE_JINJA2_EXTENSIONS="jinja2.ext.do"
#export ANSIBLE_ROLES_PATH="/usr/share/ansible/roles/:$openshift_ansible_dir/usr/share/ansible/openshift-ansible"

ansible-playbook -i integ.ini install_okd.yaml -e @vars.yaml
#ansible-playbook -i integ.ini install_extensions.yaml -e @vars.yaml


