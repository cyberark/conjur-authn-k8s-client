#!/usr/bin/env bash

cd $(dirname "$0")
docker build -t sidecar-go:builder .
docker run --rm -it -v $PWD:/go/src/github.com/cyberark/sidecar-authenticator sidecar-go:builder bash
