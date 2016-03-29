FROM golang:1.4.3

RUN go get github.com/tools/godep
WORKDIR /go/src/github.com/chenchun/cgroupfs
COPY . /go/src/github.com/chenchun/cgroupfs
ENV GOPATH /go/src/github.com/chenchun/cgroupfs/Godeps/_workspace:/go
RUN go build -o /tmp/cgroupfs cli/cli.go

