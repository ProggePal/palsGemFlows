class PalsGemflows < Formula
  desc "Run pre-made AI workflows from YAML recipes"
  homepage "https://github.com/ProggePal/palsGemFlows"
  version "0.1.1"

  base = "https://github.com/ProggePal/palsGemFlows/releases/download/0.1.1/"

  on_macos do
    if Hardware::CPU.arm?
      url base + "pals-gemflows_v0.1.1_darwin_arm64.zip"
      sha256 "3daeabea0c601b6663f772a638e3288452101c43d8f05e7304ce04069dbc95e2"
    else
      url base + "pals-gemflows_v0.1.1_darwin_amd64.zip"
      sha256 "daf3e7bba3462cb375400758836f8dd188693ddea02a491417605b79d242a30b"
    end
  end

  on_linux do
    url base + "pals-gemflows_v0.1.1_linux_amd64.zip"
    sha256 "2199e3ff68ee31b0aa1eac97498e33f454daa543072129621f48ba1fbe73261d"
  end

  def install
    bin.install "pals-gemflows"
  end

  test do
    assert_match "Pals GemFlows", shell_output("#{bin}/pals-gemflows --version")
  end
end
