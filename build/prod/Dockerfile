# Compile statically-linked library for http-parser (libhttp_parser.a),
# which is currently not available via apk. Remove this once packaging
# is fixed upstream.
FROM alpine:3.17 AS patch

WORKDIR /opt/build
RUN apk add --no-cache gcc git make musl-dev
RUN git clone --depth=1 --branch=v2.9.4 'https://github.com/nodejs/http-parser.git'

WORKDIR /opt/build/http-parser
RUN make package

# Main application build.
FROM golang:1.19-alpine3.17 AS builder

ARG SERVICE=go
ARG LIBGIT2_VERSION=1.5

RUN apk add --no-cache gcc git musl-dev pkgconf

RUN apk add --no-cache \
    "libgit2-dev~=${LIBGIT2_VERSION}" "libgit2-static~=${LIBGIT2_VERSION}" \
    libssh2-dev libssh2-static \
    openssl-dev openssl-libs-static \
    zlib-dev zlib-static

COPY --from=patch /opt/build/http-parser/libhttp_parser.a /usr/local/lib/

WORKDIR /opt/$SERVICE
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN mkdir "/opt/$SERVICE/bin" \
    && go build -v -buildvcs=true \
        -tags "static,system_libgit2" -ldflags '-extldflags "-static"' \
        -o "/opt/$SERVICE/bin" ./...

# Production image with statically-linked binary and init environment.
FROM alpine:3.17

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

# The suggested working directory for production use.
# (Though other directories should work fine.)
WORKDIR /mnt/workspace

ENTRYPOINT ["/sbin/tini", "--", "/opt/go/bin/conch"]
