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

### Unexpected keys

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

### Missing required keys or key duplicates

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

### Unexpected empty mappings

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

### Unexpected mapping values

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

Some mapping's values are stricted to some constant strings. For example, values of `permissions:` mappings should be
one of `none`, `read`, `write`. And several mapping values expect boolean value like `true` or `false`.

actionlint checks such constant strings are used properly while parsing, and reports an error when unexpected string
value is specified.

### Syntax check for expression `${{ }}`

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

### Type checks for expression syntax in `${{ }}`

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

When the type check cannot be done statically, the type is deducted to `any` (e.g. return type from `toJSON()`).

### Contexts and built-in functions

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

### Contextual type for `steps.<step_id>` objects

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

### Contextual type for `matrix` object

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
        if: contains(matrix.npm, '7.5')
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

### Contextual type for `needs` object

Example input:

```yaml
```

Output:

```
```

### 

Example input:

```yaml
```

Output:

```
```

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
bother users to configure behavior of actionlint. Running actionlint without configuration file would work fine in most
cases.

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
[syntax-doc]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
[expr-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
[contexts-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
[funcs-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#functions
[steps-doc]: https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#steps-context
