# typed: false
# frozen_string_literal: true

class Actionlint < Formula
  desc "Static checker for GitHub Actions workflow files"
  homepage "https://github.com/rhysd/actionlint#readme"
  version "1.7.9"
  license "MIT"

  odie <<~EOS
    This formula is no longer maintained after v1.7.7 in favor of `actionlint` cask package.

    This formula was replaced with a cask with the same name at v1.7.9 release.
    Please migrate to the cask:

      $ brew uninstall actionlint
      $ brew install --cask actionlint

  EOS
end
