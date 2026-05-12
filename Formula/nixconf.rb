class Nixconf < Formula
  desc "Repository manager for NixOS multi-repo configuration"
  homepage "https://github.com/xx4h/nixconf"
  url "https://github.com/xx4h/nixconf.git",
      tag:      "v0.1.0",
      revision: "0000000000000000000000000000000000000000"
  license "Apache-2.0"
  head "https://github.com/xx4h/nixconf.git", branch: "main"

  livecheck do
    url :stable
    regex(/^v?(\d+(?:\.\d+)+)$/i)
  end

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/xx4h/nixconf/cmd.version=v#{version}
      -X github.com/xx4h/nixconf/cmd.commit=#{Utils.git_head}
      -X github.com/xx4h/nixconf/cmd.date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:)

    generate_completions_from_executable(bin/"nixconf", "completion")
  end

  test do
    assert_match "nixconf", shell_output("#{bin}/nixconf --help")
  end
end
