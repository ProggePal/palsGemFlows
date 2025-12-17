class PalsGemflows < Formula
  desc "Run pre-made AI workflows from YAML recipes"
  homepage "https://github.com/ProggePal/palsGemFlows"
  version "0.1.0"

  base = "https://github.com/ProggePal/palsGemFlows/releases/download/0.1.0/"

  on_macos do
    if Hardware::CPU.arm?
      url base + "pals-gemflows_v0.1.0_darwin_arm64.zip"
      sha256 "e2bdebded2b2e4db33ea9515d45e5c5fdb04d0b85e1902031ce53d2249df4ddd"
    else
      url base + "pals-gemflows_v0.1.0_darwin_amd64.zip"
      sha256 "95da58e16ce4c953ce7a442d8c6e825101b5023dda566ec530d7f2eb0dc3d89b"
    end
  end

  on_linux do
    url base + "pals-gemflows_v0.1.0_linux_amd64.zip"
    sha256 "6c67d3f9b66b9e694fd982598d06a89bc864b5ecfbc144b2a873a52d1b52c10b"
  end

  def install
    bin.install "pals-gemflows"
  end

  test do
    assert_match "Pals GemFlows", shell_output("#{bin}/pals-gemflows --version")
  end
end
