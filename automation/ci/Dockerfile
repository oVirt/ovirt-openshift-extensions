FROM openshift/origin-ansible@sha256:030adfc1b9bc8b1ad0722632ecf469018c20a4aeaed0672f9466e433003e666c

USER root
RUN yum install -y \
	http://resources.ovirt.org/pub/yum-repo/ovirt-release42.rpm	

ARG PKGS='file python-ovirt-engine-sdk4 ovirt-ansible-roles'

RUN yum install -y ${PKGS} && yum clean all && rm -rf /var/run/cache

RUN ssh-keygen -t rsa -N '' -f id_rsa

ADD integ.ini integ.ini
ADD vars.yaml vars.yaml
ADD install_okd.yaml  install_okd.yaml
ADD install_extensions.yaml  install_extensions.yaml
ADD setup_dns.yaml  setup_dns.yaml
ADD flex_deployer_job.yaml flex_deployer_job.yaml

ENV OPTS="-v "
ENV ANSIBLE_ROLES_PATH="/usr/share/ansible/roles:/usr/share/ansible/openshift-ansible/roles/"
ENV ANSIBLE_JINJA2_EXTENSIONS="jinja2.ext.do"
ENV INVENTORY_FILE="integ.ini"
ENV PLAYBOOK_FILE="install_okd.yaml"
