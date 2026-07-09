#!/bin/sh

SLEEP=""

if [ -f "/data/options.json" ]; then
  SLEEP="$(sed -n 's/.*"sleep"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /data/options.json | head -n 1)"
fi

exec /app/scaleconnect -i -r "${SLEEP:-24h}"
