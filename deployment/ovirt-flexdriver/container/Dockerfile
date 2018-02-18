FROM fedora:27

MAINTAINER Roy Golan, rgolan@redhat.com

RUN dnf install -y golang git

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
COPY deployment/ovirt-flexdriver/deploy.yaml $PREFIX
COPY deployment/ovirt-flexdriver/ovirt-flexdriver.conf.j2 $PREFIX
COPY deployment/entrypoint.sh /entrypoint.sh

RUN make deps
RUN make build-flex -e

ENTRYPOINT /entrypoint.sh
