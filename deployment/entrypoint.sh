#!/bin/sh -e

function usage() {
cat << EOF
    USAGE
      This entrypoint implements deployment container
      It will invoke an Ansible playbook and copy the pre-built binary and distrubute to
      the master and nodes of the cluster

      entrypoint.sh --help
      entrypoint.sh [inventory_file]
      entrypoint.sh [valid Ansible variables] # like -vvv or any other input starting with '-'

    PARAMETERS
    inventory_file - pass in an inventory file or override using a volume the /etc/ansible/hosts one
EOF
}

if [[ "$#" == "1" ]]; then
    if [[ "$1" == "--help" ]]; then
        usage
        exit 1
    fi
    if [[ ! "$1" =~ "-" ]]; then
        ansible-playbook -i $1 /opt/deploy.yaml
    fi
else
    if [[ ! "$1" =~ "-" ]]; then
        usage
        exit 1
    fi
    ansible-playbook "$@" /opt/deploy.yaml
fi
