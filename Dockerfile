FROM golang:1.12
MAINTAINER Conjur Inc

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /opt/conjur-authn-k8s-client
EXPOSE 8080

RUN apt-get update && apt-get install -y jq

RUN go get -u github.com/jstemmer/go-junit-report && \
    go get github.com/smartystreets/goconvey
