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

## CI services

Please try [the download script](../scripts/download-actionlint.bash). It downloads the latest version of actionlint
(`actionlint.exe` on Windows and `actionlint` on other OSes) to the current directory automatically. On GitHub Actions
environment, it sets a file path of downloaded executable to `executable` output in order to use the executable in the
following steps easily.

Here is an example of simple workflow to run actionlint on GitHub Actions. Please ensure `shell: bash` since the default
shell for Windows runners is `pwsh`.

```yaml
name: Lint GitHub Actions workflows
on: [push, pull_request]

jobs:
  actionlint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Download actionlint
        id: get_actionlint
        run: bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
        shell: bash
      - name: Check workflow files
        run: ${{ steps.get_actionlint.outputs.executable }} -color
        shell: bash
```

or simply run

```yaml
- name: Check workflow files
  run: |
    bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
    ./actionlint -color
  shell: bash
```

If you want to enable [shellcheck integration](checks.md#check-shellcheck-integ), install `shellcheck` command as follows:

```yaml
- name: Install shellcheck to enable shellcheck integration
  run: sudo apt install shellcheck
```

## Build from source

[Go][] toolchain is necessary.

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
