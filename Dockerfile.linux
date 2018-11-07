FROM golang:1.11-alpine

RUN apk update
RUN apk add --no-cache ca-certificates openssl git make bash gcc musl-dev

RUN go get github.com/benbjohnson/ego/cmd/ego

ENV GO111MODULE=on
ENV PATH="${PATH}:/usr/local/go/bin"

WORKDIR /
