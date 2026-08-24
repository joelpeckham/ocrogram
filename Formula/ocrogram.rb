# frozen_string_literal: true

# Stub Homebrew formula. Fill in url/sha256 once there is a tagged release.
class Ocrogram < Formula
  desc "Extract text from macOS screenshots into the clipboard"
  homepage "https://github.com/joelpeckham/ocrogram"
  # url "https://github.com/joelpeckham/ocrogram/archive/refs/tags/v0.1.0.tar.gz"
  # sha256 "replace-me"
  license "MIT"
  head "https://github.com/joelpeckham/ocrogram.git", branch: "main"

  depends_on :macos
  depends_on "go" => :build
  uses_from_macos "swift" => :build

  def install
    system "make"
    bin.install "bin/ocrogram"
    bin.install "bin/ocrogram-helper"
  end

  service do
    run [opt_bin/"ocrogram", "daemon"]
    keep_alive true
    require_root false
  end

  test do
    assert_match "not implemented yet", shell_output("#{bin}/ocrogram daemon")
    assert_match "not implemented yet", shell_output("#{bin}/ocrogram-helper")
  end
end
