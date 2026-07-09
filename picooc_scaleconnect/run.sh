#!/bin/sh

SLEEP=""

if [ -f "/data/options.json" ]; then
  SLEEP="$(sed -n 's/.*"sleep"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /data/options.json | head -n 1)"
fi

cd /config || exit 1

exec /app/scaleconnect -c /config/scaleconnect.yaml -i -r "${SLEEP:-24h}"
