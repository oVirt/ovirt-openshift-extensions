From centos:7.5.1804

USER root

RUN yum install -y \ 
	http://resources.ovirt.org/pub/yum-repo/ovirt-release42.rpm	
RUN yum install -y \
	file \
	ansible \
	git \
	python-ovirt-engine-sdk4 \
	ovirt-ansible-roles \
	&& yum clean all

RUN rm -rf /var/run/cache

RUN ssh-keygen -t rsa -N '' -f ~/.ssh/id_rsa

ADD ansible.cfg /root/ansible.cfg
ADD integ.ini /root/integ.ini
ADD vars.yaml /root/vars.yaml
RUN touch /root/user_vars.yaml
ADD site.yaml  /root/site.yaml
ADD setup_dns.yaml  /root/setup_dns.yaml
ADD deploy_flex.yaml  /root/deploy_flex.yaml
ADD flex_deployer_job.yaml  /root/flex_deployer_job.yaml

WORKDIR /root
RUN git clone --branch release-3.10 https://github.com/openshift/openshift-ansible --depth 1
ENTRYPOINT ansible-playbook -i integ.ini site.yaml
