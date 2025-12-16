#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

os="$(go env GOOS)"
arch="$(go env GOARCH)"
name="pals-gemflows_${os}_${arch}"

dist_dir="dist/$name"
rm -rf "$dist_dir"
mkdir -p "$dist_dir"

# Build binary
GOOS="$os" GOARCH="$arch" go build -o "$dist_dir/pals-gemflows" ./cmd/pals-gemflows

# One-click launcher for macOS
cp "scripts/Run Pals-GemFlows.command" "$dist_dir/Run Pals-GemFlows.command"
chmod +x "$dist_dir/pals-gemflows" "$dist_dir/Run Pals-GemFlows.command"

# Bundle README + example workflows
cp README.md "$dist_dir/README.md"
cp .env.example "$dist_dir/.env.example"
mkdir -p "$dist_dir/workflows" && cp -R workflows/. "$dist_dir/workflows/"

# Zip it
(
  cd dist
  rm -f "$name.zip"
  zip -r "$name.zip" "$name" >/dev/null
)

echo "Created dist/${name}.zip"
