#!/bin/sh -ex

export EXE=xdemodb

../build.sh
mv -f $EXE ../

export EXE=xdemo
