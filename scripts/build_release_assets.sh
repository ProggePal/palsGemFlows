#!/usr/bin/env sh
set -eu

# Builds the cross-platform release zips into dist/ using scripts/package.sh.
# Usage:
#   VERSION=v0.1.0 ./scripts/build_release_assets.sh

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

version="${VERSION:-}"
if [ -z "$version" ]; then
	echo "VERSION is required (e.g. VERSION=v0.1.0)" >&2
	exit 2
fi

echo "Building release assets for $version" >&2

echo "- darwin/arm64" >&2
GOOS=darwin GOARCH=arm64 VERSION="$version" ./scripts/package.sh

echo "- darwin/amd64" >&2
GOOS=darwin GOARCH=amd64 VERSION="$version" ./scripts/package.sh

echo "- linux/amd64" >&2
GOOS=linux GOARCH=amd64 VERSION="$version" ./scripts/package.sh

echo "- windows/amd64" >&2
GOOS=windows GOARCH=amd64 VERSION="$version" ./scripts/package.sh

echo "\nArtifacts:" >&2
ls -1 "dist" | grep "pals-gemflows_${version}_" || true
ls -1 "dist" | grep "pals-gemflows_v${version}_" || true
