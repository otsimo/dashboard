#!/bin/bash

#copy-files $1 $2
#copy-files src/cmd/welcome bin/welcome

SRC_PATH="$1"
DST_PATH="$2"

cp -r "$SRC_PATH/templates" $DST_PATH/templates
