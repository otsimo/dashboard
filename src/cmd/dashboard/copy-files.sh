#!/bin/bash

#copy-files $1 $2
#copy-files src/cmd/dashboard bin/dashboard

SRC_PATH="$1"
DST_PATH="$2"

cp "$SRC_PATH/sample.yaml" $DST_PATH
