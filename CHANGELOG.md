<a name="v1.5.2"></a>
# [v1.5.2](https://github.com/rhysd/actionlint/releases/tag/v1.5.2) - 02 Aug 2021

- Outputs of [dorny/paths-filter](https://github.com/dorny/paths-filter) are now not typed strictly because the action dynamically sets outputs which are not defined in its `action.yml`. actionlint cannot check such outputs statically (#18).
- [The table](https://github.com/rhysd/actionlint/blob/main/all_webhooks.go) for [Webhooks supported by GitHub Actions](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events) is now generated from the official document automatically with [script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-webhook-events). The table continues to be updated weekly by [the CI workflow](https://github.com/rhysd/actionlint/actions/workflows/generate.yaml).
- Improve error messages while lexing expressions as follows.
- Fix column numbers are off-by-one on some lexer errors.
- Fix checking invalid numbers where some digit follows zero in a hex number (e.g. `0x01`) or an exponent part of number (e.g. `1e0123`).
- Fix a parse error message when some tokens still remain after parsing finishes.
- Refactor the expression lexer to lex an input incrementally. It slightly reduces memory consumption.

Lex error until v1.5.1:

```test.yaml:9:26: got unexpected character '+' while lexing expression, expecting '_', '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z' [expression]```

Lex error from v1.5.2:

```test.yaml:9:26: got unexpected character '+' while lexing expression, expecting 'a'..'z', 'A'..'Z', '0'..'9', ''', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', '_' [expression]```

[Changes][v1.5.2]


<a name="v1.5.1"></a>
# [v1.5.1](https://github.com/rhysd/actionlint/releases/tag/v1.5.1) - 29 Jul 2021

- Improve checking the intervals of scheduled events (#14, #15). Since GitHub Actions [limits the interval to once every 5 minutes](https://github.blog/changelog/2019-11-01-github-actions-scheduled-jobs-maximum-frequency-is-changing/), actionlint now reports an error when a workflow is configured to be run once per less than 5 minutes.
- Skip checking inputs of [octokit/request-action](https://github.com/octokit/request-action) since it allows to specify arbitrary inputs though they are not defined in its `action.yml` (#16).
  - Outputs of the action are still be typed strictly. Only its inputs are not checked.
- The help text of `actionlint` is now hosted online: https://rhysd.github.io/actionlint/usage.html
- Add new fuzzing target for parsing glob patterns.

[Changes][v1.5.1]


<a name="v1.5.0"></a>
# [v1.5.0](https://github.com/rhysd/actionlint/releases/tag/v1.5.0) - 26 Jul 2021

- `action` rule now validates inputs of popular actions at `with:`. When a required input is not specified or an undefined input is specified, actionlint will report it.
  - Popular actions are updated automatically once a week and the data set is embedded to executable directly. The check does not need any network request and does not affect performance of actionlint. Sources of the actions are listed [here](https://github.com/rhysd/actionlint/blob/main/scripts/generate-popular-actions/main.go#L51). If you have some request to support new action, please report it at [the issue form](https://github.com/rhysd/actionlint/issues/new).
  - Please see [the document](https://github.com/rhysd/actionlint#check-popular-action-inputs) for example ([Playground](https://rhysd.github.io/actionlint#eJyFj0EKwjAQRfc9xV8I1UJbcJmVK+8xDYOpqUlwEkVq725apYgbV8PMe/Dne6cQkpiiOPtOVAFEljhP4Jqc1D4LqUsupnqgmS1IIgd5W0CNJCwKpGPvnbSatOHDbf/BwL2PRq0bYPmR9efXBdiMIwyJOfYDy7asqrZqBq9tucM0/TWXyF81UI5F0wbSlk4s67u5mMKFLL8A+h9EEw==)).
- `expression` rule now types outputs of popular actions (type of `steps.{id}.outputs` object) more strictly.
  - For example, `actions/cache@v2` sets `cache-hit` output. The outputs object is typed as `{ cache-hit: any }`. Previously it was typed as `any` which means no further type check was performed.
  - Please see the second example of [the document](https://github.com/rhysd/actionlint#check-contextual-step-object) ([Playground](https://rhysd.github.io/actionlint#eJyNTksKwjAQ3fcUbyFUC0nBZVauvIakMZjY0gRnokjp3W3TUl26Gt53XugVYiJXFPfQkCoAtsTzBR6pJxEmQ2pSz0l0etayRGwjLS5AIJElBW3Yh55qo42zp+dxlQF/Vcjkxrw8O7UhoLVvhd0wwGlyZ99Z2pdVVVeyC6YtDxjHH3PUUxiyjtq0+mZpmzENVrDGhVyVN8r8V4bEMfGKhPP8bfw7dlliH1xHWso=)).
- Outputs of local actions (their names start with `./`) are also typed more strictly as well as popular actions.
- Metadata (`action.yml`) of local actions are now cached to avoid reading and parsing `action.yml` files repeatedly for the same action.
- Add new rule `permissions` to check [permission scopes](https://docs.github.com/en/actions/reference/authentication-in-a-workflow#permissions-for-the-github_token) for default `secrets.GITHUB_TOKEN`. Please see [the document](https://github.com/rhysd/actionlint#permissions) for more details ([Playground](https://rhysd.github.io/actionlint/#eJxNjd0NwyAMhN89xS3AAmwDxBK0FCOMlfUDiVr16aTv/qR5dNNM1Hl8imqRph7nKJOJXhLVEzBZ51ZgWFMnq2TR2jRXw/Zu63/gBkDKnN7ftQethPF6GByOEOuDdXL/ldw+8eCUBZlrlQvntjLp)).
- Structure of [`actionlint.Permissions`](https://pkg.go.dev/github.com/rhysd/actionlint#Permissions) struct was changed. A parser no longer checks values of `permissions:` configuration. The check is now done by `permissions` rule.

[Changes][v1.5.0]


<a name="v1.4.3"></a>
# [v1.4.3](https://github.com/rhysd/actionlint/releases/tag/v1.4.3) - 21 Jul 2021

- Support new Webhook events [`discussion` and `discussion_comment`](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#discussion) (#8).
- Read file concurrently with limiting concurrency to number of CPUs. This improves performance when checking many files and disabling shellcheck/pyflakes integration.
- Support Linux based on musl libc by [the download script](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) (#5).
- Reduce number of goroutines created while running shellcheck/pyflakes processes. This has small impact on memory usage when your workflows have many `run:` steps.
- Reduce built binary size by splitting an external library which is only used for debugging into a separate command line tool.
- Introduce several micro benchmark suites to track performance.
- Enable code scanning for Go/TypeScript/JavaScript sources in actionlint repository.

[Changes][v1.4.3]


<a name="v1.4.2"></a>
# [v1.4.2](https://github.com/rhysd/actionlint/releases/tag/v1.4.2) - 16 Jul 2021

- Fix executables in the current directory may be used unexpectedly to run `shellcheck` or `pyflakes` on Windows. This behavior could be security vulnerability since an attacker might put malicious executables in shared directories. actionlint searched an executable with [`exec.LookPath`](https://pkg.go.dev/os/exec#LookPath), but it searched the current directory on Windows as [golang/go#43724](https://github.com/golang/go/issues/43724) pointed. Now actionlint uses [`execabs.LookPath`](https://pkg.go.dev/golang.org/x/sys/execabs#LookPath) instead, which does not have the issue. (ref: [sharkdp/bat#1724](https://github.com/sharkdp/bat/pull/1724))
- Fix issue caused by running so many processes concurrently. Since checking workflows by actionlint is highly parallelized, checking many workflow files makes too many `shellcheck` processes and opens many files in parallel. This hit OS resources limitation (issue #3). Now reading files is serialized and number of processes run concurrently is limited for fixing the issue. Note that checking workflows is still done in parallel so this fix does not affect actionlint's performance.
- Ensure cleanup processes even if actionlint stops due to some fatal issue while visiting a workflow tree.
- Improve fatal error message to know which workflow file caused the error.
- [Playground](https://rhysd.github.io/actionlint/) improvements
  - "Permalink" button was added to make permalink directly linked to the current workflow source code. The source code is embedded in hash of the URL.
  - "Check" button and URL input form was added to check workflow files on https://github.com or https://gist.github.com easily. Visit a workflow file on GitHub, copy the URL, paste it to the input form and click the button. It instantly fetches the workflow file content and checks it with actionlint.
  - `u=` URL parameter was added to specify GitHub or Gist URL like https://rhysd.github.io/actionlint/?u=https://github.com/rhysd/actionlint/blob/main/.github/workflows/ci.yaml

[Changes][v1.4.2]


<a name="v1.4.1"></a>
# [v1.4.1](https://github.com/rhysd/actionlint/releases/tag/v1.4.1) - 12 Jul 2021

- A pre-built executable for `darwin/arm64` (Apple M1) was added to CI (#1)
  - Managing `actionlint` command with Homebrew on M1 Mac is now available. See [the instruction](https://github.com/rhysd/actionlint#homebrew-on-macos) for more details
  - Since the author doesn't have M1 Mac and GitHub Actions does not support M1 Mac yet, the built binary is not tested
- Pre-built executables are now built with Go 1.16 compiler (previously it was 1.15)
- Fix error message is sometimes not in one line when the error message was caused by go-yaml/yaml parser
- Fix playground does not work on Safari browsers on both iOS and Mac since they don't support `WebAssembly.instantiateStreaming()` yet
- Make URLs in error messages clickable on playground
- Code base of playground was migrated from JavaScript to Typescript along with improving error handlings

[Changes][v1.4.1]


<a name="v1.4.0"></a>
# [v1.4.0](https://github.com/rhysd/actionlint/releases/tag/v1.4.0) - 09 Jul 2021

- New rule to validate [glob pattern syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet) to filter branches, tags and paths. For more details, see [documentation](https://github.com/rhysd/actionlint#check-glob-pattern).
  - syntax errors like missing closing brackets for character range `[..]`
  - invalid usage like `?` following `*`, invalid character range `[9-1]`, ...
  - invalid character usage for Git ref names (branch name, tag name)
    - ref name cannot start/end with `/`
    - ref name cannot contain `[`, `:`, `\`, ...
- Fix column of error position is off by one when the error is caused by quoted strings like `'...'` or `"..."`.
- Add `--norc` option to `shellcheck` command to check shell scripts in `run:` in order not to be affected by any user configuration.
- Improve some error messages
- Explain playground in `man` manual

[Changes][v1.4.0]


<a name="v1.3.2"></a>
# [v1.3.2](https://github.com/rhysd/actionlint/releases/tag/v1.3.2) - 04 Jul 2021

- [actionlint playground](https://rhysd.github.io/actionlint) was implemented thanks to WebAssembly. actionlint is now available on browser without installing anything. The playground does not send user's workflow content to any remote server.
- Some margins are added to code snippets in error message. See below examples. I believe it's easier to recognize code in bunch of error messages than before.
- Line number is parsed from YAML syntax error. Since errors from [go-yaml/go](https://github.com/go-yaml/yaml) don't have position information, previously YAML syntax errors are reported at line:0, col:0. Now line number is parsed from error message and set correctly (if error message includes line number).
- Code snippet is shown in error message even if column number of the error position is unknown.
- Fix error message on detecting duplicate of step IDs.
- Fix and improve validating arguments of `format()` calls.
- All rule documents have links to actionlint playground with example code.
- `man` manual covers usage of actionlint on CI services.

Error message until v1.3.1:

```
test.yaml:4:13: invalid CRON format "0 */3 * *" in schedule event: Expected exactly 5 fields, found 4: 0 */3 * * [events]
4|     - cron: '0 */3 * *'
 |             ^~
```

Error message at v1.3.2:

```
test.yaml:4:13: invalid CRON format "0 */3 * *" in schedule event: Expected exactly 5 fields, found 4: 0 */3 * * [events]
  |
4 |     - cron: '0 */3 * *'
  |             ^~
```

[Changes][v1.3.2]


<a name="v1.3.1"></a>
# [v1.3.1](https://github.com/rhysd/actionlint/releases/tag/v1.3.1) - 30 Jun 2021

- Files are checked in parallel. This made actionlint around 1.3x faster with 3 workflow files in my environment
- Manual for `man` command was added. `actionlint.1` is included in released archives. If you installed actionlint via Homebrew, the manual is also installed automatically
- `-version` now reports how the binary was built (Go version, arch, os, ...)
- Added [`Command`](https://pkg.go.dev/github.com/rhysd/actionlint#Command) struct to manage entire command lifecycle
- Order of checked files is now stable. When all the workflows in the current repository are checked, the order is sorted by file names
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


[v1.5.2]: https://github.com/rhysd/actionlint/compare/v1.5.1...v1.5.2
[v1.5.1]: https://github.com/rhysd/actionlint/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/rhysd/actionlint/compare/v1.4.3...v1.5.0
[v1.4.3]: https://github.com/rhysd/actionlint/compare/v1.4.2...v1.4.3
[v1.4.2]: https://github.com/rhysd/actionlint/compare/v1.4.1...v1.4.2
[v1.4.1]: https://github.com/rhysd/actionlint/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/rhysd/actionlint/compare/v1.3.2...v1.4.0
[v1.3.2]: https://github.com/rhysd/actionlint/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/rhysd/actionlint/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/rhysd/actionlint/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/rhysd/actionlint/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/rhysd/actionlint/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/rhysd/actionlint/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/rhysd/actionlint/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/rhysd/actionlint/tree/v1.0.0

 <!-- Generated by changelog-from-release -->
