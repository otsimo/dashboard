#!/usr/bin/env bash

apk add --update libc-dev

export CC=gcc
export CGO_ENABLED=1

if [ "${BUILD_ENV}" = "prod" ];then
    rm -rf pkg
fi
rm -rf bin

script/build docker package
