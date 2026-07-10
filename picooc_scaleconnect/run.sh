#!/bin/sh

SLEEP=""
DB=""
DB_PATH=""

if [ -f "/data/options.json" ]; then
  SLEEP="$(sed -n 's/.*"sleep"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /data/options.json | head -n 1)"
  DB="$(sed -n 's/.*"db"[[:space:]]*:[[:space:]]*\(true\|false\).*/\1/p' /data/options.json | head -n 1)"
  DB_PATH="$(sed -n 's/.*"db_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /data/options.json | head -n 1)"
fi

cd /data || cd /config || exit 1

if [ "$DB" = "true" ]; then
  if [ -n "$DB_PATH" ]; then
    exec /app/scaleconnect -c /config/scaleconnect.yaml -i -r "${SLEEP:-24h}" -db -db-path "$DB_PATH"
  fi
  exec /app/scaleconnect -c /config/scaleconnect.yaml -i -r "${SLEEP:-24h}" -db
fi

exec /app/scaleconnect -c /config/scaleconnect.yaml -i -r "${SLEEP:-24h}"
