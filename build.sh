#!/bin/bash -e

export GO111MODULE=on

if [ -z "$EXE" ]; then
  EXE=xdemo
fi
PKG=github.com/askasoft/pangox-xdemo/app
VERSION=1.0.0
if [ -z "$REVISION" ]; then
  REVISION=`git rev-parse --short HEAD`
fi
BUILDTIME=`date -u "+%Y-%m-%dT%H:%M:%SZ"`

go build -ldflags "-X ${PKG}.Version=${VERSION} -X ${PKG}.Revision=${REVISION} -X ${PKG}.buildTime=${BUILDTIME}" -o ${EXE}
go test ./...
