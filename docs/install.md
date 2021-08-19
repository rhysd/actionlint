Installation
============

This document describes how to install [actionlint](..).

## [Homebrew][homebrew] on macOS

Tap this repository and install `actionlint` package. This is a recommended way to manage `actionlint` command on macOS.

```sh
brew tap "rhysd/actionlint" "https://github.com/rhysd/actionlint"
brew install actionlint
```

## Prebuilt binaries

Download an archive file from [the releases page][releases] for your platform, unarchive it and put the executable file to a
directory in `$PATH`.

Prebuilt binaries are built at each release by CI for the following OS and arch:

- macOS (x86_64, arm64)
- Linux (i386, x86_64, arm32, arm64)
- Windows (i386, x86_64)
- FreeBSD (i386, x86_64)

Note: The author doesn't have Apple M1 environment so `darwin/arm64` target binary is not tested.

<a name="download-script"></a>
## Download script

To install `actionlint` executable with one command, [the download script](../scripts/download-actionlint.bash) is available.
It downloads the latest version of actionlint (`actionlint.exe` on Windows and `actionlint` on other OSes) to the current
directory automatically. This is a recommended way if you install actionlint in some shell script.

```sh
bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
```

For the usage of actionlint on GitHub Actions, see [the usage document](usage.md#on-github-actions).

## Build from source

Recent [Go][] toolchain is necessary to build actionlint from source. Use Go 1.16 or later.

```sh
# Install the latest stable version
go install github.com/rhysd/actionlint/cmd/actionlint@latest

# Install the head of main branch
go install github.com/rhysd/actionlint/cmd/actionlint
```

---

[Checks](checks.md) | [Usage](usage.md) | [Configuration](config.md) | [Go API](api.md) | [References](reference.md)

[homebrew]: https://brew.sh/
[releases]: https://github.com/rhysd/actionlint/releases
[Go]: https://golang.org/
