#!/bin/sh -ex

export EXE=xdemoc

../build.sh
mv -f $EXE ../

export EXE=xdemo
