FROM openshift/origin-ansible:v3.11

USER root
RUN yum install -y \
	http://resources.ovirt.org/pub/yum-repo/ovirt-release42.rpm	

# Installing pre release of 4.2 till GA of 4.2.7 see github issues/73
RUN yum install -y ovirt-release42-pre

# patch openshift-ansible
RUN ls -la /usr/share/ansible && cd /usr/share/ansible/openshift-ansible/
# support mac_addresses PR 11442
# fix multi-node issue PR 11459
RUN curl -L -o - https://github.com/openshift/openshift-ansible/pull/11459/commits/8cab65123f07aa4d93e3763e9949611a041bf185.patch | patch -f -p1
RUN curl -L -o - https://github.com/openshift/openshift-ansible/pull/11442/commits/edfe85d79367f869671865927afc11235f675184.patch | patch -f -p1 || true

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
ENV ANSIBLE_ROLES_PATH="/usr/share/ansible/openshift-ansible/roles/"
ENV ANSIBLE_JINJA2_EXTENSIONS="jinja2.ext.do"
ENV INVENTORY_FILE="integ.ini"
ENV PLAYBOOK_FILE="install_okd.yaml"
