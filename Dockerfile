FROM golang:1.13.4 as builder

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

RUN go build ./cmd/client
RUN go build ./cmd/server

# Ideally we could use the "static" flavour but let's first start with the base flavour (which has glibc).
FROM gcr.io/distroless/base@sha256:4f25af540d54d0f43cd6bc1114b7709f35338ae97d29db2f9a06012e3e82aba8 as base

MAINTAINER Marko Mikulicic <mkm@bitnami.com>

ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99
ENV GRPC_GO_LOG_SEVERITY_LEVEL=info

#
FROM base as client
COPY --from=builder /src/client /usr/local/bin/

ENTRYPOINT ["client"]

#
FROM base as server

EXPOSE 8080

COPY --from=builder /src/server /usr/local/bin/

ENTRYPOINT ["server"]
