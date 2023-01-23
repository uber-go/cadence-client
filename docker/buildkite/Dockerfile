FROM golang:1.19.5

RUN mkdir -p /go/src/go.uber.org/cadence
WORKDIR /go/src/go.uber.org/cadence

ADD go.mod go.sum /go/src/go.uber.org/cadence/
RUN GO111MODULE=on go mod download
