#!/usr/bin/env bash

apk add --update libc-dev

export CC=gcc
export CGO_ENABLED=1

rm -rf pkg/linux-amd64
rm -rf bin

script/build docker
