#!/usr/bin/env bash
# wait-for-it.sh

host="$1"
shift

until nc -z "$host"; do
  echo "⏳ Waiting for $host to be available..."
  sleep 1
done

exec "$@"
