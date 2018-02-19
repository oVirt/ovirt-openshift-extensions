#!/bin/sh -ex

function usage() {
cat << EOF

USAGE
    Invoke ansible-playbook with default deploy.yaml file.

    This entrypoint implements deployment container
    It will invoke an Ansible playbook and copy the pre-built binary and distribute it to
    the master and nodes of the k8s/openshift cluster.

    entrypoint.sh --help
    entrypoint.sh [inventory_file]
    entrypoint.sh [valid Ansible variables] # like -vvv or any other input starting with

PARAMETERS
    inventory_file - pass in an inventory file or override using a volume the /etc/ansible/hosts one
                     example of the expected inventory groups:

                     [k8s-masters]
                     kube-master.example.com
                     
                     [k8s-nodes]
                     kube-node1.example.com
                     kube-node2.example.com
                     
                     [all:vars]
                     engine_url=https://engine.example.com:8443/ovirt-engine/api
                     engine_username=admin@internal
                     engine_password=123
                     engine_insecure=false
                     engine_ca_file=

EXMAPLE
    run the container with the ssh directory shared, so ansible will connect to all of the hosts, plus pass an inventory file

    sudo docker run --rm -it \
        -v /root/.ssh:/root/.ssh:z \
        -v /etc/ansible/hosts:/etc/ansible/hosts \
        rgolangh/ovirt-flexdriver-ansible:v0.2.0
EOF
}

if [[ "$#" == "1" ]]; then
    if [[ "$1" == "--help" ]]; then
        usage
        exit 1
    fi
    if [[ ! "$1" =~ "-" ]]; then
        exec ansible-playbook -i $1 /opt/deploy.yaml
    	exit
    fi
fi

exec ansible-playbook "$@" /opt/deploy.yaml
