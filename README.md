actionlint
==========
[![CI Badge][]][CI]
[![API Document][api-badge]][apidoc]

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
        if: ${{ github.repository.permissions.admin == true }}
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
example.yaml:17:17: receiver of object dereference "permissions" must be type of object but got "string" [expression]
17|         if: ${{ github.repository.permissions.admin == true }}
  |                 ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
example.yaml:16:20: property "platform" is not defined in object type {os: string} [expression]
16|           key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
  |                    ^~~~~~~~~~~~~~~
```

actionlint tries to catch errors as much as possible and tries to make false positive as minimal as possible.

# Why?

- **Running workflow is very time consuming.** You need to push the changes and wait until the workflow runs on GitHub even
  if it contains some trivial mistakes. [act][] is useful to run the workflow locally. But it is not suitable for CI and
  still time consuming when your workflow gets larger.
- **Checks of workflow files by GitHub is very loose.** It reports no error even if unexpected keys are in mappings
  (meant that some typos in keys). And also it reports no error when accessing to property which is actually not existing.
  For example `matrix.foo` when no `foo` is defined in `matrix:` section, it is evaluated to `null` and causes no error.
- **Some mistakes silently breaks workflow.** Most common case I saw is specifying missing property to cache key. In the case
  cache silently does not work properly but workflow itself runs without error. So you might not notice the mistake forever.

# Install

## [Homebrew][homebrew] on macOS

Tap this repository and install `actionlint` package.

```
brew tap "rhysd/actionlint" "https://github.com/rhysd/actionlint"
brew install actionlint
```

## Prebuilt binaries

Download an archive file from [releases page][releases], unarchive it and put the executable file at a directory in `$PATH`.

Prebuilt binaries are built at each release by CI for the following OS and arch:

- macOS (x86_64)
- Linux (i386, x86_64, arm32, arm64)
- Windows (i386, x86_64)
- FreeBSD (i386, x86_64)

## On CI

Please try [the download script](./scripts/download-actionlint.bash). It downloads the latest version of actionlint to
the current directory automatically.
Here is an example of simple workflow to run actionlint on GitHub Actions.

```yaml
name: Lint GitHub Actions workflows
on: [push, pull_request]

jobs:
  actionlint:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v2
      - name: Download actionlint
        run: bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
        shell: bash
      - name: Check workflow files
        run: ./actionlint
```

## Build from source

[Go][] toolchain is necessary. It builds the latest `main` branch.

```
go get github.com/rhysd/actionlint/cmd/actionlint
```

# Usage

With no argument, actionlint finds all workflow files in the current repository and checks them.

```
actionlint
```

When paths to YAML workflow files are given, actionlint checks them.

```
actionlint path/to/workflow1.yaml path/to/workflow2.yaml
```

See `actionlint -h` for all flags and options.

<a name="checks"></a>
# Checks

This section describes all checks done by actionlint with example input and output.

Note that actionlint focuses on catching mistakes in workflow files. If you want some code style checks, please consider to
use a general YAML checker like [yamllint][].

- [Unexpected keys](#check-unexpected-keys)
- [Missing required keys or key duplicates](#check-missing-required-duplicate-keys)
- [Unexpected empty mappings](#check-empty-mapping)
- [Unexpected mapping values](#check-mapping-values)
- [Syntax check for expression `${{ }}`](#check-syntax-expression)
- [Type checks for expression syntax in `${{ }}`](#check-type-check-expression)
- [Contexts and built-in functions](#check-contexts-and-builtin-func)
- [Contextual typing for `steps.<step_id>` objects](#check-contextual-step-object)
- [Contextual typing for `matrix` object](#check-contextual-matrix-object)
- [Contextual typing for `needs` object](#check-contextual-needs-object)
- [shellcheck integration](#check-shellcheck-integ)
- [Job dependencies validation](#check-job-deps)
- [Matrix values](#check-matrix-values)
- [Webhook events validation](#check-webhook-events)
- [CRON syntax check at `schedule:`](#check-cron-syntax)
- [Runner labels](#check-runner-labels)
- [Action format in `uses:`](#check-action-format)
- [Local action inputs validation at `with:`](#check-local-action-inputs)
- [Shell name validation at `shell:`](#check-shell-names)
- [Job ID and step ID uniqueness](#check-job-step-ids)
- [Hardcoded credentials](#check-hardcoded-credentials)
- [Environment variable names](#check-env-var-names)

<a name="check-unexpected-keys"></a>
## Unexpected keys

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    step:
```

Output:

```
test.yaml:5:5: unexpected key "step" for "job" section. expected one of "name", "needs", "runs-on", "permissions", "environment", "concurrency", "outputs", "env", "defaults", "if", "steps", "timeout-minutes", "strategy", "continue-on-error", "container", "services" [syntax-check]
5|     step:
 |     ^~~~~
```

[Workflow syntax][syntax-doc] defines what keys can be defined in which mapping object. When other keys are defined, they
are simply ignored and don't affect workflow behavior. It means typo in keys is not detected by GitHub.

actionlint can detect unexpected keys while parsing workflow syntax and report them as error.

<a name="check-missing-required-duplicate-keys"></a>
## Missing required keys or key duplicates

Example input:

```yaml
on: push
jobs:
  test:
    steps:
      - run: echo 'hello'
  TEST:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'bye'
```

Output:

```
test.yaml:6:3: key "test" is duplicate in "jobs" section. previously defined at line:3,col:3. note that key names are case insensitive [syntax-check]
6|   TEST:
 |   ^~~~~
test.yaml:4:5: "runs-on" section is missing in job "test" [syntax-check]
4|     steps:
 |     ^~~~~~
```

Some mappings must include specific keys. For example, job mapping must include `runs-on:` and `steps:`.

And duplicate in keys is not allowed. In workflow syntax, comparing keys is **case insensitive**. For example, job ID
`test` in lower case and job ID `TEST` in upper case are not able to exist in the same workflow.

actionlint checks these missing required keys and duplicate of keys while parsing, and reports an error.

<a name="check-empty-mapping"></a>
## Unexpected empty mappings

Example input:

```yaml
on: push
jobs:
```

Output:

```
test.yaml:2:6: "jobs" section should not be empty. please remove this section if it's unnecessary [syntax-check]
2| jobs:
 |      ^
```

Some mappings and sequences should not be empty. For example, `steps:` must include at least one step.

actionlint checks such mappings and sequences are not empty while parsing, and reports the empty mappings and sequences as error.

<a name="check-mapping-values"></a>
## Unexpected mapping values

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      issues: foo
    continue-on-error: foo
    steps:
      - run: echo 'hello'
```

Output:

```
test.yaml:6:15: permission must be one of "none", "read", "write" but got "foo" [syntax-check]
6|       issues: foo
 |               ^~~
test.yaml:7:24: expecting a string with ${{...}} expression or boolean literal "true" or "false", but found plain text node [syntax-check]
7|     continue-on-error: foo
 |                        ^~~
```

Some mapping's values are restricted to some constant strings. For example, values of `permissions:` mappings should be
one of `none`, `read`, `write`. And several mapping values expect boolean value like `true` or `false`.

actionlint checks such constant strings are used properly while parsing, and reports an error when unexpected string
value is specified.

<a name="check-syntax-expression"></a>
## Syntax check for expression `${{ }}`

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # " is not available for string literal delimiter
      - run: echo '${{ "hello" }}'
      # + operator does not exist
      - run: echo '${{ 1 + 1 }}'
      # Missing ')' paren
      - run: echo "${{ toJson(hashFiles('**/lock', '**/cache/') }}"
      # unexpected end of input
      - run: echo '${{ github.event. }}'
```

Output:

```
test.yaml:7:25: got unexpected character '"' while lexing expression, expecting '_', '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', ... [expression]
7|       - run: echo '${{ "hello" }}'
 |                         ^~~~~~
test.yaml:9:27: got unexpected character '+' while lexing expression, expecting '_', '\'', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', ... [expression]
9|       - run: echo '${{ 1 + 1 }}'
 |                           ^
test.yaml:11:65: unexpected end of input while parsing arguments of function call. expecting ",", ")" [expression]
11|       - run: echo "${{ toJson(hashFiles('**/lock', '**/cache/') }}"
  |                                                                 ^~~
test.yaml:13:38: unexpected end of input while parsing object property dereference like 'a.b' or array element dereference like 'a.*'. expecting "IDENT", "*" [expression]
13|       - run: echo '${{ github.event. }}'
  |                                      ^~~
```

actionlint lexes and parses expression in `${{ }}` following [the expression syntax document][expr-doc]. It can detect
many syntax errors like invalid characters, missing parens, unexpected end of input, ...

<a name="check-type-check-expression"></a>
## Type checks for expression syntax in `${{ }}`

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # Basic type error like index access to object
      - run: echo '${{ env[0] }}'
      # Properties in objects are strongly typed. So missing property can be caught
      - run: echo '${{ job.container.os }}'
      # github.repository is string. So accessing .owner is invalid
      - run: echo '${{ github.repository.owner }}'
```

Output:

```
test.yaml:7:28: property access of object must be type of string but got "number" [expression]
7|       - run: echo '${{ env[0] }}'
 |                            ^~
test.yaml:9:24: property "os" is not defined in object type {id: string; network: string} [expression]
9|       - run: echo '${{ job.container.os }}'
 |                        ^~~~~~~~~~~~~~~~
test.yaml:11:24: receiver of object dereference "owner" must be type of object but got "string" [expression]
11|       - run: echo '${{ github.repository.owner }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~
```

Type checks for expression syntax in `${{ }}` are done by semantics checker. Note that actual type checks by GitHub Actions
runtime is loose. For example any object value can be assigned into string value as string `"Object"`. But such loose
conversions are bugs in almost all cases. actionlint checks types more strictly.

There are two types of object types internally. One is an object which is strict for properties, which causes a type error
when trying to access to unknown properties. And another is an object which is not strict for properties, which allows to
access to unknown properties. In the case, accessing to unknown property is typed as `any`.

When the type check cannot be done statically, the type is deduced to `any` (e.g. return type from `toJSON()`).

And `${{ }}` can be used for expanding values.

Example input:

```yaml
on: push
jobs:
  test:
    strategy:
      matrix:
        env_string:
          - 'FOO=BAR'
          - 'FOO=PIYO'
        env_object:
          - FOO: BAR
          - FOO: PIYO
    runs-on: ubuntu-latest
    steps:
      # Expanding object at 'env:' section
      - run: echo "$FOO"
        env: ${{ matrix.env_object }}
      # String value cannot be expanded as object
      - run: echo "$FOO"
        env: ${{ matrix.env_string }}
```

Output:

```
test.yaml:19:14: type of expression at "env" must be object but found type string [expression]
19|         env: ${{ matrix.env_string }}
  |              ^~~
```

In above example, environment variables mapping is expanded at `env:` section. actionlint checks type of the expanded value.

<a name="check-contexts-and-builtin-func"></a>
## Contexts and built-in functions

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # Access to undefined context
      - run: echo '${{ unknown_context }}'
      # Access to undefined property of context
      - run: echo '${{ github.events }}'
      # Calling undefined function (start's'With is correct)
      - run: echo "${{ startWith('hello, world', 'lo,') }}"
      # Wrong number of arguments
      - run: echo "${{ startsWith('hello, world') }}"
      # Wrong type of parameter
      - run: echo "${{ startsWith('hello, world', github.event) }}"
      # Function overloads can be handled properly. contains() has string version and array version
      - run: echo "${{ contains('hello, world', 'lo,') }}"
      - run: echo "${{ contains(github.event.labels.*.name, 'enhancement') }}"
      # format() has special check for formating string
      - run: echo "${{ format('{0}{1}', 1, 2, 3) }}"
```

Output:

```
test.yaml:7:24: undefined variable "unknown_context". available variables are "env", "github", "job", "matrix", "needs", "runner", "secrets", "steps", "strategy" [expression]
7|       - run: echo '${{ unknown_context }}'
 |                        ^~~~~~~~~~~~~~~
test.yaml:9:24: property "events" is not defined in object type {workspace: string; env: string; event_name: string; event_path: string; ...} [expression]
9|       - run: echo '${{ github.events }}'
 |                        ^~~~~~~~~~~~~
test.yaml:11:24: undefined function "startWith". available functions are "always", "cancelled", "contains", "endswith", "failure", "format", "fromjson", "hashfiles", "join", "startswith", "success", "tojson" [expression]
11|       - run: echo "${{ startWith('hello, world', 'lo,') }}"
  |                        ^~~~~~~~~~~~~~~~~
test.yaml:13:24: number of arguments is wrong. function "startsWith(string, string) -> bool" takes 2 parameters but 1 arguments are provided [expression]
13|       - run: echo "${{ startsWith('hello, world') }}"
  |                        ^~~~~~~~~~~~~~~~~~
test.yaml:15:51: 2nd argument of function call is not assignable. "object" cannot be assigned to "string". called function type is "startsWith(string, string) -> bool" [expression]
15|       - run: echo "${{ startsWith('hello, world', github.event) }}"
  |                                                   ^~~~~~~~~~~~~
test.yaml:20:24: format string "{0}{1}" contains 2 placeholders but 3 arguments are given to format [expression]
20|       - run: echo "${{ format('{0}{1}', 1, 2, 3) }}"
  |                        ^~~~~~~~~~~~~~~~
```

[Contexts][contexts-doc] and [built-in functions][funcs-doc] are strongly typed. Typos in property access of contexts and
function names can be checked. And invalid function calls like wrong number of arguments or type mismatch at parameter also
can be checked thanks to type checker.

The semantics checker can properly handle that

- some functions are overloaded (e.g. `contains(str, substr)` and `contains(array, item)`)
- some parameters are optional (e.g. `join(strings, sep)` and `join(strings)`)
- some parameters are repeatable (e.g. `hashFiles(file1, file2, ...)`)

In addition, `format()` function has special check for placeholders in the first parameter which represents formatting string.

Note that context names and function names are case insensitive. For example, `toJSON` and `toJson` are the same function.

<a name="check-contextual-step-object"></a>
## Contextual typing for `steps.<step_id>` objects

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    outputs:
      # Step outputs can be used in job outputs since this section is evaluated after all steps were run
      foo: '${{ steps.get_value.outputs.name }}'
    steps:
      # Access to undefined step outputs
      - run: echo '${{ steps.get_value.outputs.name }}'
      # Outputs are set here
      - run: echo '::set-output name=foo::value'
        id: get_value
      # OK
      - run: echo '${{ steps.get_value.outputs.name }}'
      # OK
      - run: echo '${{ steps.get_value.conclusion }}'
  other:
    runs-on: ubuntu-latest
    steps:
      # Access to undefined step outputs. Step objects are job-local
      - run: echo '${{ steps.get_value.outputs.name }}'
```

Output:

```
test.yaml:10:24: property "get_value" is not defined in object type {} [expression]
10|       - run: echo '${{ steps.get_value.outputs.name }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:22:24: property "get_value" is not defined in object type {} [expression]
22|       - run: echo '${{ steps.get_value.outputs.name }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

Outputs of step can be accessed via `steps.<step_id>` objects. The `steps` context is dynamic:

- Accessing to the outputs before running the step are `null`
- Outputs of steps only in the job can be accessed. It cannot access to steps across jobs

It is actually common mistake to access to the wrong step outputs since people often forget fixing placeholders on
copying&pasting steps.

actionlint can catch the invalid accesses to step outputs and reports them as errors.

<a name="check-contextual-matrix-object"></a>
## Contextual typing for `matrix` object

Example input:

```yaml
on: push
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest]
        node: [14, 15]
        package:
          - name: 'foo'
            optional: true
          - name: 'bar'
            optional: false
        include:
          - node: 15
            npm: 7.5.4
    runs-on: ${{ matrix.os }}
    steps:
      # Access to undefined matrix value
      - run: echo '${{ matrix.platform }}'
      # Matrix value is strongly typed. Below line causes an error since matrix.package is {name: string, optional: bool}
      - run: echo '${{ matrix.package.dev }}'
      # OK
      - run: |
          echo 'os: ${{ matrix.os }}'
          echo 'node version: ${{ matrix.node }}'
          echo 'package: ${{ matrix.package.name }} (optional=${{ matrix.package.optional }})'
      # Additional matrix values in 'include:' are supported
      - run: echo 'npm version is specified'
        if: ${{ contains(matrix.npm, '7.5') }}
  test2:
    runs-on: ubuntu-latest
    steps:
      # Matrix values in other job is not accessible
      - run: echo '${{ matrix.os }}'
```

Output:

```
test.yaml:19:24: property "platform" is not defined in object type {os: string; node: number; package: {name: string; optional: bool}; npm: string} [expression]
19|       - run: echo '${{ matrix.platform }}'
  |                        ^~~~~~~~~~~~~~~
test.yaml:21:24: property "dev" is not defined in object type {name: string; optional: bool} [expression]
21|       - run: echo '${{ matrix.package.dev }}'
  |                        ^~~~~~~~~~~~~~~~~~
test.yaml:34:24: property "os" is not defined in object type {} [expression]
34|       - run: echo '${{ matrix.os }}'
  |                        ^~~~~~~~~
```

Types of `matrix` context is contextually checked by the semantics checker. Type of matrix values in `matrix:` section
is deduced from element values of its array. When the matrix value is an array of objects, objects' properties are checked
strictly like `package.name` in above example.

When type of the array elements is not persistent, type of the matrix value falls back to `any`.

```yaml
strategy:
  matrix:
    foo:
      - 'string value'
      - 42
      - {aaa: true, bbb: null}
    bar:
      - [42]
      - [true]
      - [{aaa: true, bbb: null}]
      - []
steps:
  # matrix.foo is any type value
  - run: echo ${{ matrix.foo }}
  # matrix.bar is array<any> type value
  - run: echo ${{ matrix.bar[0] }}
```

<a name="check-contextual-needs-object"></a>
## Contextual typing for `needs` object

Example input:

```yaml
on: push
jobs:
  install:
    outputs:
      installed: '...'
    runs-on: ubuntu-latest
    steps:
      - run: echo 'install something'
  prepare:
    outputs:
      prepared: '...'
    runs-on: ubuntu-latest
    steps:
      - run: echo 'parepare something'
      # ERROR: Outputs in other job is not accessble
      - run: echo '${{ needs.prepare.outputs.prepared }}'
  build:
    needs: [install, prepare]
    outputs:
      built: '...'
    runs-on: ubuntu-latest
    steps:
      # OK: Accessing to job results
      - run: echo 'build something with ${{ needs.install.outputs.installed }} and ${{ needs.prepare.outputs.prepared }}'
      # ERROR: Accessing to undefined output cases an error
      - run: echo '${{ needs.install.outputs.foo }}'
      # ERROR: Accessing to undefined job ID
      - run: echo '${{ needs.some_job }}'
  other:
    runs-on: ubuntu-latest
    steps:
      # ERROR: Cannot access to outptus across jobs
      - run: echo '${{ needs.build.outputs.built }}'
```

Output:

```
test.yaml:16:24: property "prepare" is not defined in object type {} [expression]
16|       - run: echo '${{ needs.prepare.outputs.prepared }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:26:24: property "foo" is not defined in object type {installed: string} [expression]
26|       - run: echo '${{ needs.install.outputs.foo }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:28:24: property "some_job" is not defined in object type {install: {outputs: {installed: string}; result: string}; prepare: {outputs: {prepared: string}; result: string}} [expression]
28|       - run: echo '${{ needs.some_job }}'
  |                        ^~~~~~~~~~~~~~
test.yaml:33:24: property "build" is not defined in object type {} [expression]
33|       - run: echo '${{ needs.build.outputs.built }}'
  |                        ^~~~~~~~~~~~~~~~~~~~~~~~~
```

Job dependencies can be defined at [`needs:`][needs-doc]. A job runs after all jobs defined in `needs:` are done.
Outputs from the jobs can be accessed only from jobs following them via [`needs` context][needs-context-doc].

actionlint defines type of `needs` variable contextually looking at each job's `outputs:` section and `needs:` section.

<a name="check-shellcheck-integ"></a>
## [shellcheck][] integration

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo $FOO
  test-win:
    runs-on: windows-latest
    steps:
      # Shell on Windows is PowerShell by default.
      # shellcheck is not run in this case.
      - run: echo $FOO
      # This script is run with bash due to 'shell:' configuration
      - run: echo $FOO
        shell: bash
```

Output:

```
test.yaml:6:9: shellcheck reported issue in this script: SC2086:info:1:6: Double quote to prevent globbing and word splitting [shellcheck]
6|       - run: echo $FOO
 |         ^~~~
test.yaml:14:9: shellcheck reported issue in this script: SC2086:info:1:6: Double quote to prevent globbing and word splitting [shellcheck]
14|       - run: echo $FOO
  |         ^~~~
```

[shellcheck][] is a famous linter for ShellScript. actionlint runs shellcheck for scripts at `run:` step in workflow.

actionlint detects which shell is used to run the scripts following [the documentation][shell-doc]. On Linux or macOS,
the default shell is `bash` and on Windows it is `pwsh`. Shell can be configured by `shell:` configuration at workflow level
or job level. Each step can configure shell to run scripts by `shell:`.

actionlint remembers the default shell and checks what OS the job runs on. Only when the shell is `bash` or `sh`, actionlint
applies shellcheck to scripts.

By default, actionlint checks if `shellcheck` command exists in your system and uses it when it is found. The `-shellcheck`
option on running `actionlint` command specifies the executable path of shellcheck. Setting empty string by `shellcheck=`
disables shellcheck integration explicitly.

Since both `${{ }}` expression syntax and ShellScript's variable access `$FOO` use `$`, remaining `${{ }}` confuses shellcheck.
To avoid it, actionlint replaces `${{ }}` with underscores. For example `echo '${{ matrix.os }}'` is replaced with
`echo '________________'`.

<a name="check-job-deps"></a>
## Job dependencies validation

Example input:

```yaml
on: push
jobs:
  prepare:
    needs: [build]
    runs-on: ubuntu-latest
    steps:
      - run: echo 'prepare'
  install:
    needs: [prepare]
    runs-on: ubuntu-latest
    steps:
      - run: echo 'install'
  build:
    needs: [install]
    runs-on: ubuntu-latest
    steps:
      - run: echo 'build'
```

Output:

```
test.yaml:8:3: cyclic dependencies in "needs" configurations of jobs are detected. detected cycle is "install" -> "prepare", "prepare" -> "build", "build" -> "install" [job-needs]
8|   install:
 |   ^~~~~~~~
```

Job dependencies can be defined at [`needs:`][needs-doc]. If cyclic dependencies exist, jobs never start to run. actionlint
detects cyclic dependencies in `needs:` sections of jobs and reports it as error.

actionlint also detects undefined jobs and duplicate jobs in `needs:` section.

Example input:

```yaml
on: push
jobs:
  foo:
    needs: [bar, BAR]
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'
  bar:
    needs: [unknown]
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'
```

Output:

```
test.yaml:4:18: job ID "BAR" duplicates in "needs" section. note that job ID is case insensitive [job-needs]
4|     needs: [bar, BAR]
 |                  ^~~~
test.yaml:8:3: job "bar" needs job "unknown" which does not exist in this workflow [job-needs]
8|   bar:
 |   ^~~~
```

<a name="check-matrix-values"></a>
## Matrix values

Example input:

```yaml
on: push
jobs:
  test:
    strategy:
      matrix:
        node: [10, 12, 14, 14]
        os: [ubuntu-latest, macos-latest]
        exclude:
          - node: 13
            os: ubuntu-latest
          - node: 10
            platform: ubuntu-latest
    runs-on: ${{ matrix.os }}
    steps:
      - run: echo ...
```

Output:

```
test.yaml:6:28: duplicate value "14" is found in matrix "node". the same value is at line:6,col:24 [matrix]
6|         node: [10, 12, 14, 14]
 |                            ^~~
test.yaml:9:19: value "13" in "exclude" does not exist in matrix "node" combinations. possible values are "10", "12", "14", "14" [matrix]
9|           - node: 13
 |                   ^~
test.yaml:12:13: "platform" in "exclude" section does not exist in matrix. available matrix configurations are "node", "os" [matrix]
12|             platform: ubuntu-latest
  |             ^~~~~~~~~
```

[`matrix:`][matrix-doc] defines combinations of multiple values. Nested `include:` and `exclude:` can add/remove specific
combination of matrix values. actionlint checks

- values in `exclude:` appear in `matrix:` or `include:`
- duplicate in variations of matrix values

<a name="check-webhook-events"></a>
## Webhook events validation

Example input:

```yaml
on:
  push:
    # Unexpected filter. 'branches' is correct
    branch: foo
  issues:
    # Unexpected type. 'opened' is correct
    types: created
  pullreq:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo ...
```

Output:

```
test.yaml:4:5: unexpected key "branch" for "push" section. expected one of "types", "branches", "branches-ignore", "tags", "tags-ignore", ... [syntax-check]
4|     branch: foo
 |     ^~~~~~~
test.yaml:7:12: invalid activity type "created" for "issues" Webhook event. available types are "opened", "edited", "deleted", "transferred", ... [events]
7|     types: created
 |            ^~~~~~~
test.yaml:8:3: unknown Webhook event "pullreq". see https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events for list of all Webhook event names [events]
8|   pullreq:
 |   ^~~~~~~~
```

At `on:`, Webhook events can be specified to trigger the workflow. [Webhook event documentation][webhook-doc] defines
which Webhook events are available and what types can be specified at `types:` for each event.

actionlint validates the Webhook configurations:

- unknown Webhook event name
- unknown type for Webhook event
- invalid filter names

<a name="check-cron-syntax"></a>
## CRON syntax check at `schedule:`

Example input:

```yaml
on:
  schedule:
    # Cron syntax is not correct
    - cron: '0 */3 * *'
    # Interval of scheduled job is too small (job runs too frequently)
    - cron: '* */3 * * *'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo ...
```

Output:

```
test.yaml:4:13: invalid CRON format "0 */3 * *" in schedule event: Expected exactly 5 fields, found 4: 0 */3 * * [events]
4|     - cron: '0 */3 * *'
 |             ^~
test.yaml:6:13: scheduled job runs too frequently. it runs once per 60 seconds [events]
6|     - cron: '* */3 * * *'
 |             ^~
```

To trigger a workflow in specific interval, [scheduled event][schedule-event-doc] can be defined in [POSIX CRON syntax][cron-syntax].

actionlint checks the CRON syntax and frequency of running the job. When a job is run more frequently than once per 1 minute,
actionlint reports it as error.

<a name="check-runner-labels"></a>
## Runner labels

Example input:

```yaml
on: push
jobs:
  test:
    strategy:
      matrix:
        runner:
          # OK
          - macos-latest
          # ERROR: Unknown runner
          - linux-latest
          # OK: Preset labels for self-hosted runner
          - [self-hosted, linux, x64]
          # OK: Single preset label for self-hosted runner
          - arm64
          # ERROR: Unknown label "gpu". Custom label must be defined in actionlint.yaml config file
          - gpu
    runs-on: ${{ matrix.runner }}
    steps:
      - run: echo ...

  test2:
    # ERROR: Too old macOS worker
    runs-on: macos-10.13
    steps:
      - run: echo ...
```

Output:

```
test.yaml:10:13: label "linux-latest" is unknown. available labels are "windows-latest", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
10|           - linux-latest
  |             ^~~~~~~~~~~~
test.yaml:16:13: label "gpu" is unknown. available labels are "windows-latest", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
16|           - gpu
  |             ^~~
test.yaml:23:14: label "macos-10.13" is unknown. available labels are "windows-latest", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
23|     runs-on: macos-10.13
  |              ^~~~~~~~~~~
```

GitHub Actions provides two kinds of job runners, [GitHub-hosted runner][gh-hosted-runner] and [self-hosted runner][self-hosted-runner].
Each runner has one or more labels. GitHub Actions runtime finds a proper runner based on label(s) specified at `runs-on:`
to run the job. So specifying proper labels at `runs-on:` is important.

actionlint checks proper label is used at `runs-on:` configuration. Even if an expression is used in the section like
`runs-on: ${{ matrix.foo }}`, actionlint parses the expression and resolves the possible values, then validates the values.

When you define some custom labels for your self-hosted runner, actionlint does not know the labels. Please set the label
names in [`actionlint.yaml` configuration file](#config-file) to let actionlint know them.


<a name="check-action-format"></a>
## Action format in `uses:`

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # ERROR: ref is missing
      - uses: actions/checkout
      # ERROR: owner name is missing
      - uses: checkout@v2
      # ERROR: tag is empty
      - uses: 'docker://image:'
      # ERROR: local action does not exist
      - uses: ./github/actions/my-action
```

Output:

```
test.yaml:7:15: specifying action "actions/checkout" in invalid format because ref is missng. available formats are "{owner}/{repo}@{ref}" or "{owner}/{repo}/{path}@{ref}" [action]
7|       - uses: actions/checkout
 |               ^~~~~~~~~~~~~~~~
test.yaml:9:15: specifying action "checkout@v2" in invalid format because owner is missing. available formats are "{owner}/{repo}@{ref}" or "{owner}/{repo}/{path}@{ref}" [action]
9|       - uses: checkout@v2
 |               ^~~~~~~~~~~
test.yaml:11:15: tag of Docker action should not be empty: "docker://image" [action]
11|       - uses: 'docker://image:'
  |               ^~~~~~~~~~~~~~~~~
test.yaml:13:15: Neither action.yaml nor action.yml is found in directory "github/actions/my-action" [action]
13|       - uses: ./github/actions/my-action
  |               ^~~~~~~~~~~~~~~~~~~~~~~~~~
```

Action needs to be specified in a format defined in [the document][action-uses-doc]. There are 3 types of actions:

- action hosted on GitHub: `owner/repo/path@ref`
- local action: `./path/to/my-action`
- Docker action: `docker://image:tag`

actionlint checks values at `uses:` sections follow one of these formats.

<a name="check-local-action-inputs"></a>
## Local action inputs validation at `with:`

`.github/actions/my-action/action.yaml`:

```yaml
name: 'My action'
author: 'rhysd <https://rhysd.github.io>'
description: 'my action'

inputs:
  name:
    description: your name
    default: anonymous
  message:
    description: message to this action
    required: true
  addition:
    description: additional information
    required: false

runs:
  using: 'node14'
  main: 'index.js'
```

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # missing required input "message"
      - uses: ./.github/actions/my-action
      # unexpected input "additions"
      - uses: ./.github/actions/my-action
        with:
          name: rhysd
          message: hello
          additions: foo, bar
```

Output:

```
test.yaml:7:15: missing input "message" which is required by action "My action" defined at "./.github/actions/my-action" [action]
7|       - uses: ./.github/actions/my-action
 |               ^~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:13:11: input "additions" is not defined in action "./.github/actions/my-action" defined at "My action". available inputs are "addition", "message", "name" [action]
13|           additions: foo, bar
  |           ^~~~~~~~~~
```

When a local action is run in `uses:` of `step:`, actionlint reads `action.yaml` file in the local action directory and
validates inputs at `with:` in the workflow are correct. Missing required inputs and unexpected inputs can be detected.

<a name="check-shell-names"></a>
## Shell name validation at `shell:`

Example input:

```yaml
on: push
jobs:
  linux:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hello'
        # ERROR: Unavailable shell
        shell: dash
      - run: echo 'hello'
        # ERROR: 'powershell' is only available on Windows
        shell: powershell
      - run: echo 'hello'
        # OK: 'powershell' is only available on Windows
        shell: powershell
  mac:
    runs-on: macos-latest
    defaults:
      run:
        # ERROR: default config is also checked. fish is not supported
        shell: fish
    steps:
      - run: echo 'hello'
        # OK: Custom shell
        shell: 'perl {0}'
  windows:
    runs-on: windows-latest
    steps:
      - run: echo 'hello'
        # ERROR: 'sh' is only available on Windows
        shell: sh
```

Output:

```
test.yaml:8:16: shell name "dash" is invalid. available names are "bash", "pwsh", "python", "sh" [shell-name]
8|         shell: dash
 |                ^~~~
test.yaml:11:16: shell name "powershell" is invalid on macOS or Linux. available names are "bash", "pwsh", "python", "sh" [shell-name]
11|         shell: powershell
  |                ^~~~~~~~~~
test.yaml:14:16: shell name "powershell" is invalid on macOS or Linux. available names are "bash", "pwsh", "python", "sh" [shell-name]
14|         shell: powershell
  |                ^~~~~~~~~~
test.yaml:20:16: shell name "fish" is invalid. available names are "bash", "pwsh", "python", "sh" [shell-name]
20|         shell: fish
  |                ^~~~
test.yaml:30:16: shell name "sh" is invalid on Windows. available names are "bash", "pwsh", "python", "cmd", "powershell" [shell-name]
30|         shell: sh
  |                ^~
```

Available shells for runners are defined in [the documentation][shell-doc]. actionlint checks shell names at `shell:`
configuration are properly using the available shells.

<a name="check-job-step-ids"></a>
## Job ID and step ID uniqueness

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hello'
        id: step_id
      - run: echo 'bye'
        # ERROR: Duplicate of step ID
        id: STEP_ID
  # ERROR: Duplicate of job ID
  TEST:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hello'
        # OK. Step ID uniqueness is job-local
        id: step_id
```

Output:

```
test.yaml:12:3: key "test" is duplicate in "jobs" section. previously defined at line:3,col:3. note that key names are case insensitive [syntax-check]
12|   TEST:
  |   ^~~~~
test.yaml:10:13: step ID "line:7,col:13" duplicates. previously defined at STEP_ID. step ID must be unique within a job. note that step ID is case insensitive [step-id]
10|         id: STEP_ID
  |             ^~~~~~~
```

Job IDs and step IDs in each job must be unique. IDs are compared in case insensitive. actionlint checks all job IDs
and step IDs and reports errors when some IDs duplicate.

<a name="check-hardcoded-credentials"></a>
## Hardcoded credentials

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    container:
      image: 'example.com/owner/image'
      credentials:
        username: user
        # ERROR: Hardcoded password
        password: pass
    services:
      redis:
        image: redis
        credentials:
          username: user
          # ERROR: Hardcoded password
          password: pass
    steps:
      - run: echo 'hello'
```

Output:

```
test.yaml:10:19: "password" section in "container" section should be specified via secrets. do not put password value directly [credentials]
10|         password: pass
  |                   ^~~~
test.yaml:17:21: "password" section in "redis" service should be specified via secrets. do not put password value directly [credentials]
17|           password: pass
  |                     ^~~~
```

[Credentials for container][credentials-doc] can be put in `container:` configuration. Password should be put in secrets
and the value should be expanded with `${{ }}` syntax at `password:`. actionlint checks hardcoded credentials and reports
them as error.

<a name="check-env-var-names"></a>
## Environment variable names

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    env:
      FOO=BAR: foo
      FOO BAR: foo
    steps:
      - run: echo 'hello'
```

Output:

```
test.yaml:6:7: environment variable name "foo=bar" is invalid. '&', '=' and spaces should not be contained [env-var]
6|       FOO=BAR: foo
 |       ^~~~~~~~
test.yaml:7:7: environment variable name "foo bar" is invalid. '&', '=' and spaces should not be contained [env-var]
7|       FOO BAR: foo
 |       ^~~
```

`=` must not be included in environment variable names. And `&` and spaces should not be included in them. In almost all
cases they are mistakes and they may cause some issues on using them in shell since they have special meaning in shell syntax.

actionlint checks environment variable names are correct in `env:` configuration.

<a name="config-file"></a>
# Configuration file

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
  - `labels`: Label names added to your self-hosted runners as list of string

Note that configuration file is optional. The author tries to keep configuration file as minimal as possible not to
bother users to configure behavior of actionlint. Running actionlint without configuration file would work fine in most
cases.

# Use actionlint as library

actionlint can be used from Go programs. See [the documentation][apidoc] to know the list of all APIs. It contains
workflow file parser built on top of `go-yaml/yaml`, expression `${{ }}` lexer/parser/checker, etc.
Followings are unexhaustive list of interesting APIs.

- `Linter` manages linter lifecycle and applies checks to given files. If you want to run actionlint checks in your
  program, please use this struct.
- `Project` and `Projects` detect a project (Git repository) in a given directory path and find configuration in it.
- `Config` represents structure of `actionlint.yaml` config file. It can be decoded by [go-yaml/yaml][go-yaml] library.
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

Note that the version of this repository is for command line tool `actionlint`. So it does not represent version of the
library, meant that patch version bump may introduce some breaking changes.

# Testing

- All examples in ['Checks' section](#checks) are tested in [`example_test.go`](./example_test.go)
- I cloned GitHub top 1000 repositories and extracted 1400+ workflow files. And I tried actionlint with the collected workflow
  files. All bugs found while the trial were fixed and I confirmed no more false positives.

# Bug reporting

When you 're seeing some bugs or false positives, it is helpful to [file a new issue][issue-form] with a minimal example
of input. Giving me some feedbacks like feature requests or idea of additional checks is also welcome.

# License

actionlint is distributed under [the MIT license](./LICENSE.txt).

[CI Badge]: https://github.com/rhysd/actionlint/workflows/CI/badge.svg?branch=main&event=push
[CI]: https://github.com/rhysd/actionlint/actions?query=workflow%3ACI+branch%3Amain
[api-badge]: https://pkg.go.dev/badge/github.com/rhysd/actionlint.svg
[apidoc]: https://pkg.go.dev/github.com/rhysd/actionlint
[repo]: https://github.com/rhysd/actionlint
[shellcheck]: https://github.com/koalaman/shellcheck
[yamllint]: https://github.com/adrienverge/yamllint
[act]: https://github.com/nektos/act
[homebrew]: https://brew.sh/
[releases]: https://github.com/rhysd/actionlint/releases
[Go]: https://golang.org/
[syntax-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
[expr-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
[contexts-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
[funcs-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#functions
[steps-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#steps-context
[needs-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idneeds
[needs-context-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#needs-context
[shell-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell
[matrix-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
[webhook-doc]: https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events
[schedule-event-doc]: https://docs.github.com/en/actions/reference/events-that-trigger-workflows#scheduled-events
[cron-syntax]: https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html#tag_20_25_07
[gh-hosted-runner]: https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners
[self-hosted-runner]: https://docs.github.com/en/actions/hosting-your-own-runners/about-self-hosted-runners
[action-uses-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
[action-metadata-doc]: https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
[go-yaml]: https://github.com/go-yaml/yaml
[credentials-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
[opengroup-env-vars]: https://pubs.opengroup.org/onlinepubs/007904875/basedefs/xbd_chap08.html
[issue-form]: https://github.com/rhysd/actionlint/issues/new
