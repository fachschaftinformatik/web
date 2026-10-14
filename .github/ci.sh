#!/usr/bin/env bash
set -euo pipefail

# Read NUL-delimited paths so whitespace and special characters remain literal.
api=false
website=false
workflows=false
while IFS= read -r -d '' path; do
  case "$path" in
    .github/workflows/ci.yaml|.github/ci.sh)
      api=true
      website=true
      workflows=true
      ;;
  esac
  case "$path" in
    *.go|go.mod|go.sum|database/*|internal/*|Dockerfile)
      api=true
      ;;
  esac
  case "$path" in
    website/*|docs/*|.node-version)
      website=true
      ;;
  esac
  case "$path" in
    .dockerignore|compose.yaml|garage.toml)
      api=true
      website=true
      ;;
  esac
  case "$path" in
    .github/workflows/*)
      workflows=true
      ;;
  esac
done
printf 'api=%s\nwebsite=%s\nworkflows=%s\n' "$api" "$website" "$workflows"
