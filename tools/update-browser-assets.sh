#!/bin/sh
# Run with curl, tar and Dart Sass available; no Node build is required.
set -eu

version=6.6.0
root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
assets="$root/internal/handler/http/assets"
temp=$(mktemp -d)
trap 'rm -rf "$temp"' EXIT HUP INT TERM

curl --fail --location --silent --show-error \
    "https://registry.npmjs.org/govuk-frontend/-/govuk-frontend-$version.tgz" \
    --output "$temp/govuk.tgz"
tar -xzf "$temp/govuk.tgz" -C "$temp"
sass --no-source-map --style=compressed --quiet-deps \
    --load-path="$temp/package/dist" \
    "$root/internal/handler/http/styles/govuk.scss" "$assets/govuk.css"
cp "$temp/package/dist/govuk/govuk-frontend.min.js" "$assets/govuk.js"
mkdir -p "$assets/fonts"
install -m 0644 "$temp/package/dist/govuk/assets/fonts/"*.woff2 "$assets/fonts/"
curl --fail --location --silent --show-error \
    "https://raw.githubusercontent.com/alphagov/govuk-frontend/v$version/LICENSE.txt" \
    --output "$assets/GOVUK-LICENSE.txt"

curl --fail --location --silent --show-error \
    https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js \
    --output "$assets/htmx.js"
curl --fail --location --silent --show-error \
    https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/LICENSE \
    --output "$assets/HTMX-LICENSE.txt"
