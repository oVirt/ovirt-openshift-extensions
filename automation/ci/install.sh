
INSTALL_OKD=${INSTALL_OKD:-1}
INSTALL_EXTENSIONS=${INSTALL_EXTENSIONS:-1}

if [ "$INSTALL_OKD" == "1" ]; then
  podman run \
    -it \
    --rm \
    -v $(pwd)/vars.yaml:/usr/share/ansible/openshift-ansible/vars.yaml:Z \
    -e OPTS="-e @vars.yaml" \
    -e ANSIBLE_SSH_ARGS="-o ControlMaster=no" \
    quay.io/rgolangh/ovirt-openshift-installer
fi


if [ "$INSTALL_EXTENSIONS" == "1" ]; then
  podman run \
    -it \
    --rm \
    -v $(pwd)/vars.yaml:/usr/share/ansible/openshift-ansible/vars.yaml:Z \
    -e OPTS="-e @vars.yaml" \
    -e PLAYBOOK_FILE="install_extensions.yaml" \
    -e ANSIBLE_SSH_ARGS="-o ControlMaster=no" \
    quay.io/rgolangh/ovirt-openshift-installer
fi
