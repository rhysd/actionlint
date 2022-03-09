All checks done by actionlint
=============================

This document describes all checks done by [actionlint](..) with example inputs, outputs, and playground links.

List of checks:

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
- [shellcheck integration for `run:`](#check-shellcheck-integ)
- [pyflakes integration for `run:`](#check-pyflakes-integ)
- [Script injection by potentially untrusted inputs](#untrusted-inputs)
- [Job dependencies validation](#check-job-deps)
- [Matrix values](#check-matrix-values)
- [Webhook events validation](#check-webhook-events)
- [Workflow dispatch event validation](#check-workflow-dispatch-events)
- [Glob filter pattern syntax validation](#check-glob-pattern)
- [CRON syntax check at `schedule:`](#check-cron-syntax)
- [Runner labels](#check-runner-labels)
- [Action format in `uses:`](#check-action-format)
- [Local action inputs validation at `with:`](#check-local-action-inputs)
- [Popular action inputs validation at `with:`](#check-popular-action-inputs)
- [Shell name validation at `shell:`](#check-shell-names)
- [Job ID and step ID uniqueness](#check-job-step-ids)
- [Hardcoded credentials](#check-hardcoded-credentials)
- [Environment variable names](#check-env-var-names)
- [Permissions](#permissions)
- [Reusable workflows](#check-reusable-workflows)
- [Job ID naming convention](#job-id-naming-convention)

Note that actionlint focuses on catching mistakes in workflow files. If you want some general code style checks, please consider
using a general YAML checker like [yamllint][].

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
test.yaml:5:5: unexpected key "step" for "job" section. expected one of "concurrency", "container", "continue-on-error", "defaults", "env", "environment", "if", "name", "needs", "outputs", "permissions", "runs-on", "secrets", "services", "steps", "strategy", "timeout-minutes", "uses", "with" [syntax-check]
  |
5 |     step:
  |     ^~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJzLz7NSKCgtzuDKyk8qtuJSUChJLS4B0QoKRaV5xbr5QPnSpNK8klLdnESQHFiquCS1wAoAMZgSwQ==)

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
test.yaml:3:3: "runs-on" section is missing in job "test" [syntax-check]
  |
3 |   test:
  |   ^~~~~
test.yaml:6:3: key "test" is duplicate in "jobs" section. previously defined at line:3,col:3. note that key names are case insensitive [syntax-check]
  |
6 |   TEST:
  |   ^~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJzLz7NSKCgtzuDKyk8qtuJSUChJLS4B0QoKxSWpBcUQpoKCrkJRKVBpanJGvoJ6RmpOTr46UCbENTgEogIoW6ybD1RRmlSaV1Kqm5MIMoiAOUmVqeoAYX8k5Q==)

Some mappings must include specific keys. For example, job mappings must include `runs-on:` and `steps:`.

And duplicate in keys is not allowed. In workflow syntax, comparing keys is **case insensitive**. For example, the job ID
`test` in lower case and the job ID `TEST` in upper case are not able to exist in the same workflow.

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
  |
2 | jobs:
  |      ^
```

[Playground](https://rhysd.github.io/actionlint#eJzLz7NSKCgtzuDKyk8qtgIAJQsE6g==)

Some mappings and sequences should not be empty. For example, `steps:` must include at least one step.

actionlint checks such mappings and sequences are not empty while parsing, and reports the empty mappings and sequences as error.

<a name="check-mapping-values"></a>
## Unexpected mapping values

Example input:

```yaml
on: push
jobs:
  test:
    strategy:
      # ERROR: Boolean value "true" or "false" is expected
      fail-fast: off
      # ERROR: Integer value is expected
      max-parallel: 1.5
    runs-on: ubuntu-latest
    steps:
      - run: sleep 200
        # ERROR: Float value is expected
        timeout-minutes: two minutes
```

Output:

```
test.yaml:6:18: expecting a string with ${{...}} expression or boolean literal "true" or "false", but found plain text node [syntax-check]
  |
6 |       fail-fast: off
  |                  ^~~
test.yaml:8:21: expected scalar node for integer value but found scalar node with "!!float" tag [syntax-check]
  |
8 |       max-parallel: 1.5
  |                     ^~~
test.yaml:13:26: expecting a string with ${{...}} expression or float number literal, but found plain text node [syntax-check]
   |
13 |         timeout-minutes: two minutes
   |                          ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJw1jssNAjEMRO9bxTQQtCBxSTdeyYGg/BTbArongXCy5o395Fo8msl9e9RD/AYoi84JiHZSvr1/CQgUkws0atQQFsz0co06pcTJ43y6fnm3Iq4OtR1W1FyiqV1WbvJXurnpIYm54bLvC48vYuZq6nIsNk499FmxwgdsuTVm)

Some mapping's values are restricted to some constant strings. Several mapping values expect boolean value like `true` or
`false`. And some mapping values expect integer or floating number values.

actionlint checks such constant strings are used properly while parsing, and reports an error when unexpected value is
specified.

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
test.yaml:7:24: got unexpected character '"' while lexing expression, expecting 'a'..'z', 'A'..'Z', '_', '0'..'9', ''', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', ' '. do you mean string literals? only single quotes are available for string delimiter [expression]
  |
7 |       - run: echo '${{ "hello" }}'
  |                        ^~~~~~~
test.yaml:9:26: got unexpected character '+' while lexing expression, expecting 'a'..'z', 'A'..'Z', '_', '0'..'9', ''', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', ' ' [expression]
  |
9 |       - run: echo '${{ 1 + 1 }}'
  |                          ^
test.yaml:11:65: unexpected end of input while parsing arguments of function call. expecting ",", ")" [expression]
   |
11 |       - run: echo "${{ toJson(hashFiles('**/lock', '**/cache/') }}"
   |                                                                 ^~~
test.yaml:13:38: unexpected end of input while parsing object property dereference like 'a.b' or array element dereference like 'a.*'. expecting "IDENT", "*" [expression]
   |
13 |       - run: echo '${{ github.event. }}'
   |                                      ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJx1jTEOwjAMRfeewoqQUgptxZoDMHCLJLJwIYor7LBUvTsNrHT48pfekz9nB3MRah4cxDUAiqL1ArxKlp43XkLJWvrkK/siUZzlZwH01XSAkRjsYVnAEKbEBtbV7ikXOG35L5gqKN+Ec0te6DollNZ23Zg4Pu0Zao0+Eo72uP0weyP3SamEAd+YdahjH8ffRDM=)

actionlint lexes and parses expression in `${{ }}` following [the expression syntax document][expr-doc]. It can detect
many syntax errors like invalid characters, missing parens, unexpected end of input, ...

<a name="check-type-check-expression"></a>
## Type checks for expression syntax in `${{ }}`

actionlint checks types of expressions in `${{ }}` placeholders of templates. The following types are supported by the type
checker.

| Type          | Description                                                                                | Notation                 |
|---------------|--------------------------------------------------------------------------------------------|--------------------------|
| Any           | Any value like `any` type in TypeScript. Fallback type when a value can no longer be typed | `any`                    |
| Number        | Number value (integer or float)                                                            | `number`                 |
| Bool          | Boolean value                                                                              | `bool`                   |
| String        | String value                                                                               | `string`                 |
| Null          | Type of `null` value                                                                       | `null`                   |
| Array         | Array of specific type elements                                                            | `array<T>`               |
| Loose object  | Object which can contain any properties                                                    | `object`                 |
| Strict object | Object whose properties are strictly typed                                                 | `{prop1: T1, prop2: T2}` |
| Map object    | Object who has specific type values like `env` context                                     | `{string => T}`          |

Type check by actionlint is more strict than GitHub Actions runtime.

- Only `any` and `number` are allowed to be converted to string implicitly
- Implicit conversion to `number` is not allowed
- Object, array, and null are not allowed to be evaluated at `${{ }}`

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # ERROR: `env` is object. Index access to object is invalid
      - run: echo '${{ env[0] }}'
      # ERROR: Properties in objects are strongly typed. Missing property can be caught
      - run: echo '${{ job.container.os }}'
      # ERROR: `github.repository` is string. Trying to access .owner property is invalid
      - run: echo '${{ github.repository.owner }}'
      # ERROR: Objects, arrays and null should not be evaluated at ${{ }} since the outputs are useless
      - run: echo '${{ env }}'
```

Output:

```
test.yaml:7:28: property access of object must be type of string but got "number" [expression]
  |
7 |       - run: echo '${{ env[0] }}'
  |                            ^~
test.yaml:9:24: property "os" is not defined in object type {id: string; network: string} [expression]
  |
9 |       - run: echo '${{ job.container.os }}'
  |                        ^~~~~~~~~~~~~~~~
test.yaml:11:24: receiver of object dereference "owner" must be type of object but got "string" [expression]
   |
11 |       - run: echo '${{ github.repository.owner }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~
test.yaml:13:20: object, array, and null values should not be evaluated in template with ${{ }} but evaluating the value of type {string => string} [expression]
   |
13 |       - run: echo '${{ env }}'
   |                    ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJx9jrEKAjEQRPv7iimEqxKs8yticTkWE5FsyO4qcty/m2jtVVO8N8xwCagmabpzlDABSqIjgWZFHHdu0YqaeyyDfZEoVflZgBtmAK2JMZ+2DVSel/MV+z7/M/qYX7nokgs1z3Lk3rImi75RZcnK7e351VtHlX5g4A+nCkLw)

Type checks for expression syntax in `${{ }}` are done by semantics checker. Note that actual type checks by GitHub Actions
runtime is loose.

Any object value can be assigned into string value as string `'Object'`. `echo '${{ env }}'` will be replaced with
`echo 'Object'`. And an array can also be converted into `'Array'` string. Such loose conversions are bugs in almost all cases.
actionlint checks types more strictly. actionlint checks values evaluated at `${{ }}` are not object (replaced with string
`'Object'`), array (replaced with string `'Array'`), nor null (replaced with string `''`). If you want to check a content of
object or array, use `toJSON()` function.

```
echo '${{ toJSON(github.event) }}'
```

There are two types of object types internally. One is an object which is strict for properties, which causes a type error
when trying to access to unknown properties. And another is an object which is not strict for properties, which allows to
access to unknown properties. In the case, accessing to unknown property is typed as `any`.

When the type check cannot be done statically, the type is deduced to `any` (e.g. return type of `toJSON()`).

As special case of `${{ }}`, it can be used for expanding object and array values.

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
      # OK: Expanding object at 'env:' section
      - run: echo "$FOO"
        env: ${{ matrix.env_object }}
      # ERROR: String value cannot be expanded as object
      - run: echo "$FOO"
        env: ${{ matrix.env_string }}
```

Output:

```
test.yaml:19:14: type of expression at "env" must be object but found type string [expression]
   |
19 |         env: ${{ matrix.env_string }}
   |              ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJydkM0KgzAQhO8+xSCCp/QBAj20h0JPFm89lUSCP7SJmE1pEd+9SVWseOtp2WHm20mM5midraLGSMsjgJSlMAFLnSBVvscNeAjq6te8AUo/b95T63LRAIb0lGX74yFPt+rlfM3SFcDIRhW0BngnhwdsxZD/qp3Tlhnf3UmnybG7CL2n2qq1M5AFJ4cqKoM48Yz49zpH0vfTu3ZLGwzDf/HxN3z8A4EEWVQ=)

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
  |
7 |       - run: echo '${{ unknown_context }}'
  |                        ^~~~~~~~~~~~~~~
test.yaml:9:24: property "events" is not defined in object type {workspace: string; env: string; event_name: string; event_path: string; ...} [expression]
  |
9 |       - run: echo '${{ github.events }}'
  |                        ^~~~~~~~~~~~~
test.yaml:11:24: undefined function "startWith". available functions are "always", "cancelled", "contains", "endswith", "failure", "format", "fromjson", "hashfiles", "join", "startswith", "success", "tojson" [expression]
   |
11 |       - run: echo "${{ startWith('hello, world', 'lo,') }}"
   |                        ^~~~~~~~~~~~~~~~~
test.yaml:13:24: number of arguments is wrong. function "startsWith(string, string) -> bool" takes 2 parameters but 1 arguments are given [expression]
   |
13 |       - run: echo "${{ startsWith('hello, world') }}"
   |                        ^~~~~~~~~~~~~~~~~~
test.yaml:15:51: 2nd argument of function call is not assignable. "object" cannot be assigned to "string". called function type is "startsWith(string, string) -> bool" [expression]
   |
15 |       - run: echo "${{ startsWith('hello, world', github.event) }}"
   |                                                   ^~~~~~~~~~~~~
test.yaml:20:24: format string "{0}{1}" does not contain placeholder {2}. remove argument which is unused in the format string [expression]
   |
20 |       - run: echo "${{ format('{0}{1}', 1, 2, 3) }}"
   |                        ^~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJydkNGKwjAQRd/9ikGEuJIWdd/6Iz5KWmeNazojnYkKpf9uorAorH3wKYR7zs0lTBWcovjJL9dSTQAURfMJ0EWSglMe60gai+Bydo9E8SQPCqDIZAXYeAYz63uIdCS+0LZhUrwqDIN5h+4P6mNd4hlJ5Q04zaCo63ST6LnxGAJbuHAXdsaCSRfzldzpqCv/yJ9Z9mX1aEf+AXcg+WD0n/r8WBlcjUHKRUmuxVSD5B012KZsvO6Hu9bp3PTLoV8NacHKwtrC9126AZ31neg=)

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
      # ERROR: Access to undefined step outputs
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
      # ERROR: Access to undefined step outputs. Step objects are job-local
      - run: echo '${{ steps.get_value.outputs.name }}'
```

Output:

```
test.yaml:10:24: property "get_value" is not defined in object type {} [expression]
   |
10 |       - run: echo '${{ steps.get_value.outputs.name }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:22:24: property "get_value" is not defined in object type {} [expression]
   |
22 |       - run: echo '${{ steps.get_value.outputs.name }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJytkEsOglAMRees4g5MGD0W0MS1GMAqGHwltHVC2Ls+Pg6MiTE66uCec9tUIqF3bbKLVEoZYKyWJjB41CCP3CuP5qErUzZH4ta76cIBJxFCvhtHqHGvxZntcCs752IFi1heGdOUz8IMbW5IewhcN/JFxYtHpGxhIZHAfTqJ5oJNANoj4dn7z/XvvFpi3bm2EldLrOHh42d/+80ddrSUCw==)

Outputs of step can be accessed via `steps.<step_id>` objects. The `steps` context is dynamic:

- Accessing to the outputs before running the step are `null`
- Outputs of steps only in the job can be accessed. It cannot access to steps across jobs

It is actually common mistake to access to the wrong step outputs since people often forget fixing placeholders on
copying&pasting steps. actionlint can catch the invalid accesses to step outputs and reports them as errors.

When the outputs are set by popular actions, the outputs object is more strictly typed.

Example input:

```yaml
on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # ERROR: The step is not run yet at this point
      - run: echo ${{ steps.cache.outputs.cache-hit }}
      # actions/cache sets cache-hit output
      - uses: actions/cache@v2
        id: cache
        with:
          key: ${{ hashFiles('**/*.lock') }}
          path: ./packages
      # OK
      - run: echo ${{ steps.cache.outputs.cache-hit }}
      # ERROR: Typo at output name
      - run: echo ${{ steps.cache.outputs.cache_hit }}
```

Output:

```
test.yaml:8:23: property "cache" is not defined in object type {} [expression]
  |
8 |       - run: echo ${{ steps.cache.outputs.cache-hit }}
  |                       ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:18:23: property "cache_hit" is not defined in object type {cache-hit: string} [expression]
   |
18 |       - run: echo ${{ steps.cache.outputs.cache_hit }}
   |                       ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyNTksKwjAQ3fcUbyFUC0nBZVauvIakMZjY0gRnokjp3W3TUl26Gt53XugVYiJXFPfQkCoAtsTzBR6pJxEmQ2pSz0l0etayRGwjLS5AIJElBW3Yh55qo42zp+dxlQF/Vcjkxrw8O7UhoLVvhd0wwGlyZ99Z2pdVVVeyC6YtDxjHH3PUUxiyjtq0+mZpmzENVrDGhVyVN8r8V4bEMfGKhPP8bfw7dlliH1xHWso=)

In the above example, [actions/cache][actions-cache] action sets `cache-hit` output so that following steps can know the
cache was hit or not. At line 8, the cache action is not run yet. So `cache` property does not exit in `steps` context yet.
On running the step whose ID is `cache`, `steps.cache` object is typed as `{outputs: {cache-hit: any}, conclusion: string, outcome: string}`.
At line 18, the expression has typo in output name. actionlint can check it because properties of `steps.cache.outputs` are
typed.

This strict outputs typing is also applied to local actions. Let's say we have the following local action.

```yaml
name: 'My action with output'
author: 'rhysd <https://rhysd.github.io>'
description: 'my action with outputs'

outputs:
  some_value:
    description: some value returned from this action

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
      # ERROR: The step is not yet run
      - run: echo ${{ steps.my_action.outputs.some_value }}
      # The action runs here and sets its outputs
      - uses: ./.github/actions/my-action-with-output
        id: my_action
      # OK
      - run: echo ${{ steps.my_action.outputs.some_value }}
      # ERROR: No output named 'some-value' (typo)
      - run: echo ${{ steps.my_action.outputs.some-value }}
```

Output:

```
test.yaml:8:23: property "my_action" is not defined in object type {} [expression]
  |
8 |       - run: echo ${{ steps.my_action.outputs.some_value }}
  |                       ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:15:23: property "some-value" is not defined in object type {some_value: string} [expression]
   |
15 |       - run: echo ${{ steps.my_action.outputs.some-value }}
   |                       ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

The 'My action with output' action defines one output `some_value`. The property is typed at `steps.my_action.outputs` object
so that actionlint can check incorrect property accesses like a typo in output name.

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
   |
19 |       - run: echo '${{ matrix.platform }}'
   |                        ^~~~~~~~~~~~~~~
test.yaml:21:24: property "dev" is not defined in object type {name: string; optional: bool} [expression]
   |
21 |       - run: echo '${{ matrix.package.dev }}'
   |                        ^~~~~~~~~~~~~~~~~~
test.yaml:34:24: property "os" is not defined in object type {} [expression]
   |
34 |       - run: echo '${{ matrix.os }}'
   |                        ^~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyNUstuwyAQvOcr5lCJRIotpUpUCalfUvVAbJzQ2oBYSFql+fdCHOo8XLUntMsMOzuD0Rw20HbyZtbEJ4CX5NMJkHfCy81nXwGd8E595AowxPES1kH7ULQi8ebYK12bPZ3r1x+sNrWM6MVyjsVqaFtRvYuNHN4ECmjRRSxrjGEX/TjPemW0aDm8C3KMshbuN0ojWho4SldtqG/nnjQuVlcvaNtxPJWrcnlqu6CpMNGzh8PhbEhpCMfj2TFpKT9aJDCHrLYG7AJuozeNcV0ksb+gvT1lLXf36K8LnT0zBXKri92h0prYSUfqZo/TxRgjp4QRacn5SMI0W/08Asp3ETgb3Tm6nCVBEcjKSjVK1oMW1fTjK6O9UJqmWbTt5mAxIDbrU0j/7pFfh3X1Sf+fVG/gN6xO5N4=)

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
      # ERROR: Outputs in other job is not accessible
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
   |
16 |       - run: echo '${{ needs.prepare.outputs.prepared }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:26:24: property "foo" is not defined in object type {installed: string} [expression]
   |
26 |       - run: echo '${{ needs.install.outputs.foo }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:28:24: property "some_job" is not defined in object type {install: {outputs: {installed: string}; result: string}; prepare: {outputs: {prepared: string}; result: string}} [expression]
   |
28 |       - run: echo '${{ needs.some_job }}'
   |                        ^~~~~~~~~~~~~~
test.yaml:33:24: property "build" is not defined in object type {} [expression]
   |
33 |       - run: echo '${{ needs.build.outputs.built }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJylUs1uwjAMvvMUPiD1QvMAeRU0oZZ4S6cujmpHHBDvTh1CKjY0EBwiJfHn78cJBQsxsV99U892BTAElm4cdQtASWISvhxqDZ2FxhjT5OspBW5ppkl9CpLasRNkySUWjLW5VaQF3HuCpjAB0w+KH8KXcsUJYzfhXelSe19ZWXTdSv+BrY9HCIiOTVE2xdD17OB00s4+DaO7KGW8hW0Jt7ma/rgXSPvk7TRZfYkCh0E8LN6Lk+q9PuBsHrrg4OmY/wzot8gn0eMmtbyb/1xBknic7OtzWIjzRKqXPGXVOAPMHOsV)

Job dependencies can be defined at [`needs:`][needs-doc]. A job runs after all jobs defined in `needs:` are done.
Outputs from the jobs can be accessed only from jobs following them via [`needs` context][needs-context-doc].

actionlint defines type of `needs` variable contextually looking at each job's `outputs:` section and `needs:` section.

<a name="check-shellcheck-integ"></a>
## [shellcheck][] integration for `run:`

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
  |
6 |       - run: echo $FOO
  |         ^~~~
test.yaml:14:9: shellcheck reported issue in this script: SC2086:info:1:6: Double quote to prevent globbing and word splitting [shellcheck]
   |
14 |       - run: echo $FOO
   |         ^~~~
```

[shellcheck][] is a famous linter for ShellScript. actionlint runs shellcheck for scripts at `run:` step in a workflow.
For installing shellcheck, see [the official installation document][shellcheck-install].

actionlint detects which shell is used to run the scripts following [the documentation][shell-doc]. On Linux or macOS,
the default shell is `bash` and on Windows it is `pwsh`. Shell can be configured by `shell:` configuration at a workflow
level or job level. Each step can configure shell to run scripts by `shell:`.

In the above example output, `SC2086:info:1:6:` means that shellcheck reported SC2086 rule violation and the location is at
line 1, column 6. Note that the location is relative to the script of the `run:` section.

actionlint remembers the default shell and checks what OS the job runs on. Only when the shell is `bash` or `sh`, actionlint
applies shellcheck to scripts.

By default, actionlint checks if `shellcheck` command exists in your system and uses it when it is found. The `-shellcheck`
option on running `actionlint` command specifies the executable path of shellcheck. Setting empty string by `shellcheck=`
disables shellcheck integration explicitly.

Since both `${{ }}` expression syntax and ShellScript's variable access `$FOO` use `$`, remaining `${{ }}` confuses shellcheck.
To avoid it, actionlint replaces `${{ }}` with underscores. For example `echo '${{ matrix.os }}'` is replaced with
`echo '________________'`.

Some shellcheck rules conflict with the `${{ }}` expression syntax. To avoid errors due to the syntax, [SC1091][sc1091],
[SC2050][sc2050], [SC2194][sc2194] are disabled.

When what shell is used cannot be determined statically, actionlint assumes `shell: bash` optimistically. For example,

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
steps:
  - name: Show file content
    run: Get-Content -Path xxx\yyy.txt
    if: ${{ matrix.os == 'windows-latest' }}
```

The 'Show file content' script is only run by `pwsh` due to `matrix.os == 'windows-latest'` guard. However actionlint does not
know that. It checks the script with shellcheck and it'd probably cause a false-positive (due to file separator). This kind of
false positives can be avoided by showing the shell name explicitly. It is also better in terms of maintenance of the workflow.

```yaml
- name: Show file content
  run: Get-Content -Path xxx\yyy.txt
  if: ${{ matrix.os == 'windows-latest' }}
  shell: pwsh
```

<a name="check-pyflakes-integ"></a>
## [pyflakes][] integration for `run:`

Example input:

```yaml
on: push
jobs:
  linux:
    runs-on: ubuntu-latest
    steps:
      # Yay! No error
      - run: print('${{ runner.os }}')
        shell: python
      # ERROR: Undefined variable
      - run: print(hello)
        shell: python
  linux2:
    runs-on: ubuntu-latest
    defaults:
      run:
        # Run script with Python by default
        shell: python
    steps:
      - run: |
          import sys
          for sys in ['system1', 'system2']:
            print(sys)
      - run: |
          from time import sleep
          print(100)
```

Output:

```
test.yaml:10:9: pyflakes reported issue in this script: 1:7 undefined name 'hello' [pyflakes]
   |
10 |       - run: print(hello)
   |         ^~~~
test.yaml:19:9: pyflakes reported issue in this script: 2:5 import 'sys' from line 1 shadowed by loop variable [pyflakes]
   |
19 |       - run: |
   |         ^~~~
test.yaml:23:9: pyflakes reported issue in this script: 1:1 'time.sleep' imported but unused [pyflakes]
   |
23 |       - run: |
   |         ^~~~
```

Python script can be written in `run:` when `shell: python` is configured.

[pyflakes][] is a famous linter for Python. It is suitable for linting small code like scripts at `run:` since it focuses
on finding mistakes (not a code style issue) and tries to make false positives as minimal as possible. Install pyflakes
by `pip install pyflakes`.

actionlint runs pyflakes for scripts at `run:` steps in a workflow and reports errors found by pyflakes. actionlint detects
Python scripts in a workflow by checking `shell: python` at each step and `defaults:` configurations at workflows and jobs.

By default, actionlint checks if `pyflakes` command exists in your system and uses it when found. The `-pyflakes` option
of `actionlint` command allows to specify the executable path of pyflakes. Setting empty string by `pyflakes=` disables
pyflakes integration explicitly.

Since both `${{ }}` expression syntax is invalid as Python, remaining `${{ }}` might confuse pyflakes. To avoid it,
actionlint replaces `${{ }}` with underscores. For example `print('${{ matrix.os }}')` is replaced with
`print('________________')`.

<a name="untrusted-inputs"></a>
## Script injection by potentially untrusted inputs

Example input:

```yaml
name: Test
on: pull_request

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Print pull request title
        # ERROR: Using the potentially untrusted input can cause script injection
        run: echo '${{ github.event.pull_request.title }}'
      - uses: actions/stale@v4
        with:
          repo-token: ${{ secrets.TOKEN }}
          # This is OK because action input is not evaluated by shell
          stale-pr-message: ${{ github.event.pull_request.title }} was closed
      - uses: actions/github-script@v4
        with:
          # ERROR: Using the potentially untrusted input can cause script injection
          script: console.log('${{ github.event.head_commit.author.name }}')
      - name: Get comments
        # ERROR: Accessing to untrusted inputs via `.*` object filter; bodies of comment, review, and review_comment
        run: echo '${{ toJSON(github.event.*.body) }}'
```

Output:

```
test.yaml:10:24: "github.event.pull_request.title" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details [expression]
   |
10 |         run: echo '${{ github.event.pull_request.title }}'
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:19:36: "github.event.head_commit.author.name" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details [expression]
   |
19 |           script: console.log('${{ github.event.head_commit.author.name }}')
   |                                    ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:22:31: object filter extracts potentially untrusted properties "github.event.comment.body", "github.event.discussion.body", "github.event.issue.body", "github.event.pull_request.body", "github.event.review.body", "github.event.review_comment.body". avoid using the value directly in inline scripts. instead, pass the value through an environment variable. see https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions for more details [expression]
   |
22 |         run: echo '${{ toJSON(github.event.*.body) }}'
   |                               ^~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyFkUFLAzEQhe/9FXMQ2gqJF085eRFBoRXsvWSzQ3drNrNmJi1S+t9NdkupSvUUknnzvcdLsB0aWCHLhIKBPnm/jviRysNkSxWbCYDkWzkBYgqsijBVKUhS3pbZMGLBnkcVgIIwgF9jG2SgwokK0orHk2wAGkDXEExvDgfYtNKkSuMOg+jLMHpYg+NxenZIjGzAOmkp8B2L9fiwuz+T9xllzrfshD0poXfMhsWK0UUU1qvly+Migy+kA0v1UXXIbDc4LvyfDfaWwXlirK+kHBmKXWx7+SvtqDDg8hZ51J42s98NNWjrtaOua0XbJA1FXXovNc1//MQTChRlXuNr7Qs9vy0Xs28Wt7qi+nNekF/IR69F)

Since `${{ }}` placeholders are evaluated and replaced directly by GitHub Actions runtime, you need to use them carefully in
inline scripts at `run:`. For example, if we have step as follows,

```yaml
- run: echo 'issue ${{github.event.issue.title}}'
```

an attacker can create a new issue with title `'; malicious_command ...`, and the inline script will run
`echo 'issue'; malicious_command ...` in your workflow. The remediation of such script injection is passing potentially untrusted
inputs via environment variables. See [the official document][security-doc] for more details.

```yaml
- run: echo "issue ${TITLE}"
  env:
    TITLE: ${{github.event.issue.title}}
```

actionlint recognizes the following inputs as potentially untrusted and checks your inline scripts at `run:`. When they are used
directly in a script, actionlint will report it as error.

- `github.event.issue.title`
- `github.event.issue.body`
- `github.event.pull_request.title`
- `github.event.pull_request.body`
- `github.event.comment.body`
- `github.event.review.body`
- `github.event.review_comment.body`
- `github.event.pages.*.page_name`
- `github.event.commits.*.message`
- `github.event.head_commit.message`
- `github.event.head_commit.author.email`
- `github.event.head_commit.author.name`
- `github.event.commits.*.author.email`
- `github.event.commits.*.author.name`
- `github.event.pull_request.head.ref`
- `github.event.pull_request.head.label`
- `github.event.pull_request.head.repo.default_branch`
- `github.head_ref`

Not only direct accesses to the untrusted properties, actionlint also detects those properties indirectly accessed via
[object filter syntax][object-filter-syntax]. For example, `github.event.*.body` collects all `body` properties in child objects
of `github.event` as array. Those properties include untrusted inputs like `github.event.comment.body`,
`github.event.pull_request.body`, ...

```sh
# Echo list of github.event.comment.body, github.event.pull_request.body, ...
echo '${{ toJSON(github.event.*.body) }}'
```

Instead, you should store the JSON string in environment variable:

```sh
- run: echo "${BODIES}"
  env:
    BODIES: '${{ toJSON(github.event.*.body) }}'
```

At last, the popular action [actions/github-script][github-script] has the same issue in its `script` input. actionlint also
checks the input.

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
  |
8 |   install:
  |   ^~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyljjEOxCAMBPu8YjsqPsBXTikgsZREyCBs/z+Bo0mdzvJ4Z104oJocy1WShAWojWps1EeAiXYJ+CU7876OVTMWX56UJWM1n6OS6ECiVOUfBHy/DKDtKHBT6h52smjM+e2f/EPD1PaG8ezbP+kH/5C6G78nW+Q=)

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
  |
4 |     needs: [bar, BAR]
  |                  ^~~~
test.yaml:8:3: job "bar" needs job "unknown" which does not exist in this workflow [job-needs]
  |
8 |   bar:
  |   ^~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyljD0OgiEQRHtOMR2NXIDO7wi2xgJ0v+BPdgnLxusLWFlbTfJm5glHVNPiHpI1OmAXmQEw0U0jzjm1A7bj6bJoM9Yg42TZuFt4pU7aV6Wdqn6/QJjLCLoWgS93P/AQ/ZqNnyxv/k/8AXoNOHs=)

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
  |
6 |         node: [10, 12, 14, 14]
  |                            ^~~
test.yaml:9:19: value "13" in "exclude" does not exist in matrix "node" combinations. possible values are "10", "12", "14", "14" [matrix]
  |
9 |           - node: 13
  |                   ^~
test.yaml:12:13: "platform" in "exclude" section does not exist in matrix. available matrix configurations are "node", "os" [matrix]
   |
12 |             platform: ubuntu-latest
   |             ^~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJxtkMEOgjAQRO9+xRw80gbUU3/FeACsooEu6bYJhvDvtqGgJBw2zezOvOyWjELvuTm8qWJ1AJxmF1+AnS2dfn5mBXSls69hUYChu1a4FnmG4hTqEuu2jonD0FfeOC/aMmKzgKiJk/o59VC3PrDWBiASvTj/NWfmBrkXyTeRPhgfZLu9oPWGBYUfOI5jOk8SY5rS/brnZSkRzQq6bghSyi9UwlNB)

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
    # ERROR: Incorrect filter. 'branches' is correct
    branch: foo
  issues:
    # ERROR: Incorrect type. 'opened' is correct
    types: created
  # ERROR: Unknown event name
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
  |
4 |     branch: foo
  |     ^~~~~~~
test.yaml:7:12: invalid activity type "created" for "issues" Webhook event. available types are "opened", "edited", "deleted", "transferred", ... [events]
  |
7 |     types: created
  |            ^~~~~~~
test.yaml:9:3: unknown Webhook event "pullreq". see https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#webhook-events for list of all Webhook event names [events]
  |
8 |   pullreq:
  |   ^~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJwtjMERAyEMA/9UoQagALoB4gzJMDbB9iPd5+DysjS7snAOwHTt+wJ1FW494yly9Zeqk97EvvOKaIuK0eOMxlj0ySG8pR7JSO2Wl7NG4QyvzuZxlM0OUqP5fwnEbWZQ64KU0g/W7Cnh)

At `on:`, Webhook events can be specified to trigger the workflow. [Webhook event documentation][webhook-doc] defines
which Webhook events are available and what types can be specified at `types:` for each event.

actionlint validates the Webhook configurations:

- unknown Webhook event name
- unknown type for Webhook event
- invalid filter names

The table of available Webhooks and their types are defined in [`all_webhooks.go`](../all_webhooks.go). It is generated
by [a script][generate-webhook-events] and kept to the latest by CI workflow triggered weekly.

<a name="check-workflow-dispatch-events"></a>
## Workflow dispatch event validation

Example input:

```yaml
on:
  workflow_dispatch:
    inputs:
      # Unknown input type
      id:
        type: number
      # ERROR: No options for 'choice' input type
      kind:
        type: choice
      name:
        type: choice
        options:
          - Tama
          - Mike
        # ERROR: Default value is not in options
        default: Chobi
      message:
        type: string
      verbose:
        type: boolean
        # ERROR: Boolean value must be 'true' or 'false'
        default: yes

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # ERROR: Undefined input
      - run: echo "${{ github.event.inputs.massage }}"
      # ERROR: Bool value is not available for object key
      - run: echo "${{ env[github.event.inputs.verbose] }}"
```

Output:

```
test.yaml:6:15: input type of workflow_dispatch event must be one of "string", "boolean", "choice", "environment" but got "number" [syntax-check]
  |
6 |         type: number
  |               ^~~~~~
test.yaml:8:7: input type of "kind" is "choice" but "options" is not set [events]
  |
8 |       kind:
  |       ^~~~~
test.yaml:16:18: default value "Chobi" of "name" input is not included in its options "\"Tama\", \"Mike\"" [events]
   |
16 |         default: Chobi
   |                  ^~~~~
test.yaml:22:18: type of "verbose" input is "boolean". its default value "yes" must be "true" or "false" [events]
   |
22 |         default: yes
   |                  ^~~
test.yaml:29:24: property "massage" is not defined in object type {message: string; verbose: bool; id: any; kind: string; name: string} [expression]
   |
29 |       - run: echo "${{ github.event.inputs.massage }}"
   |                        ^~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:31:28: property access of object must be type of string but got "bool" [expression]
   |
31 |       - run: echo "${{ env[github.event.inputs.verbose] }}"
   |                            ^~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJx9kDFuwzAMRXefggi62gfQmrlbt6IoJJuxVVukIFIOgiB3b2yrDQq33cjH/yl+MZkK4MxpPE18fu+8RKvtsEAATzGrbPW9674qAL1ENEA5OEwFjp52gnZg32KBZAP+KwDgqJ5JHiqAGl5ssD/Asx8flg5PNk9q4Diw8wUHFLH97jXR5KkvcMbkWHYaxzyhpf3+C0pVfbBbr1MU3Zwpk9RMBrLLpLme7DJbR6IYv7PUi9IA3hPD4el6hd7rkF2DM5I22083wa53w+12+MuGNL/+Zi1x3hbvJ3Q1hA8=)

[`workflow_dispatch`][workflow-dispatch-event] is an event to trigger a workflow manually. The event can have parameters called
'inputs'. Each input has its name, description, default value, and [input type][workflow-dispatch-input-type-announce].

actionlint checks several mistakes around `workflow_dispatch` configuration.

- Input type must be one of 'choice', 'string', 'boolean', 'environment'
- `options:` must be set for 'choice' input type
- The default value of 'choice' input must be included in options
- The default value of 'boolean' input must be `true` or `false`

In addition, `github.event.inputs` object is typed based on the input definitions. properties not defined in `inputs:` will cause
a type error thanks to a type checker.

For example,

```yaml
inputs:
  string_input:
    type: string
  choice_input:
    type: choice
    options: ['hello']
  bool_input:
    type: boolean
  env_input:
    type: environment
  no_type_input:
```

`github.event.inputs` is typed as follows from these definitions:

```
{
  "string_input": string;
  "choice_input": string;
  "bool_input": bool;
  "env_input": string;
  "no_type_input": any;
}
```

<a name="check-glob-pattern"></a>
## Glob filter pattern syntax validation

Example input:

```yaml
on:
  push:
    branches:
      # ^ is not available for branch name. This kind of mistake is usually caused by misunderstanding
      # that regular expression is available here
      - '^foo-'
    tags:
      # Invalid syntax. + cannot follow special character *
      - 'v*+'
      # Invalid character range 9-1
      - 'v[9-1]'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo ...
```

Output:

```
test.yaml:6:10: character '^' is invalid for branch and tag names. ref name cannot contain spaces, ~, ^, :, [, ?, *. see `man git-check-ref-format` for more details. note that regular expression is unavailable. note: filter pattern syntax is explained at https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet [glob]
  |
6 |       - '^foo-'
  |          ^~~~~~
test.yaml:9:12: invalid glob pattern. unexpected character '+' while checking special character + (one or more). the preceding character must not be special character. note: filter pattern syntax is explained at https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet [glob]
  |
9 |       - 'v*+'
  |            ^~
test.yaml:11:14: invalid glob pattern. unexpected character '1' while checking character range in []. start of range '9' (57) is larger than end of range '1' (49). note: filter pattern syntax is explained at https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet [glob]
   |
11 |       - 'v[9-1]'
   |              ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJxNjEEKAjEQBO95Rd8CygQ8mq+IQrJEg8jMsjPj+3Wzl5yaoooWzgFYXfu+QN0KL73pQQAhPp4iFAdbec3mezrHiW5XutxjCG+po7KmdtSbs5Jwhldnc/qU3Q2l1tbp819mtKULUko/10snvA==)

For filtering branches, tags and paths in Webhook events, [glob syntax][filter-pattern-doc] is available.
actionlint validates glob patterns `branches:`, `branches-ignore:`, `tags:`, `tags-ignore:`, `paths:`, `paths-ignore:` in a
workflow. It checks:

- syntax errors like missing closing brackets for character range `[..]`
- invalid usage like `?` following `*`, invalid character range `[9-1]`, ...
- invalid character usage for Git ref names (branch name, tag name)
  - ref name cannot start/end with `/`
  - ref name cannot contain `[`, `:`, `\`, ...

Most common mistake I have ever seen here is misunderstanding that regular expression is available for filtering. This rule
can catch the mistake so that users can notice their mistakes.

<a name="check-cron-syntax"></a>
## CRON syntax check at `schedule:`

Example input:

```yaml
on:
  schedule:
    # ERROR: Cron syntax is not correct
    - cron: '0 */3 * *'
    # ERROR: Interval of scheduled job is too small (job runs too frequently)
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
  |
4 |     - cron: '0 */3 * *'
  |             ^~
test.yaml:6:13: scheduled job runs too frequently. it runs once per 60 seconds. the shortest interval is once every 5 minutes [events]
  |
6 |     - cron: '* */3 * * *'
  |             ^~
```

[Playground](https://rhysd.github.io/actionlint#eJxVjEEKgDAMBO99xd6EQKvgrb/RGhAprTTN/zWKB2/LzuzWEh0gaedNM1sGPFKrJWKYQOMMAg3/nr7eiDvqKjbsLP09aFrEm6mrlq4+L8YeJJ1PeS07vM0ITntFCOEChKgjxA==)

To trigger a workflow in specific interval, [scheduled event][schedule-event-doc] can be defined in [POSIX CRON syntax][cron-syntax].

actionlint checks the CRON syntax and frequency of running a job. [The official document][schedule-event-doc] says:

> The shortest interval you can run scheduled workflows is once every 5 minutes.

When the job is run more frequently than once every 5 minutes, actionlint reports it as an error.

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
test.yaml:10:13: label "linux-latest" is unknown. available labels are "windows-latest", "windows-2022", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
   |
10 |           - linux-latest
   |             ^~~~~~~~~~~~
test.yaml:16:13: label "gpu" is unknown. available labels are "windows-latest", "windows-2022", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
   |
16 |           - gpu
   |             ^~~
test.yaml:23:14: label "macos-10.13" is unknown. available labels are "windows-latest", "windows-2022", "windows-2019", "windows-2016", "ubuntu-latest", ... [runner-label]
   |
23 |     runs-on: macos-10.13
   |              ^~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyFj8EKgzAQRO/5ijn0aEJtxUN+pfSQ2lQtNpFsAhbx36skUgKFnpbZnR3mWSMxBurY095IMsBr8tsEyDvldfuOCngp7/ppV4ALxmj31QBfPY0lPqgtJTsMvQnTr8OF9PDgnSWv70W0FZjq6pq5lHvVVbZpx8BSC+J2pTjMc6ooYjMsS+LQI+01+fYgoZvOQgjBEvFJ5mGRozyK8vw34wNI+VUQ)

GitHub Actions provides two kinds of job runners, [GitHub-hosted runner][gh-hosted-runner] and [self-hosted runner][self-hosted-runner].
Each runner has one or more labels. GitHub Actions runtime finds a proper runner based on label(s) specified at `runs-on:`
to run the job. So specifying proper labels at `runs-on:` is important.

actionlint checks proper label is used at `runs-on:` configuration. Even if an expression is used in the section like
`runs-on: ${{ matrix.foo }}`, actionlint parses the expression and resolves the possible values, then validates the values.

When you define some custom labels for your self-hosted runner, actionlint does not know the labels. Please set the label
names in [`actionlint.yaml` configuration file](config.md) to let actionlint know them.

In addition to checking label values, actionlint checks combinations of labels. `runs-on:` section can be an array which contains
multiple labels. In the case, a runner which has all the labels will be selected. However, those labels combinations can have
conflicts.

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: [ubuntu-latest, windows-latest]
    steps:
      - run: echo ...
```

Output:

```
test.yaml:4:30: label "windows-latest" conflicts with label "ubuntu-latest" defined at line:4,col:15. note: to run your job on each worker, use matrix [runner-label]
  |
4 |     runs-on: [ubuntu-latest, windows-latest]
  |                              ^~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJwti0EOgCAMBO+8Yh8gPICvGA+iJEhMS2wbvq+op81mZpgimklxlZNEB2gWHQtcRuL54bMlIzV/rgNO6Aft3OX/yyuL5iZfB/jRRuStMEIIN17iHww=)

In most cases this is a misunderstanding that a matrix combination can be specified at `runs-on:` directly. It should use
`matrix:` and expand it with `${{ }}` at `runs-on:` to run the workflow on multiple runners.

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
      # ERROR: local action must starts with './'
      - uses: .github/my-actions/do-something
```

Output:

```
test.yaml:7:15: specifying action "actions/checkout" in invalid format because ref is missing. available formats are "{owner}/{repo}@{ref}" or "{owner}/{repo}/{path}@{ref}" [action]
  |
7 |       - uses: actions/checkout
  |               ^~~~~~~~~~~~~~~~
test.yaml:9:15: specifying action "checkout@v2" in invalid format because owner is missing. available formats are "{owner}/{repo}@{ref}" or "{owner}/{repo}/{path}@{ref}" [action]
  |
9 |       - uses: checkout@v2
  |               ^~~~~~~~~~~
test.yaml:11:15: tag of Docker action should not be empty: "docker://image" [action]
   |
11 |       - uses: 'docker://image:'
   |               ^~~~~~~~~~~~~~~~~
test.yaml:13:15: specifying action ".github/my-actions/do-something" in invalid format because ref is missing. available formats are "{owner}/{repo}@{ref}" or "{owner}/{repo}/{path}@{ref}" [action]
   |
13 |       - uses: .github/my-actions/do-something
   |               ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJxdzTEOgzAMBdCdU3hjSi119NSrJKlFUkqMsF2pty8UsWTy139fsjSC1bUML0lKA4Cx2nEBNm8aZHdP3szDOx72JzVe9VwBBHBlJYjZqjTFXDjP4tbxVT8+907Gp+SZN0KsS5yYxs5vOFUrnvD6sHzDGX98DjoH)

Action needs to be specified in a format defined in [the document][action-uses-doc]. There are 3 types of actions:

- action hosted on GitHub: `owner/repo/path@ref`
- local action: `./path/to/my-action`
- Docker action: `docker://image:tag`

actionlint checks values at `uses:` sections follow one of these formats.

Note that actionlint does not report any error when a direcotyr for a local action does not exist in the repository because it is
a common case that the action is managed in a separate repository and the action directory is cloned at running the workflow.
(See [#25][issue-25] and [#40][issue-40] for more details).

<a name="check-local-action-inputs"></a>
## Local action inputs validation at `with:`

My action definition at `.github/actions/my-action/action.yaml`:

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
test.yaml:7:15: missing input "message" which is required by action "My action" defined at "./.github/actions/my-action". all required inputs are "message" [action]
  |
7 |       - uses: ./.github/actions/my-action
  |               ^~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:13:11: input "additions" is not defined in action "My action" defined at "./.github/actions/my-action". available inputs are "addition", "message", "name" [action]
   |
13 |           additions: foo, bar
   |           ^~~~~~~~~~
```

When a local action is run in `uses:` of `step:`, actionlint reads `action.yml` file in the local action directory and
validates inputs at `with:` in the workflow are correct. Missing required inputs and unexpected inputs can be detected.

<a name="check-popular-action-inputs"></a>
## Popular action inputs validation at `with:`

Example input:

```yaml
on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/cache@v2
        with:
          keys: |
            ${{ hashFiles('**/*.lock') }}
            ${{ hashFiles('**/*.cache') }}
          path: ./packages
      - run: make
```

Output:

```
test.yaml:7:15: missing input "key" which is required by action "actions/cache@v2". all required inputs are "key", "path" [action]
  |
7 |       - uses: actions/cache@v2
  |               ^~~~~~~~~~~~~~~~
test.yaml:9:11: input "keys" is not defined in action "actions/cache@v2". available inputs are "key", "path", "restore-keys", "upload-chunk-size" [action]
  |
9 |           keys: |
  |           ^~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyFj0EKwjAQRfc9xV8I1UJbcJmVK+8xDYOpqUlwEkVq725apYgbV8PMe/Dne6cQkpiiOPtOVAFEljhP4Jqc1D4LqUsupnqgmS1IIgd5W0CNJCwKpGPvnbSatOHDbf/BwL2PRq0bYPmR9efXBdiMIwyJOfYDy7asqrZqBq9tucM0/TWXyF81UI5F0wbSlk4s67u5mMKFLL8A+h9EEw==)

actionlint checks inputs of many popular actions such as `actions/checkout@v2`. It checks

- some input is required by the action but it not set at `with:`
- input set at `with:` is not defined in the action (this commonly occurs by typo)

this is done by checking `with:` section items with a small database collected at building `actionlint` binary. actionlint
can check popular actions without fetching any `action.yml` of the actions from remote so that it can run efficiently.

Note that it only supports the case of specifying major version like `actions/checkout@v2`. Fixing version of action like
`actions/checkout@v2.3.4` and using the HEAD of action like `actions/checkout@main` are not supported for now.

So far, actionlint supports more than 100 popular actions The data set is embedded at [`popular_actions.go`](../popular_actions.go)
and were automatically collected by [a script][generate-popular-actions]. If you want more checks for other actions, please
make a request [as an issue][issue-form].

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
      - run: echo 'hello'
        # OK: 'powershell' is only available on Windows
        shell: powershell
```

Output:

```
test.yaml:8:16: shell name "dash" is invalid. available names are "bash", "pwsh", "python", "sh" [shell-name]
  |
8 |         shell: dash
  |                ^~~~
test.yaml:11:16: shell name "powershell" is invalid on macOS or Linux. available names are "bash", "pwsh", "python", "sh" [shell-name]
   |
11 |         shell: powershell
   |                ^~~~~~~~~~
test.yaml:14:16: shell name "powershell" is invalid on macOS or Linux. available names are "bash", "pwsh", "python", "sh" [shell-name]
   |
14 |         shell: powershell
   |                ^~~~~~~~~~
test.yaml:20:16: shell name "fish" is invalid. available names are "bash", "pwsh", "python", "sh" [shell-name]
   |
20 |         shell: fish
   |                ^~~~
test.yaml:30:16: shell name "sh" is invalid on Windows. available names are "bash", "pwsh", "python", "cmd", "powershell" [shell-name]
   |
30 |         shell: sh
   |                ^~
```

[Playground](https://rhysd.github.io/actionlint#eJylkLsKwzAMRfd8hbZMhs7+GydWcIpqGcsihdJ/r52GUjz1sUm6R48rjhaSShjOPIkdAGiNem0BQNYohiugk8aihlxBKbskBZM8KQDTSAs4B4YxIBGPh1LBllvwrq74mE68Yd7jH3subu4s1ArLuwOPi1MqLxNtQT9zWY+rv7U7JswEt9O9KdsaPW/SHXRU/3mqhAdbk36k)

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
test.yaml:10:13: step ID "STEP_ID" duplicates. previously defined at line:7,col:13. step ID must be unique within a job. note that step ID is case insensitive [step-id]
   |
10 |         id: STEP_ID
   |             ^~~~~~~
test.yaml:12:3: key "test" is duplicate in "jobs" section. previously defined at line:3,col:3. note that key names are case insensitive [syntax-check]
   |
12 |   TEST:
   |   ^~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJzLz7NSKCgtzuDKyk8qtuJSUChJLS4B0QoKRaV5xbr5QPnSpNK8klLdnESQHFiquCS1oBiiSkFBF6TSSiE1OSNfQT0jNScnXx0qo6CQmWIFVhyfmYJNdVJlKqra4BDXgHhPF6BYiGtwCE3cAQCKgUNq)

Job IDs and step IDs in each jobs must be unique. IDs are compared in case insensitive. actionlint checks all job IDs
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
   |
10 |         password: pass
   |                   ^~~~
test.yaml:17:21: "password" section in "redis" service should be specified via secrets. do not put password value directly [credentials]
   |
17 |           password: pass
   |                     ^~~~
```

[Playground](https://rhysd.github.io/actionlint#eJx1kLEOwyAMRPd8hTemNDt/4xCroQKMMDT9/AJNUYd0wrp357PgoCEW2acHr6IngEyS2wuQSpCZKy9rCbnMDhvryHDIaAOljxPAeryTBkUv9NHRzbBf+KiGpRN12kyijUK26OSbBChCKaCv8TYNOaLIwWnTfepyxU9raGTrNvuz6Dyiq0O8rPxbel2bKY7w3P5FA5mdQe3kHKs3Uktdww==)

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
  |
6 |       FOO=BAR: foo
  |       ^~~~~~~~
test.yaml:7:7: environment variable name "foo bar" is invalid. '&', '=' and spaces should not be contained [env-var]
  |
7 |       FOO BAR: foo
  |       ^~~
```

[Playground](https://rhysd.github.io/actionlint#eJzLz7NSKCgtzuDKyk8qtuJSUChJLS4B0QoKRaV5xbr5QPnSpNK8klLdnESQHFgqNa8MokZBwc3f39bJMchKIS0/HyGkgCJUXJJaUAzToAsy2EohNTkjX0E9IzUnJ18dAPhYJMc=)

`=` must not be included in environment variable names. And `&` and spaces should not be included in them. In almost all
cases they are mistakes and they may cause some issues on using them in shell since they have special meaning in shell syntax.

actionlint checks environment variable names are correct in `env:` configuration.

<a name="permissions"></a>
## Permissions

Example input:

```yaml
on: push

# ERROR: Available values for whole permissions are "write-all", "read-all" or "none"
permissions: write

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      # ERROR: "checks" is correct scope name
      check: write
      # ERROR: Available values are "read", "write" or "none"
      issues: readable
    steps:
      - run: echo hello
```

Output:

```
test.yaml:4:14: "write" is invalid for permission for all the scopes. available values are "read-all" and "write-all" [permissions]
  |
4 | permissions: write
  |              ^~~~~
test.yaml:11:7: unknown permission scope "check". all available permission scopes are "actions", "checks", "contents", "deployments", "id-token", "issues", "metadata", "packages", "pull-requests", "repository-projects", "security-events", "statuses" [permissions]
   |
11 |       check: write
   |       ^~~~~~
test.yaml:13:15: "readable" is invalid for permission of scope "issues". available values are "read", "write" or "none" [permissions]
   |
13 |       issues: readable
   |               ^~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJxNjd0NwyAMhN89xS3AAmwDxBK0FCOMlfUDiVr16aTv/qR5dNNM1Hl8imqRph7nKJOJXhLVEzBZ51ZgWFMnq2TR2jRXw/Zu63/gBkDKnN7ftQethPF6GByOEOuDdXL/ldw+8eCUBZlrlQvntjLp)

Permissions of `GITHUB_TOKEN` token can be configured at workflow-level or job-level by [`permissions:` section][perm-config-doc].
Each permission scopes have their access levels. The default levels are described in [the document][permissions-doc].

actionlint checks permission scopes and access levels in a workflow are correct.

<a name="check-reusable-workflows"></a>
## Reusable workflows

[Reusable workflows][reusable-workflow-doc] is a feature to call a workflow from another workflow.

As of now, the integrity between a workflow caller and a workflow callee is not checked. actionlint does only the syntax check
and very basic validations.

### Check input definitions of `workflow_call` event

Example input:

```yaml
on:
  workflow_call:
    inputs:
      scheme:
        description: Scheme of URL
        # OK: Type is string
        default: https
        type: string
      host:
        # ERROR: Missing description
        default: example.com
        type: string
      port:
        description: Port of URL
        # ERROR: Type is number but default value is string
        default: ':1234'
        type: number
      query:
        description: Scheme of URL
        # ERROR: Type must be one of number, string, boolean
        type: object
jobs:
  do:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ inputs.scheme }}://${{ inputs.host }}:${{ inputs.port }}"
  nested:
    # ERROR: Nested workflow call is not allowed
    uses: onwer/repo/w.yml@main
```

Output:

```
test.yaml:9:7: "description" is missing at "host" input of workflow_call event [syntax-check]
  |
9 |       host:
  |       ^~~~~
test.yaml:16:18: input of workflow_call event "port" is typed as number but its default value ":1234" cannot be parsed as a float number: strconv.ParseFloat: parsing ":1234": invalid syntax [events]
   |
16 |         default: ':1234'
   |                  ^~~~~~~
test.yaml:21:15: invalid value "object" for input type of workflow_call event. it must be one of "boolean", "number", or "string" [syntax-check]
   |
21 |         type: object
   |               ^~~~~~
test.yaml:29:11: reusable workflow cannot be nested. but this workflow hooks "workflow_call" event at line:2,col:3 [workflow-call]
   |
29 |     uses: onwer/repo/w.yml@main
   |           ^~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJyVkcFuwjAMhu88hYUmcSposFOegcPExBm1waVlaZzFjqCqeHeatkMVCCRuyff7t+zfZNUE4ET+Nzd02unUmAgASuuCcP8GYF1ghf8/gD2y9qWTkqyCn04EymG7WY9K8jQYUVCIOL5hqR0qYPGlPQywIBb16MNzWjmDc03VK7cjL08G+26lp2PN1Ody9TW7a21DlaEf4F9AX7+1dN+EsiNqmRwp6wLcU9/DB8tJ9IYsWAmJSQVZOokF3S3sJFa2++uCYPrRNMMt5v0R4HJRi8UIx/giHKGYSYumVxXQiwY=)

Unlike inputs of action, inputs of workflow must specify their types. actionlint validates input types and checks the default
values are correctly typed. For more details, see [the official document][create-reusable-workflow-doc].

In addition, nested workflow call is not allowed. actionlint checks reusable workflow is not called when the workflow hooks
`workflow_call` event.

### Check reusable workflow call syntax

Example input:

```yaml
on: push
jobs:
  job1:
    uses: owner/repo/path/to/workflow.yml@v1
    # ERROR: 'runs-on' is not available on calling reusable workflow
    runs-on: ubuntu-latest
  job2:
    # ERROR: Local file path with ref is not available
    uses: ./.github/workflows/ci.yml@main
  job3:
    # ERROR: 'with' is only available on calling reusable workflow
    with:
      foo: bar
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
```

Output:

```
test.yaml:6:5: when a reusable workflow is called with "uses", "runs-on" is not available. only following keys are allowed: "name", "uses", "with", "secrets", "needs", "if", and "permissions" in job "job1" [syntax-check]
  |
6 |     runs-on: ubuntu-latest
  |     ^~~~~~~~
test.yaml:9:11: reusable workflow call "./.github/workflows/ci.yml@main" at "uses" is not following the format "owner/repo/path/to/workflow.yml@ref" nor "./path/to/workflow.yml". see https://docs.github.com/en/actions/learn-github-actions/reusing-workflows for more details [workflow-call]
  |
9 |     uses: ./.github/workflows/ci.yml@main
  |           ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
test.yaml:12:5: "with" is only available for a reusable workflow call with "uses" but "uses" is not found in job "job3" [syntax-check]
   |
12 |     with:
   |     ^~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJx9jTEOwjAMRfeewhdoo8KWiasklUsKaRzFNlFvT9sIxMT0Pbz3TMlCVg7dgzzbDmDf8VgAZWQLVBMWUzCTyU6CETKVynOOVIdtjbfXeMJFE/e0x9RrEu2jE2RpvctvbzDDfZGg/pthMy1nanVLasa1GXUH2wUwE1nwrvz/BsCCmT9Sf5AWcAoEAWOkN6pqSm4=)

When calling an external workflow, [only specific keys are available][reusable-workflow-call-keys] at job configuration.
For example, `secrets:` is not available when running steps as normal job. And `runs-on:` is not available when calling
a reusable workflow since the called workflow determines which OS is used. actionlint checks such keys are used correctly
to call a reusable workflow or to run steps as normal job.

And the workflow syntax at `uses:` must follow the format `owner/repo/path/to/workflow.yml@ref` as described in
[the official document][create-reusable-workflow-doc]. actionlint checks if the value follows the format.

### Check types of `inputs.*` and `secrets.*` in reusable workflow

Example input:

```yaml
on:
  workflow_call:
    inputs:
      url:
        description: 'your URL'
        type: string
      lucky_number:
        description: 'your lucky number'
        type: number
    secrets:
      credential:
        description: 'your credential'

jobs:
  test:
    runs-on: ubuntu-20.04
    steps:
      - name: Send data
        # ERROR: uri is typo of url
        run: curl ${{ inputs.uri }} -d ${{ inputs.lucky_number }}
        env:
          # ERROR: credentials is typo of credential
          TOKEN: ${{ secrets.credentials }}
```

Output:

```
test.yaml:20:23: property "uri" is not defined in object type {url: string; lucky_number: number} [expression]
   |
20 |         run: curl ${{ inputs.uri }} -d ${{ inputs.lucky_number }}
   |                       ^~~~~~~~~~
test.yaml:23:22: property "credentials" is not defined in object type {credential: string} [expression]
   |
23 |           TOKEN: ${{ secrets.credentials }}
   |                      ^~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJx9UD1PwzAQ3fMr3oCUKVGFmLwzgUAqMFeOfSDT9BzZZ6qo6n/HjUNSMXS7e/c+7s6zqoCjD/vP3h93Rvf9BQAcD0liqYEU+r8SsBRNcIM4zwr16FPAx/a5XuYyDqQQJTj+msE+mf2443ToKNw0mogoxP+OBZ3ASCbQul5uLLE4fXvLlVZX1bfvJr1QlKIKiWNzYacusaTmftNuHkqc0LCENWB9yOu8EVtYLXqJzAYKJv8Kd6fT/ME2BYfzGY29Bq//kaeLA/HPegHw/vr0+KIm4Xxxu94Qs/AXqVaEog==)

Inputs of reusable workflow call are set to `inputs.*` properties following the definitions at `on.workflow_call.inputs`.
And in a job of a reusable workflow, `secrets.*` are passed from caller of the workflow so it is set following the definitions at
`on.workflow_call.secrets`. See [the official document][create-reusable-workflow-doc] for more details.

actionlint contextually defines types of `inputs` and `secrets` contexts looking at `workflow_call` event. Keys of `inputs` only
allow keys at `on.workflow_call.inputs` and their values are typed based on `on.workflow_call.inputs.<input_name>.type`. Type of
`secrets` is also strictly typed following `on.workflow_call.secrets`.

### Check outputs of reusable workflow

Example input:

```yaml
on:
  workflow_call:
    outputs:
      image-version:
        description: "Docker image version"
        # ERROR: 'imagetag' does not exist (typo of 'image_tag')
        value: ${{ jobs.gen-image-version.outputs.imagetag }}
jobs:
  gen-image-version:
    runs-on: ubuntu-latest
    outputs:
      image_tag: "${{ steps.get_tag.outputs.tag }}"
    steps:
      - run: ./output_image_tag.sh
        id: get_tag
```

Output:

```
test.yaml:6:20: property "imagetag" is not defined in object type {image_tag: string} [expression]
  |
6 |         value: ${{ jobs.gen-image-version.outputs.imagetag }}
  |                    ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJx1j0EOgyAQRfc9xcR0C92z7j0M6tRSKRgYdGG8ex1Rkqbpkp/Hf3+8UxeA2YfhYf1ct9paDgB8ojFRzA8A89Y9iglDNN6dIUCHsQ1mJA6huvt2wJBZONiqsJO2CRVclwVevomyRye+auXhlHtKuod1vTDKvh86jwjJRcHy1CRHSVhNGOnvBfXWug3lDZFw5BHEWVFnax69E+d3wSoF8pbJutTJ+Cwnmk7B0fgB2ORuYw==)

Outputs of a reusable workflow can be defined at `on.workflow_call.outputs` as described in [the document][reusable-workflow-outputs].
The `jobs` context is available to define an output value to refer outputs of jobs in the workflow. actionlint checks
the context is used correctly.


<a name="job-id-naming-convention"></a>
## Job ID naming convention

Example input:

```yaml
on: push
jobs:
  # ERROR: '.' cannot be contained in job ID
  foo-v1.2.3:
    runs-on: ubuntu-latest
    steps:
      - run: 'job ID with version'
  # ERROR: Job ID cannot start with '-'
  -hello-world-:
    runs-on: ubuntu-latest
    steps:
      - run: 'hello'
```

Output:

```
test.yaml:4:3: invalid job ID "foo-v1.2.3". job ID must start with a letter or _ and contain only alphanumeric characters, -, or _ [job-needs]
  |
4 |   foo-v1.2.3:
  |   ^~~~~~~~~~~
test.yaml:9:3: invalid job ID "-hello-world-". job ID must start with a letter or _ and contain only alphanumeric characters, -, or _ [job-needs]
  |
9 |   -hello-world-:
  |   ^~~~~~~~~~~~~~
```

[Playground](https://rhysd.github.io/actionlint#eJydzDEOhDAMRNGeU0yXykhAl5qGY4DIKqAoRrEN1yfADbYczdPn7HGYxGbnRXwD/Jjp7Nq+HZ4FFMtCXJUtltUozRpE30s0HPIpgB7p4WoH04hr04gzFNk4u0oohpSYLi5ppf/Kb8HdY9w0wA==)

Job ID must start with a letter or `_` and contain only alphanumeric characters, `-` or `_`. actionlint checks the naming
convention and reports invalid IDs as error.

---

[Installation](install.md) | [Usage](usage.md) | [Configuration](config.md) | [Go API](api.md) | [References](reference.md)

[yamllint]: https://github.com/adrienverge/yamllint
[issue-form]: https://github.com/rhysd/actionlint/issues/new
[syntax-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions
[filter-pattern-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
[shellcheck]: https://github.com/koalaman/shellcheck
[shellcheck-install]: https://github.com/koalaman/shellcheck#installing
[sc1091]: https://github.com/koalaman/shellcheck/wiki/SC1091
[sc2050]: https://github.com/koalaman/shellcheck/wiki/SC2050
[sc2194]: https://github.com/koalaman/shellcheck/wiki/SC2194
[pyflakes]: https://github.com/PyCQA/pyflakes
[expr-doc]: https://docs.github.com/en/actions/learn-github-actions/expressions
[contexts-doc]: https://docs.github.com/en/actions/learn-github-actions/contexts
[funcs-doc]: https://docs.github.com/en/actions/learn-github-actions/expressions#functions
[needs-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idneeds
[needs-context-doc]: https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
[shell-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell
[matrix-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
[webhook-doc]: https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#webhook-events
[schedule-event-doc]: https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#scheduled-events
[cron-syntax]: https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html#tag_20_25_07
[gh-hosted-runner]: https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners
[self-hosted-runner]: https://docs.github.com/en/actions/hosting-your-own-runners/about-self-hosted-runners
[action-uses-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
[credentials-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
[actions-cache]: https://github.com/actions/cache
[permissions-doc]: https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token
[perm-config-doc]: https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#permissions
[generate-webhook-events]: https://github.com/rhysd/actionlint/tree/main/scripts/generate-webhook-events
[generate-popular-actions]: https://github.com/rhysd/actionlint/tree/main/scripts/generate-popular-actions
[issue-25]: https://github.com/rhysd/actionlint/issues/25
[issue-40]: https://github.com/rhysd/actionlint/issues/40
[security-doc]: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
[reusable-workflow-doc]: https://docs.github.com/en/actions/learn-github-actions/reusing-workflows
[create-reusable-workflow-doc]: https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#creating-a-reusable-workflow
[reusable-workflow-call-keys]: https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#supported-keywords-for-jobs-that-call-a-reusable-workflow
[object-filter-syntax]: https://docs.github.com/en/actions/learn-github-actions/expressions#object-filters
[github-script]: https://github.com/actions/github-script
[workflow-dispatch-event]: https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#workflow_dispatch
[workflow-dispatch-input-type-announce]: https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/
[reusable-workflow-outputs]: https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow
