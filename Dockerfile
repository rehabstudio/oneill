FROM golang:latest
MAINTAINER Patrick Carey "patrick@rehabstudio.com"

RUN go get github.com/tools/godep
WORKDIR /go/src/github.com/rehabstudio/oneill
