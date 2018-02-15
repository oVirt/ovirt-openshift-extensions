FROM fedora:27

MAINTAINER Roy Golan, rgolan@redhat.com

RUN dnf install -y golang git

RUN mkdir -p /go/{src,bin,pkg}
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN curl https://glide.sh/get | sh
RUN go get -u github.com/golang/dep/cmd/dep

#
RUN ls
WORKDIR /go/src/github.com/rgolangh/ovirt-flexdriver
COPY cmd cmd/
COPY internal internal/
COPY Makefile .
COPY Gopkg.toml .
COPY Gopkg.lock .
COPY glide.lock .
COPY glide.yaml .

RUN make deps
RUN make build-flex
