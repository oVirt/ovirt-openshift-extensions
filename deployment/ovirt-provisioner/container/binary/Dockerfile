# Copyright 2018 oVirt-maintainers
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM fedora:27

MAINTAINER Roy Golan, rgolan@redhat.com

RUN dnf install -y golang git && dnf clean all

RUN mkdir -p /go/{src,bin,pkg}
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ARG PREFIX=/opt
RUN curl https://glide.sh/get | sh
RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/rgolangh/ovirt-flexdriver
COPY cmd cmd/
COPY internal internal/
COPY Makefile .
COPY glide.lock .
COPY glide.yaml .

RUN make deps build-provisioner PREFIX=$PREFIX

ENTRYPOINT ["/opt/ovirt-provisioner"]
