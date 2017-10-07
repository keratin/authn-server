FROM golang:1.9-alpine
ARG glide_version="v0.13.0"

RUN apk update
RUN apk add --no-cache ca-certificates openssl git make bash gcc musl-dev

RUN wget https://github.com/Masterminds/glide/releases/download/${glide_version}/glide-${glide_version}-linux-amd64.tar.gz
RUN tar -xvzf glide-${glide_version}-linux-amd64.tar.gz
RUN mv linux-amd64/glide /usr/local/bin/glide

ENV PATH="${PATH}:/usr/local/go/bin"

WORKDIR /
