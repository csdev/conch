#!/bin/bash
set -eo pipefail

docker-compose build app
cid="$(docker create csang/conch:latest)"
trap "docker rm $cid" EXIT
docker cp "${cid}:/opt/go/bin/conch" .
