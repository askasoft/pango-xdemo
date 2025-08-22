#!/bin/bash -ex

export GO111MODULE=on

VERSION=1.2.0
if [ -z "$REVISION" ]; then
  REVISION=`git rev-parse --short HEAD`
fi
BUILDTIME=`date -u "+%Y-%m-%dT%H:%M:%SZ"`

PKG=github.com/askasoft/pangox/xwa
LDF="-X ${PKG}.Version=${VERSION} -X ${PKG}.Revision=${REVISION} -X ${PKG}.Buildtime=${BUILDTIME}"

go build -ldflags "${LDF}" -o xdemo
go build -ldflags "${LDF}" -o xdemodb ./cmd

go test ./...
