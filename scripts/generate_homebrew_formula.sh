#!/usr/bin/env sh
set -eu

# Generates a Homebrew formula for a tap repo.
# It expects versioned release zips to already exist in dist/.
#
# Usage:
#   VERSION=v0.1.0 ./scripts/build_release_assets.sh
#   VERSION=v0.1.0 ./scripts/generate_homebrew_formula.sh > pals-gemflows.rb
#
# Then copy pals-gemflows.rb into your tap repo under Formula/pals-gemflows.rb.

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

version="${VERSION:-}"
if [ -z "$version" ]; then
	echo "VERSION is required (e.g. VERSION=v0.1.0)" >&2
	exit 2
fi

repo="ProggePal/palsGemFlows"

tag="$version"

# Normalize to 'vX.Y.Z' for filenames used by package.sh
case "$version" in
	v*) filever="$version" ;;
	*)  filever="v$version" ;;
esac

sha() {
	# macOS has shasum by default
	shasum -a 256 "$1" | awk '{print $1}'
}

asset_darwin_arm64="dist/pals-gemflows_${filever}_darwin_arm64.zip"
asset_darwin_amd64="dist/pals-gemflows_${filever}_darwin_amd64.zip"
asset_linux_amd64="dist/pals-gemflows_${filever}_linux_amd64.zip"
asset_windows_amd64="dist/pals-gemflows_${filever}_windows_amd64.zip"

for f in "$asset_darwin_arm64" "$asset_darwin_amd64" "$asset_linux_amd64" "$asset_windows_amd64"; do
	if [ ! -f "$f" ]; then
		echo "Missing artifact: $f" >&2
		echo "Run: VERSION=$version ./scripts/build_release_assets.sh" >&2
		exit 2
	fi
done

sha_darwin_arm64="$(sha "$asset_darwin_arm64")"
sha_darwin_amd64="$(sha "$asset_darwin_amd64")"
sha_linux_amd64="$(sha "$asset_linux_amd64")"

# Homebrew won't install Windows binaries; we still build it for GitHub Releases.

cat <<EOF
class PalsGemflows < Formula
  desc "Run pre-made AI workflows from YAML recipes"
  homepage "https://github.com/$repo"
  version "${filever#v}"

  base = "https://github.com/$repo/releases/download/$tag/"

  on_macos do
    if Hardware::CPU.arm?
      url base + "pals-gemflows_${filever}_darwin_arm64.zip"
      sha256 "$sha_darwin_arm64"
    else
      url base + "pals-gemflows_${filever}_darwin_amd64.zip"
      sha256 "$sha_darwin_amd64"
    end
  end

  on_linux do
    url base + "pals-gemflows_${filever}_linux_amd64.zip"
    sha256 "$sha_linux_amd64"
  end

  def install
    bin.install "pals-gemflows"
  end

  test do
    assert_match "Pals GemFlows", shell_output("#{bin}/pals-gemflows --version")
  end
end
EOF
