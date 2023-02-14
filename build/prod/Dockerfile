FROM golang:1.19-alpine3.17 AS builder

ARG SERVICE=go
ARG LIBGIT2_VERSION=1.5.1

RUN apk add --no-cache "libgit2-dev~=${LIBGIT2_VERSION} pkgconf"

WORKDIR /opt/$SERVICE
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN mkdir "/opt/$SERVICE/bin" \
    && go build -o "/opt/$SERVICE/bin" -v -tags static,system_libgit2 ./...

FROM alpine:latest

ARG SERVICE=go
ARG UID=1000
ARG GID=1000

RUN apk add --no-cache tini

RUN apk add --no-cache --virtual .temp shadow \
    && groupadd --gid "$GID" "$SERVICE" \
    && useradd --uid "$UID" --gid "$GID" --system --shell "/sbin/nologin" "$SERVICE" \
    && apk del .temp

WORKDIR /opt/$SERVICE/bin
COPY --from=builder /opt/$SERVICE/bin .

USER $SERVICE
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/bin/sh"]
