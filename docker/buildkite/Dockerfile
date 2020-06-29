FROM golang:1.12.1

RUN go get -u github.com/golang/dep/cmd/dep \
  && go get -u github.com/Masterminds/glide \
  && go get -u golang.org/x/lint/golint

RUN mkdir -p /go/src/github.com/sail1024/cadence
WORKDIR /go/src/github.com/sail1024/cadence
