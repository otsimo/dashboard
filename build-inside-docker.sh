#!/usr/bin/env bash

BUILD_IMAGE="sercand/go-docker:1.6.2-v4"

docker run -t --rm=true \
    --entrypoint=/bin/bash \
    -v $PWD:/opt/otsimo/mono:rw \
    -w /opt/otsimo/mono \
    ${BUILD_IMAGE} \
    -i script/alpine.sh

gcrenv=$1

if [ "$gcrenv" = "" ];then
    gcrenv="prod"
fi

script/docker $2 gcr ${gcrenv}
