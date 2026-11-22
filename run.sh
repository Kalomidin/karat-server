#!/bin/sh
set -eu

cd "$(dirname "$0")"
export PATH="/opt/homebrew/bin:/usr/local/bin:${PATH}"

if [ ! -f .env ]; then
  echo "missing .env — copy .env.example and fill it in" >&2
  exit 1
fi

docker compose up -d --wait db
exec go run ./cmd/karat-server
