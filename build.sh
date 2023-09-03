#!/bin/sh

export GO111MODULE=on

EXENAME=xdemo
PKG=github.com/askasoft/pango-xdemo/app
VERSION=1.0.0
REVISION=`git rev-parse --short HEAD`
BUILDTIME=`date -u "+%Y-%m-%dT%H:%M:%SZ"`

go build -ldflags "-X ${PKG}.Version=${VERSION} -X ${PKG}.Revision=${REVISION} -X ${PKG}.buildTime=${BUILDTIME}" -o ${EXENAME}
