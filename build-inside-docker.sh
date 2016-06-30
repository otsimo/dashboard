#!/usr/bin/env bash

BUILD_IMAGE="sercand/go-docker"

gcrenv=$1

if [ "$gcrenv" = "" ];then
    gcrenv="prod"
fi
docker run -t --rm=true \
    --entrypoint=/bin/bash \
    -v $PWD:/opt/otsimo/mono:rw \
    -w /opt/otsimo/mono \
    -e "BUILD_ENV=$gcrenv" \
    ${BUILD_IMAGE} \
    -i script/alpine.sh

script/docker $2 gcr ${gcrenv}
