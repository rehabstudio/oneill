#!/bin/bash
set -e

usage() {
    base="$(basename "$0")"
    cat <<EOUSAGE

usage: $base command

This script will build a docker container and compile your application, outputting a binary to the root of your application.

Available Commands:

  $base local
    (equivalent to "godep go build -a")

  $base docker
    (runs "$base local" inside a container)

EOUSAGE
}

local_build() {

    # build templates into go source code for embedding in our binary
    go-bindata -o proxy/bindata.go -pkg=proxy -prefix=proxy/ proxy/templates/
    if [ -n $2 ]; then
        chown $2:$2 proxy/bindata.go
    fi

    # run the application's tests before building
    godep go test -a -cover ./...

    # build the application and output a binary
    godep go build -a
    if [ -n $2 ]; then
        chown $2:$2 oneill
    fi
}

docker_build() {
    docker build -q -t "oneill-build" . >/dev/null
    docker run -ti --rm -v `pwd`:/go/src/github.com/rehabstudio/oneill "oneill-build" bash /go/src/github.com/rehabstudio/oneill/build.sh "local" "`id -u`"
}

case "$1" in
    local)
        local_build $@ >&2
        ;;

    docker)
        docker_build >&2
        ;;

    *)
        echo >&2 'error: unknown command:' "$1"
        usage >&2
        exit 1
        ;;
esac
