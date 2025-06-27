#!/bin/sh
# wait-for-postgres.sh

set -e

host="$1"
port="$2"
shift 2
cmd="$@"

until nc -z "$host" "$port"; do
  echo "Waiting for Postgres at $host:$port..."
  sleep 1
done

exec $cmd
