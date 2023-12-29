# Dev image, supporting dynamic linking with libgit2.so for faster rebuilds.
FROM golang:1.20-alpine3.18

# update this if different user and directory names are needed
ARG SERVICE=go

# should match host uid and gid
ARG UID=1000
ARG GID=1000

# TODO: newer alpine versions have libgit2 1.7, which is not yet supported by git2go
# https://github.com/libgit2/git2go#which-go-version-to-use
ARG LIBGIT2_VERSION=1.5

RUN apk add --no-cache gcc "libgit2-dev~=${LIBGIT2_VERSION}" musl-dev pkgconf tini

RUN apk add --no-cache --virtual .temp shadow \
    && groupadd --gid "$GID" "$SERVICE" \
    && useradd --uid "$UID" --gid "$GID" -m "$SERVICE" \
    && apk del .temp

# install deps and source code as non-root user, so that local development
# can also be done as non-root, while avoiding permissions issues on
# bind-mounted volumes (this works because $GOPATH is world-writable)
USER $SERVICE
WORKDIR /opt/$SERVICE
COPY --chown=$SERVICE go.mod go.sum ./
RUN go mod download && go mod verify

COPY --chown=$SERVICE . .

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/bin/sh"]
