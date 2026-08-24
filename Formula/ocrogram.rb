# frozen_string_literal: true

# Draft formula. Copy into github.com/joelpeckham/homebrew-ocrogram
# and fill url/sha256 when tagging a release.

class Ocrogram < Formula
  desc "Extract text from macOS screenshots into the clipboard"
  homepage "https://github.com/joelpeckham/ocrogram"
  # url "https://github.com/joelpeckham/ocrogram/archive/refs/tags/v0.1.0.tar.gz"
  # sha256 "replace-me"
  license "MIT"
  head "https://github.com/joelpeckham/ocrogram.git", branch: "main"

  depends_on :macos
  depends_on "go" => :build
  depends_on xcode: :build

  def install
    system "make", "VERSION=#{version}"
    bin.install "bin/ocrogram"
    bin.install "bin/ocrogram-helper"
  end

  def caveats
    <<~EOS
      Run `ocrogram start` to enable the login item.
      `ocrogram stop` removes it.
    EOS
  end

  test do
    assert_match "ocrogram", shell_output("#{bin}/ocrogram --version")
    assert_match "usage: ocrogram-helper", shell_output("#{bin}/ocrogram-helper 2>&1", 2)
  end
end
