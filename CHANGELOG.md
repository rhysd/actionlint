<a name="v1.3.1"></a>
# [v1.3.1](https://github.com/rhysd/actionlint/releases/tag/v1.3.1) - 30 Jun 2021

- Files are checked in parallel. This made actionlint around 1.3x faster with 3 workflow files in my environment
- Manual for `man` command was added. `actionlint.1` is included in released archives
- `-version` now reports how the binary was built (Go version, arch, os, ...)
- Added `Command` struct to manage entire command lifecycle
- Order of checked files is now stable
- Added fuzz target for rule checkers

[Changes][v1.3.1]


<a name="v1.3.0"></a>
# [v1.3.0](https://github.com/rhysd/actionlint/releases/tag/v1.3.0) - 26 Jun 2021

- `-version` now outputs how the executable was installed.
- Fix errors output to stdout was not colorful on Windows.
- Add new `-color` flag to force to enable colorful outputs. This is useful when running actionlint on GitHub Actions since scripts at `run:` don't enable colors.
- `Linter.LintFiles` and `Linter.LintFile` methods take `project` parameter to explicitly specify what project the files belong to. Leaving it `nil` automatically detects projects from their file paths.
- `LintOptions.NoColor` is replaced by `LintOptions.Color`.

Example of `-version` output:

```console
$ brew install actionlint
$ actionlint -version
1.3.0
downloaded from release page

$ go install github.com/rhysd/actionlint/cmd/actionlint@v1.3.0
go: downloading github.com/rhysd/actionlint v1.3.0
$ actionlint -version
v1.3.0
built from source
```

Example of running actionlint on GitHub Actions forcing to enable color output:

```yaml
- name: Check workflow files
  run: |
    bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
    ./actionlint -color
  shell: bash
```

[Changes][v1.3.0]


<a name="v1.2.0"></a>
# [v1.2.0](https://github.com/rhysd/actionlint/releases/tag/v1.2.0) - 25 Jun 2021

- [pyflakes](https://github.com/PyCQA/pyflakes) integration was added. If `pyflakes` is installed on your system, actionlint checks Python scripts in `run:` (when `shell: python`) with it. See [the rule document](https://github.com/rhysd/actionlint#check-pyflakes-integ) for more details.
- Error handling while running rule checkers was improved. When some internal error occurs while applying rules, actionlint stops correctly due to the error. Previously, such errors were only shown in debug logs and actionlint continued checks.
- Fixed sanitizing `${{ }}` expressions in scripts before passing them to shellcheck or pyflakes. Previously expressions were not correctly sanitized when `}}` came before `${{`.

[Changes][v1.2.0]


<a name="v1.1.2"></a>
# [v1.1.2](https://github.com/rhysd/actionlint/releases/tag/v1.1.2) - 21 Jun 2021

- Run `shellcheck` command for scripts at `run:` in parallel. Since executing an external process is heavy and running shellcheck was bottleneck of actionlint, this brought better performance. In my environment, it was **more than 3x faster** than before.
- Sort errors by their positions in the source file.

[Changes][v1.1.2]


<a name="v1.1.1"></a>
# [v1.1.1](https://github.com/rhysd/actionlint/releases/tag/v1.1.1) - 20 Jun 2021

- [`download-actionlint.yaml`](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) now sets `executable` output when it is run in GitHub Actions environment. Please see [instruction in 'Install' document](https://github.com/rhysd/actionlint#ci-services) for the usage.
- Redundant type `ArrayDerefType` was removed. Instead, [`Deref` field](https://pkg.go.dev/github.com/rhysd/actionlint#ArrayType) is now provided in `ArrayType`.
- Fix crash on broken YAML input.
- `actionlint -version` returns correct version string even if the `actionlint` command was installed via `go install`.

[Changes][v1.1.1]


<a name="v1.1.0"></a>
# [v1.1.0](https://github.com/rhysd/actionlint/releases/tag/v1.1.0) - 19 Jun 2021

- Ignore [SC1091](https://github.com/koalaman/shellcheck/wiki/SC1091) and [SC2194](https://github.com/koalaman/shellcheck/wiki/SC2194) on running shellcheck. These are reported as false positives due to sanitization of `${{ ... }}`. See [the check doc](https://github.com/rhysd/actionlint#check-shellcheck-integ) to know the sanitization.
- actionlint replaces `${{ }}` in `run:` scripts before passing them to shellcheck. v1.0.0 replaced `${{ }}` with whitespaces, but it caused syntax errors in some scripts (e.g. `if ${{ ... }}; then ...`). Instead, v1.1.0 replaces `${{ }}` with underscores. For example, `${{ matrix.os }}` is replaced with `________________`.
- Add [`download-actionlint.bash`](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) script to download pre-built binaries easily on CI services. See [installation document](https://github.com/rhysd/actionlint#on-ci) for the usage.
- Better error message on lexing `"` in `${{ }}` expression since double quote is usually misused for string delimiters
- `-ignore` option can now be specified multiple times.
- Fix `github.repositoryUrl` was not correctly resolved in `${{ }}` expression
- Reports an error when `if:` condition does not use `${{ }}` but the expression contains any operators. [The official document](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsif) prohibits this explicitly to avoid conflicts with YAML syntax.
- Clarify that the version of this repository is for `actionlint` CLI tool, not for library. It means that the APIs may have breaking changes on minor or patch version bumps.
- Add more tests and refactor some code. Enumerating quoted items in error message is now done more efficiently and in deterministic order.

[Changes][v1.1.0]


<a name="v1.0.0"></a>
# [v1.0.0](https://github.com/rhysd/actionlint/releases/tag/v1.0.0) - 16 Jun 2021

First release :tada:

See documentation for more details:

- [Installation](https://github.com/rhysd/actionlint#install)
- [Usage](https://github.com/rhysd/actionlint#usage)
- [Checks done by actionlint](https://github.com/rhysd/actionlint#checks)

[Changes][v1.0.0]


[v1.3.1]: https://github.com/rhysd/actionlint/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/rhysd/actionlint/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/rhysd/actionlint/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/rhysd/actionlint/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/rhysd/actionlint/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/rhysd/actionlint/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/rhysd/actionlint/tree/v1.0.0

 <!-- Generated by changelog-from-release -->
