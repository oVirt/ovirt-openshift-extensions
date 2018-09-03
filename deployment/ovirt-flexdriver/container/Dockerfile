FROM centos:7

LABEL maintainer="Roy Golan, rgolan@redhat.com"

ARG RPM
ARG RPM_DIR

COPY ${RPM_DIR}/${RPM} /root/${RPM}
COPY deployment/ovirt-flexdriver/entrypoint.sh /entrypoint.sh

RUN yum install -y /root/${RPM} && yum clean all

ENTRYPOINT ["/entrypoint.sh"]
