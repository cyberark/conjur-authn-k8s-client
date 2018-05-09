FROM golang:1.8
MAINTAINER Conjur Inc

RUN apt-get update && apt-get install jq
RUN go get -u github.com/jstemmer/go-junit-report
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get github.com/smartystreets/goconvey

RUN mkdir -p /go/src/github.com/cyberark/sidecar-authenticator
WORKDIR /go/src/github.com/cyberark/sidecar-authenticator

ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080
