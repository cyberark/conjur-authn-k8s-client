#!/usr/bin/env bash

cd $(dirname "$0")
TAG="${1:-sidecar:dev}"

echo "---"
echo "building sidecar-go with tag ${TAG} <<"

docker build -t sidecar-go:builder .
docker run --rm -v $PWD:/go/src/github.com/cyberark/sidecar-authenticator sidecar-go:builder env CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o authenticator .
docker build -f Dockerfile.scratch -t "${TAG}" .

echo "---"
