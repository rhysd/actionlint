actionlint
==========
[![CI Status][ci-badge]][ci]
[![API Document][apidoc-badge]][apidoc]

[actionlint][repo] is a static checker for GitHub Actions workflow files. [Try it online!][playground]

Features:

- **Syntax check for workflow files** to check unexpected or missing keys following [workflow syntax][syntax-doc]
- **Strong type check for `${{ }}` expressions** to catch several semantic errors like access to not existing property,
  type mismatches, ...
- **Actions usage check** to check that inputs at `with:` and outputs in `steps.{id}.outputs` are correct
- **Reusable workflow check** to check inputs/outputs/secrets of reusable workflows and workflow calls
- **[shellcheck][] and [pyflakes][] integrations** for scripts at `run:`
- **Security checks**; [script injection][script-injection-doc] by untrusted inputs, hard-coded credentials
- **Other several useful checks**; [glob syntax][filter-pattern-doc] validation, dependencies check for `needs:`,
  runner label validation, cron syntax validation, ...

See the [full list][checks] of checks done by actionlint.

<img src="https://github.com/rhysd/ss/blob/master/actionlint/main.gif?raw=true" alt="actionlint reports 7 errors" width="806" height="492"/>

**Example of broken workflow:**

```yaml
on:
  push:
    branch: main
    tags:
      - 'v\d+'
jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, linux-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - run: echo "Checking commit '${{ github.event.head_commit.message }}'"
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node_version: 18.x
      - uses: actions/cache@v4
        with:
          path: ~/.npm
          key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
        if: ${{ github.repository.permissions.admin == true }}
      - run: npm install && npm test
```

**actionlint reports 7 errors:**

```
test.yaml:3:5: unexpected key "branch" for "push" section. expected one of "branches", "branches-ignore", "paths", "paths-ignore", "tags", "tags-ignore", "types", "workflows" [syntax-check]
  |
3 |     branch: main
  |     ^~~~~~~
test.yaml:5:11: character '\' is invalid for branch and tag names. only special characters [, ?, +, *, \, ! can be escaped with \. see `man git-check-ref-format` for more details. note that regular expression is unavailable. note: filter pattern syntax is explained at https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet [glob]
  |
5 |       - 'v\d+'
  |           ^~~~
test.yaml:10:28: label "linux-latest" is unknown. available labels are "windows-latest", "windows-latest-8-cores", "windows-2025", "windows-2022", "windows-11-arm", "ubuntu-slim", "ubuntu-latest", "ubuntu-latest-4-cores", "ubuntu-latest-8-cores", "ubuntu-latest-16-cores", "ubuntu-24.04", "ubuntu-24.04-arm", "ubuntu-22.04", "ubuntu-22.04-arm", "macos-latest", "macos-latest-xlarge", "macos-latest-large", "macos-26-xlarge", "macos-26", "macos-15-intel", "macos-15-xlarge", "macos-15-large", "macos-15", "macos-14-xlarge", "macos-14-large", "macos-14", "self-hosted", "x64", "arm", "arm64", "linux", "macos", "windows". if it is a custom label for self-hosted runner, set list of labels in actionlint.yaml config file [runner-label]
   |
10 |         os: [macos-latest, linux-latest]
   |                            ^~~~~~~~~~~~~
test.yaml:13:41: "github.event.head_commit.message" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/reference/security/secure-use#good-practices-for-mitigating-script-injection-attacks for more details [expression]
   |
13 |       - run: echo "Checking commit '${{ github.event.head_commit.message }}'"
   |                                         ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:17:11: input "node_version" is not defined in action "actions/setup-node@v4". available inputs are "always-auth", "architecture", "cache", "cache-dependency-path", "check-latest", "node-version", "node-version-file", "registry-url", "scope", "token" [action]
   |
17 |           node_version: 18.x
   |           ^~~~~~~~~~~~~
test.yaml:21:20: property "platform" is not defined in object type {os: string} [expression]
   |
21 |           key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
   |                    ^~~~~~~~~~~~~~~
test.yaml:22:17: receiver of object dereference "permissions" must be type of object but got "string" [expression]
   |
22 |         if: ${{ github.repository.permissions.admin == true }}
   |                 ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

## Quick start

Install `actionlint` command by downloading [the released binary][releases] or by Homebrew or by `go install`. See
[the installation document][install] for more details like how to manage the command with several package managers
or run via Docker container.

```sh
go install github.com/rhysd/actionlint/cmd/actionlint@latest
```

Basically all you need to do is run the `actionlint` command in your repository. actionlint automatically detects workflows and
checks errors. actionlint focuses on finding out mistakes. It tries to catch errors as much as possible and make false positives
as minimal as possible.

```sh
actionlint
```

Another option to try actionlint is [the online playground][playground]. Your browser can run actionlint through WebAssembly.

See [the usage document][usage] for more details.

## Documents

- [Checks][checks]: Full list of all checks done by actionlint with example inputs, outputs, and playground links.
- [Installation][install]: Installation instructions. Prebuilt binaries, a Docker image, building from source, a download script
  (for CI), supports by several package managers are available.
- [Usage][usage]: How to use `actionlint` command locally or on GitHub Actions, the online playground, an official Docker image,
  and integrations with reviewdog, Problem Matchers, super-linter, pre-commit, VS Code.
- [Configuration][config]: How to configure actionlint behavior. Currently, the labels of self-hosted runners, the configuration
  variables, and ignore patterns of errors for each file paths can be set.
- [Go API][api]: How to use actionlint as Go library.
- [References][refs]: Links to resources.

## Bug reporting

When you see some bugs or false positives, it is helpful to [file a new issue][issue-form] with a minimal example
of input. Giving me some feedbacks like feature requests or ideas of additional checks is also welcome.

See the [contribution guide](./CONTRIBUTING.md) for more details.

## License

actionlint is distributed under [the MIT license](./LICENSE.txt).

[ci-badge]: https://github.com/rhysd/actionlint/actions/workflows/ci.yaml/badge.svg
[ci]: https://github.com/rhysd/actionlint/actions/workflows/ci.yaml
[apidoc-badge]: https://pkg.go.dev/badge/github.com/rhysd/actionlint.svg
[apidoc]: https://pkg.go.dev/github.com/rhysd/actionlint
[repo]: https://github.com/rhysd/actionlint
[playground]: https://rhysd.github.io/actionlint/
[shellcheck]: https://github.com/koalaman/shellcheck
[pyflakes]: https://github.com/PyCQA/pyflakes
[syntax-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
[filter-pattern-doc]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
[script-injection-doc]: https://docs.github.com/en/actions/reference/security/secure-use#good-practices-for-mitigating-script-injection-attacks
[releases]: https://github.com/rhysd/actionlint/releases
[checks]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/checks.md
[install]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/install.md
[usage]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/usage.md
[config]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/config.md
[api]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/api.md
[refs]: https://github.com/rhysd/actionlint/blob/v1.7.10/docs/reference.md
[issue-form]: https://github.com/rhysd/actionlint/issues/new
