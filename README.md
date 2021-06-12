actionlint
==========
[![CI Badge][]][CI]

[actionlint][repo] is a static checker for GitHub Actions workflow files.

Features:

- **Syntax check for workflow files**: When some keys are missing or unexpected, actionlint reports them
- **Strong type check for `${{ }}` expressions**: It can catch access to not existing property as well as type mismatches
- **[shellcheck][] integration** for `run:` section
- **Other several useful checks**; dependencies check for `needs:` section, runner label validation, cron syntax validation, ...

See ['Checks' section](#checks) for full list of checks done by actionlint.

**Example of broken workflow:**

```yaml
on:
  push:
    branch: main
jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, linux-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v2
      - uses: actions/cache@v2
        with:
          path: ~/.npm
          key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
        if: github.repository.permissions.admin == true
      - run: npm install && npm test
```

**Output from actionlint:**

```
example.yaml:3:5: unexpected key "branch" for "push" section. expected one of "types", "branches", "branches-ignore", "tags", "tags-ignore", "paths", "paths-ignore", "workflows" [syntax-check]
3|     branch: main
 |     ^~~~~~~
example.yaml:9:28: label "linux-latest" is unknown. available labels are "windows-latest", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
9|         os: [macos-latest, linux-latest]
 |                            ^~~~~~~~~~~~~
example.yaml:17:13: receiver of object dereference "permissions" must be type of object but got "string" [expression]
17|         if: github.repository.permissions.admin == true
  |             ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
example.yaml:16:20: property "platform" is not defined in object type {os: string} [expression]
16|           key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
  |                    ^~~~~~~~~~~~~~~
```

actionlint tries to catch errors as much as possible and tries to make false positive as minimal as possible.

## Why?

- **Running workflow is very time consuming.** You need to push the changes and wait until the workflow runs on GitHub even
  if it contains some trivial mistakes. [act][] is useful to run the workflow locally. But it is not suitable for CI and
  still time consuming when your workflow gets larger.
- **Checks of workflow files by GitHub is very loose.** It reports no error even if unexpected keys are in mappings
  (meant that some typos in keys). And also it reports no error when accessing to property which is actually not existing.
  For example `matrix.foo` when no `foo` is defined in `matrix:` section, it is evaluated to `null` and causes no error.
- **Some mistakes silently breaks workflow.** Most common case I saw is specifying missing property to cache key. In the case
  cache silently does not work properly but workflow itself runs without error. So you might not notice the mistake forever.

## Install

As of now, no prebuilt binary is provided. Install [Go][] toolchain and build actionlint from source.

```
go get github.com/rhysd/actionlint/cmd/actionlint
```

## Usage

With no argument, actionlint finds all workflow files in the current repository and checks them.

```
actionlint
```

When path to YAML workflow files are given, actionlint checks them.

```
actionlint path/to/workflow1.yaml path/to/workflow2.yaml
```

See `actionlint -h` for all flags and options.

<a name="checks"></a>
## Checks

This section describes all checks done by actionlint with example input and output.

Note: actionlint focuses on catching mistakes in workflow files. If you want some code style checks, please consider to
use a general YAML checker like [yamllint][].

TODO


## Configuration file

Configuration file `actionlint.yaml` or `actionlint.yml` can be put in `.github` directory.

You don't need to write first configuration file by your hand. `actionlint` command can generate a default configuration.

```
actionlint -init-config
vim .github/actionlint.yaml
```

Since the author tries to keep configuration file as minimal as possible, currently only one item can be configured.

```yaml
self-hosted-runner:
  # Labels of self-hosted runner in array of string
  labels:
    - linux.2xlarge
    - windows-latest-xl
    - linux-multi-gpu
```

- `self-hosted-runner`: Configuration for your self-hosted runner environment
  - `labels`: Label names added to your self-hoted runners as list of string

Note that configuration file is optional. The author tries to keep configuration file as minimal as possible not to
bother users to configure behavior of it. Running actionlint without configuration file would work fine in most cases.

## Use actionlint as library

actionlint can be used from Go programs. See [the documentation][apidoc] to know the list of all APIs. Followings are
unexhaustive list of interesting APIs.

- `Linter` manages linter lifecycle and applies checks to given files. If you want to run actionlint checks in your
  program, please use this struct.
- `Project` and `Projects` detect a project (Git repository) in a given directory path and find configuration in it.
- `Workflow`, `Job`, `Step`, ... are nodes of workflow syntax tree. `Workflow` is a root node.
- `Parse()` parses given contents into a workflow syntax tree. It tries to find syntax errors as much as possible and
  returns found errors as slice.
- `Pass` is a visitor to traverse a workflow syntax tree. Multiple passes can be applied at single pass using `Visitor`.
- `Rule` is an interface for rule checkers and `RuneBase` is a base struct to implement a rule checker.
  - `RuleExpression` is a rule checker to check expression syntax in `${{ }}`.
  - `RuleShellcheck` is a rule checker to apply `shellcheck` command to `run:` sections and collect errors from it.
  - `RuleJobNeeds` is a rule checker to check dependencies in `needs:` section. It can detect cyclic dependencies.
  - ...
- `ExprLexer` lexes expression syntax in `${{ }}` and returns slice of `Token`.
- `ExprParser` parses given slice of `Token` and returns syntax tree for expression in `${{ }}`. `ExprNode` is an
  interface for nodes in the expression syntax tree.
- `ExprType` is an interface of types in expression syntax `${{ }}`. `ObjectType`, `ArrayType`, `StringType`,
  `NumberType`, ... are structs to represent actual types of expression.
- `ExprSemanticsChecker` checks semantics of expression syntax `${{ }}`. It traverses given expression syntax tree and
  deduces its type, checking types and resolving variables (contexts).



## Testing

## License

actionlint is distributed under [the MIT license](./LICENSE.txt).

[CI Badge]: https://github.com/rhysd/actionlint/workflows/CI/badge.svg?branch=master&event=push
[CI]: https://github.com/rhysd/actionlint/actions?query=workflow%3ACI+branch%3Amaster
[repo]: https://github.com/rhysd/actionlint
[shellcheck]: https://github.com/koalaman/shellcheck
[yamllint]: https://github.com/adrienverge/yamllint
[act]: https://github.com/nektos/act
[Go]: https://golang.org/
[apidoc]: https://pkg.go.dev/github.com/rhysd/actionlint
