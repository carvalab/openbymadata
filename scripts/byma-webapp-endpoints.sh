#!/usr/bin/env bash
# Lists what the open.bymadata.com.ar web app calls, so changes in BYMA's
# API show up before users hit them. Prints:
#   1. the "Options" header each endpoint gets (compare with
#      sectionHeader in internal/api/endpoints.go)
#   2. free-API endpoints this library does not call yet
# Usage: scripts/byma-webapp-endpoints.sh
set -euo pipefail

base="https://open.bymadata.com.ar"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

curl -fsSL -A "Mozilla/5.0" "$base/" -o "$work/index.html"
main_js="$(grep -oE 'src="main\.[a-f0-9]+\.js"' "$work/index.html" | sed 's/src="//; s/"//')"
curl -fsSL -A "Mozilla/5.0" "$base/$main_js" -o "$work/main.js"
echo "web app bundle: $main_js"

echo
echo "== Options header per endpoint (request interceptor) =="
grep -oE 'if\(\[[^]]*\]\.some\([a-zA-Z]+=>[a-zA-Z]+\.url\.includes\([a-zA-Z]+\)\)\)\{[a-zA-Z]+\.Options="[a-z-]+"' "$work/main.js" |
  sed -E 's/if\(\[(.*)\]\.some.*\.Options="([a-z-]+)"/\2: \1/' | tr -d '"'

echo
echo "== Endpoints the web app calls that this library does not =="
repo="$(cd "$(dirname "$0")/.." && pwd)"
known="$(grep -oE '= "[^"]+"' "$repo/internal/api/endpoints.go" | sed 's/= "//; s/"$//')"
{
  # API paths referenced directly, plus every path in the interceptor lists.
  grep -oE '"/?bymadata/free/[^"?]+' "$work/main.js" | sed -E 's#^"/?bymadata/free/##'
  grep -oE 'if\(\[[^]]*\]\.some\([a-zA-Z]+=>[a-zA-Z]+\.url\.includes' "$work/main.js" |
    grep -oE '"/[^"]+"' | tr -d '"' | sed -E 's#^/##; s#\?.*$##'
} | sort -u |
  while read -r ep; do
    grep -qxF "$ep" <<<"$known" || grep -qF "$ep" <<<"$known" || echo "$ep"
  done
