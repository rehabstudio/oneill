FROM golang:latest
MAINTAINER Patrick Carey "patrick@rehabstudio.com"

# install godep and coverage tools
RUN go get github.com/jteeuwen/go-bindata/go-bindata
RUN go get github.com/tools/godep
RUN go get golang.org/x/tools/cmd/cover

WORKDIR /go/src/github.com/rehabstudio/oneill
