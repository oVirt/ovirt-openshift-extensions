FROM centos:7

LABEL maintainer="Roy Golan, rgolan@redhat.com"

ARG RPM=ovirt-flexvolume-driver*.rpm

COPY ${RPM} /root
COPY deployment/entrypoint.sh /entrypoint.sh

RUN yum install -y /root/${RPM} && yum clean all

ENTRYPOINT ["/entrypoint.sh"]
