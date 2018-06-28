From centos:7

USER root

RUN yum install -y \ 
	centos-release-openshift-origin \
	http://resources.ovirt.org/pub/yum-repo/ovirt-release42.rpm	
RUN yum install -y \
	file \
	ansible \
	atomic-openshift-utils \
	git \
	python-ovirt-engine-sdk4 \
	ovirt-ansible-roles \
	&& yum clean all

RUN rm -rf /var/run/cache

RUN ssh-keygen -t rsa -N '' -f ~/.ssh/id_rsa

ADD ansible.cfg /root/ansible.cfg
ADD integ.ini /root/integ.ini
ADD vars.yaml /root/vars.yaml
ADD site.yaml  /root/site.yaml
ADD setup_dns.yaml  /root/setup_dns.yaml
ADD deploy_flex.yaml  /root/deploy_flex.yaml
ADD flex_deployer_pod.yaml  /root/flex_deployer_pod.yaml

WORKDIR /root
ENTRYPOINT ansible-playbook -i integ.ini site.yaml
