#!/bin/sh
set -eu

wget -q -T 3 -O /dev/null http://127.0.0.1:8080/api/v1/office-status
wget -q -T 3 -O /dev/null "http://127.0.0.1:${PORT:-80}/"
