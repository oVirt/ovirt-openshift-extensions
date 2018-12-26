
INSTALL_OKD=${INSTALL_OKD:-1}
INSTALL_EXTENSIONS=${INSTALL_EXTENSIONS:-1}

if [ "$INSTALL_OKD" == "1" ]; then
  podman run \
    -it \
    --rm \
    -v $(pwd)/customization.yaml:/usr/share/ansible/openshift-ansible/customization.yaml:Z \
    -e OPTS="-e @customization.yaml" \
    quay.io/rgolangh/okd-on-ovirt-installer
fi


if [ "$INSTALL_EXTENSIONS" == "1" ]; then
  podman run \
    -it \
    --rm \
    -v $(pwd)/customization.yaml:/usr/share/ansible/openshift-ansible/customization.yaml:Z \
    -e OPTS="-e @customization.yaml" \
    -e PLAYBOOK_FILE="install_extensions.yaml" \
    quay.io/rgolangh/okd-on-ovirt-installer
fi
