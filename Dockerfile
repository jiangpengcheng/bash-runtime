FROM golang:alpine3.15 AS builder

WORKDIR /root/pre-build

RUN apk add git gcc musl-dev
## download dependancy
### go
ADD go.mod go.sum ./
RUN go mod download

# build
WORKDIR /root/bash-runtime
ADD . /root/bash-runtime

# Make sure the bin directory exists
RUN mkdir -p bin
RUN go build -o bin/bash-runtime

FROM alpine:3.15

WORKDIR /root/bash-runtime
COPY --from=builder /root/bash-runtime/bin/bash-runtime .

RUN apk add bash && mkdir -p scripts
ADD scripts/exec.sh scripts/exec.sh

ENTRYPOINT [ "/root/bash-runtime/bash-runtime" ]
