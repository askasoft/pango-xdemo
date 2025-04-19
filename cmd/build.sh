#!/bin/bash -e

BASEDIR=$(dirname $0)

pushd $BASEDIR

export EXE=xdemodb

../build.sh
mv -f $EXE ../

export -n EXE
