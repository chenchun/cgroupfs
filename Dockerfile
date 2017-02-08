FROM golang:1.6.2

RUN apt-get update && apt-get install -y fuse
WORKDIR /go/src/github.com/chenchun/cgroupfs
COPY . /go/src/github.com/chenchun/cgroupfs

RUN go get bazil.org/fuse &&\
    go get github.com/Sirupsen/logrus &&\
    go get github.com/opencontainers/runc/libcontainer/cgroups &&\
    go get golang.org/x/net/context

RUN go build -o /tmp/cgroupfs cli/cli.go

