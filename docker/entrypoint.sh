#!/bin/sh
set -eu

if [ "${1:-serve}" != serve ]; then
    exec "$@"
fi

# The backend is always local to the combined application container.
export BACKEND_URL=http://127.0.0.1:8080
export HTTP_PORT=8080
exec s6-svscan /etc/s6
