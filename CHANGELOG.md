<a name="v1.6.10"></a>
# [v1.6.10](https://github.com/rhysd/actionlint/releases/tag/v1.6.10) - 11 Mar 2022

- Support outputs in reusable workflow call. See [the official document](https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow) for the usage of the outputs syntax. (#119, #121)
  Example of reusable workflow definition:
  ```yaml
  on:
    workflow_call:
      outputs:
        some_output:
          description: "Some awesome output"
          value: 'result value of workflow call'
  jobs:
    job:
      runs-on: ubuntu-latest
      steps:
        ...
  ```
  Example of reusable workflow usage:
  ```yaml
  jobs:
    job1:
      uses: ./.github/workflows/some_workflow.yml
    job2:
      runs-on: ubuntu-latest
      needs: job1
      steps:
        - run: echo ${{ needs.job1.outputs.some_output }}
  ```
- Support checking `jobs` context, which is only available in `on.workflow_call.outputs.<name>.value`. Outputs of jobs can be referred via the context. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-outputs-of-reusable-workflow) for more details.
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
- Add new major releases in `actions/*` actions including `actions/checkout@v3`, `actions/setup-go@v3`, `actions/setup-python@v3`, ...
- Check job IDs. They must start with a letter or `_` and contain only alphanumeric characters, `-` or `_`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#job-id-naming-convention) for more details. (#80)
  ```yaml
  on: push
  jobs:
    # ERROR: '.' cannot be contained in job ID
    foo-v1.2.3:
      runs-on: ubuntu-latest
      steps:
        - run: 'job ID with version'
  ```
- Fix `windows-latest` now means `windows-2022` runner. See [virtual-environments#4856](https://github.com/actions/virtual-environments/issues/4856) for the details. (#120)
- Update [the playground](https://rhysd.github.io/actionlint/) dependencies to the latest.
- Update Go module dependencies

[Changes][v1.6.10]


<a name="v1.6.9"></a>
# [v1.6.9](https://github.com/rhysd/actionlint/releases/tag/v1.6.9) - 24 Feb 2022

- Support [`runner.arch` context value](https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context). (thanks @shogo82148, #101)
  ```yaml
  steps:
    - run: ./do_something_64bit.sh
      if: ${{ runner.arch == 'x64' }}
  ```
- Support [calling reusable workflows in local directories](https://docs.github.com/en/actions/using-workflows/reusing-workflows#calling-a-reusable-workflow). (thanks @jsok, #107)
  ```yaml
  jobs:
    call-workflow-in-local-repo:
      uses: ./.github/workflows/useful_workflow.yml
  ```
- Add [a document](https://github.com/rhysd/actionlint/blob/main/docs/install.md#asdf) to install actionlint via [asdf](https://asdf-vm.com/) version manager. (thanks @crazy-matt, #99)
- Fix using `secrets.GITHUB_TOKEN` caused a type error when some other secret is defined. (thanks @mkj-is, #106)
- Fix nil check is missing on parsing `uses:` step. (thanks @shogo82148, #102)
- Fix some documents including broken links. (thanks @ohkinozomu, #105)
- Update popular actions data set to the latest. More arguments are added to many actions. And a few actions had new major versions.
- Update webhook payload data set to the latest. `requested_action` type was added to `check_run` hook. `requested` and `rerequested` types were removed from `check_suite` hook. `updated` type was removed from `project` hook.


[Changes][v1.6.9]


<a name="v1.6.8"></a>
# [v1.6.8](https://github.com/rhysd/actionlint/releases/tag/v1.6.8) - 15 Nov 2021

- [Untrusted inputs](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions) detection can detect untrusted inputs in object filter syntax. For example, `github.event.*.body` filters `body` properties and it includes the untrusted input `github.event.comment.body`. actionlint detects such filters and causes an error. The error message includes all untrusted input names which are filtered by the object filter so that you can know what inputs are untrusted easily. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#untrusted-inputs) for more details.
  Input example:
  ```yaml
  - name: Get comments
    run: echo '${{ toJSON(github.event.*.body) }}'
  ```
  Error message:
  ```
  object filter extracts potentially untrusted properties "github.event.comment.body", "github.event.discussion.body", "github.event.issue.body", ...
  ```
  Instead you should do:
  ```yaml
  - name: Get comments
    run: echo "$JSON"
    env:
      JSON: {{ toJSON(github.event.*.body) }}
  ```
- Support [the new input type syntax for `workflow_dispatch` event](https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/), which was introduced recently. You can declare types of inputs on triggering a workflow manually. actionlint does two things with this new syntax.
  - actionlint checks the syntax. Unknown input types, invalid default values, missing options for 'choice' type.
    ```yaml
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
      verbose:
        type: boolean
        # ERROR: Boolean value must be 'true' or 'false'
        default: yes
    ```
  - actionlint give a strict object type to `github.event.inputs` so that a type checker can check unknown input names and type mismatches on using the value.
    ```yaml
    on:
      workflow_dispatch:
        inputs:
          message:
            type: string
          verbose:
            type: boolean
    # Type of `github.event.inputs` is {"message": string; "verbose": bool}
    jobs:
      test:
        runs-on: ubuntu-latest
        steps:
          # ERROR: Undefined input
          - run: echo "${{ github.event.inputs.massage }}"
          # ERROR: Bool value is not available for object key
          - run: echo "${{ env[github.event.inputs.verbose] }}"
    ```
  - See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-workflow-dispatch-events) for more details.
- Add missing properties in `github` context. See [the contexts document](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) to know the full list of properties.
  - `github.ref_name` (thanks @dihmandrake, #72)
  - `github.ref_protected`
  - `github.ref_type`
- Filtered array by object filters is typed more strictly.
  ```
  # `env` is a map object { string => string }
  # Previously typed as array<any> now it is typed as array<string>
  env.*
  ```
- Update Go module dependencies and playground dependencies.

[Changes][v1.6.8]


<a name="v1.6.7"></a>
# [v1.6.7](https://github.com/rhysd/actionlint/releases/tag/v1.6.7) - 08 Nov 2021

- Fix missing property `name` in `runner` context object (thanks @ioanrogers, #67).
- Fix a false positive on type checking at `x.*` object filtering syntax where the receiver is an object. actionlint previously only allowed arrays as receiver of object filtering (#66).
  ```ruby
  fromJSON('{"a": "from a", "b": "from b"}').*
  # => ["from a", "from b"]

  fromJSON('{"a": {"x": "from a.x"}, "b": {"x": "from b.x"}}').*.x
  # => ["from a.x", "from b.x"]
  ```
- Add [rust-cache](https://github.com/Swatinem/rust-cache) as new popular action.
- Remove `bottle: unneeded` from Homebrew formula (thanks @oppara, #63).
- Support `branch_protection_rule` webhook again.
- Update popular actions data set to the latest (#64, #70).

[Changes][v1.6.7]


<a name="v1.6.6"></a>
# [v1.6.6](https://github.com/rhysd/actionlint/releases/tag/v1.6.6) - 17 Oct 2021

- `inputs` and `secrets` objects are now typed looking at `workflow_call` event at `on:`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-types-of-inputs-and-secrets-in-reusable-workflow) for more details.
  - `inputs` object is typed with definitions at `on.workflow_call.inputs`. When the workflow is not callable, it is typed at `{}` (empty object) so any `inputs.*` access causes a type error.
  - `secrets` object is typed with definitions at `on.workflow_call.secrets`.
  ```yaml
  on:
    workflow_call:
      # `inputs` object is typed {url: string; lucky_number: number}
      inputs:
        url:
          description: 'your URL'
          type: string
        lucky_number:
          description: 'your lucky number'
          type: number
      # `secrets` object is typed {user: string; credential: string}
      secrets:
        user:
          description: 'your user name'
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
- `id-token` is added to permissions (thanks @cmmarslender, #62).
- Report an error on nested workflow calls since it is [not allowed](https://docs.github.com/en/actions/learn-github-actions/reusing-workflows#limitations).
  ```yaml
  on:
    # This workflow is reusable
    workflow_call:

  jobs:
    test:
      # ERROR: Nested workflow call is not allowed
      uses: owner/repo/path/to/workflow.yml@ref
  ```
- Parse `uses:` at reusable workflow call more strictly following `{owner}/{repo}/{path}@{ref}` format.
- Popular actions data set was updated to the latest (#61).
- Dependencies of playground were updated to the latest (including eslint v8).

[Changes][v1.6.6]


<a name="v1.6.5"></a>
# [v1.6.5](https://github.com/rhysd/actionlint/releases/tag/v1.6.5) - 08 Oct 2021

- Support [reusable workflows](https://docs.github.com/en/actions/learn-github-actions/reusing-workflows) syntax which is now in beta. Only very basic syntax checks are supported at this time. Please see [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-reusable-workflows) to know checks for reusable workflow syntax.
  - Example of `workflow_call` event
    ```yaml
    on:
      workflow_call:
        inputs:
          name:
            description: your name
            type: string
        secrets:
          token:
            required: true

    jobs:
      ...
    ```
  - Example of reusable workflow call with `uses:` at `job.<job_id>`
    ```yaml
    on: ...
    jobs:
      hello:
        uses: owner/repo/path/to/workflow.yml@main
        with:
          name: Octocat
        secrets:
          token: ${{ secrets.token }}
    ```
- Support `github.run_attempt` property in `${{ }}` expression (#57).
- Add support for `windows-2022` runner which is now in [public beta](https://github.com/actions/virtual-environments/issues/3949).
- Remove support for `ubuntu-16.04` runner which was [removed from GitHub Actions at the end of September](https://github.com/actions/virtual-environments/issues/3287).
- Ignore [SC2154](https://github.com/koalaman/shellcheck/wiki/SC2154) shellcheck rule which can cause false positive (#53).
- Fix error position was not correct when required keys are not existing in job configuration.
- Update popular actions data set. New major versions of github-script and lock-threads actions are supported (#59).
- Fix document (thanks @fornwall at #52, thanks @equal-l2 at #56).
  - Now actionlint is [an official package of Homebrew](https://formulae.brew.sh/formula/actionlint). Simply executing `brew install actionlint` can install actionlint.

[Changes][v1.6.5]


<a name="v1.6.4"></a>
# [v1.6.4](https://github.com/rhysd/actionlint/releases/tag/v1.6.4) - 21 Sep 2021

- Implement 'map' object types `{ string => T }`, where all properties of the object are typed as `T`. Since a key of object is always string, left hand side of `=>` is fixed to `string`. For example, `env` context only has string properties so it is typed as `{ string => string}`. Previously its properties were typed `any`.
  ```yaml
  # typed as string (previously any)
  env.FOO

  # typed as { id: string; network: string; ports: object; } (previously any)
  job.services.redis
  ```
- `github.event.discussion.title` and `github.event.discussion.body` are now checked as untrusted inputs.
- Update popular actions data set. (#50, #51)
- Update webhooks payload data set. `branch_protection_rule` hook was dropped from the list due to [github/docs@179a6d3](https://github.com/github/docs/commit/179a6d334e92b9ade8626ef42a546dae66b49951). (#50, #51)

[Changes][v1.6.4]


<a name="v1.6.3"></a>
# [v1.6.3](https://github.com/rhysd/actionlint/releases/tag/v1.6.3) - 04 Sep 2021

- Improve guessing a type of matrix value. When a matrix contains numbers and strings, previously the type fell back to `any`. Now it is deduced as string.
  ```yaml
  strategy:
    matrix:
      # matrix.node is now deduced as `string` instead of `any`
      node: [14, 'latest']
  ```
- Fix types of `||` and `&&` expressions. Previously they were typed as `bool` but it was not correct. Correct type is sum of types of both sides of the operator like TypeScript. For example, type of `'foo' || 'bar'` is a string, and `github.event && matrix` is an object.
- actionlint no longer reports an error when a local action does not exist in the repository. It is a popular pattern that a local action directory is cloned while a workflow running. (#25, #40)
- Disable [SC2050](https://github.com/koalaman/shellcheck/wiki/SC2050) shellcheck rule since it causes some false positive. (#45)
- Fix `-version` did not work when running actionlint via [the Docker image](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) (#47).
- Fix pre-commit hook file name. (thanks @xsc27, #38)
- [New `branch_protection_rule` event](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#branch_protection_rule) is supported. (#48)
- Update popular actions data set. (#41, #48)
- Update Go library dependencies.
- Update playground dependencies.

[Changes][v1.6.3]


<a name="v1.6.2"></a>
# [v1.6.2](https://github.com/rhysd/actionlint/releases/tag/v1.6.2) - 23 Aug 2021

- actionlint now checks evaluated values at `${{ }}` are not an object nor an array since they are not useful. See [the check document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-type-check-expression) for more details.
```yaml
# ERROR: This will always be replaced with `echo 'Object'`
- run: echo '${{ runner }}'
# OK: Serialize an object into JSON to check the content
- run: echo '${{ toJSON(runner) }}'
```
- Add [pre-commit](https://pre-commit.com/) support. pre-commit is a framework for managing Git `pre-commit` hooks. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pre-commit) for more details. (thanks @xsc27 for adding the integration at #33) (#23)
- Add [an official Docker image](https://hub.docker.com/repository/docker/rhysd/actionlint). The Docker image contains shellcheck and pyflakes as dependencies. Now actionlint can be run with `docker run` command easily. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) for more details. (thanks @xsc27 for the help at #34)
```sh
docker run --rm -v $(pwd):/repo --workdir /repo rhysd/actionlint:latest -color
```
- Go 1.17 is now a default compiler to build actionlint. Built binaries are faster than before by 2~7% when the process is CPU-bound. Sizes of built binaries are about 2% smaller. Note that Go 1.16 continues to be supported.
- `windows/arm64` target is added to released binaries thanks to Go 1.17.
- Now any value can be converted into bool implicitly. Previously this was not permitted as actionlint provides stricter type check. However it is not useful that a condition like `if: github.event.foo` causes a type error.
- Fix a prefix operator cannot be applied repeatedly like `!!42`.
- Fix a potential crash when type checking on expanding an object with `${{ }}` like `matrix: ${{ fromJSON(env.FOO) }}`
- Update popular actions data set (#36)

[Changes][v1.6.2]


<a name="v1.6.1"></a>
# [v1.6.1](https://github.com/rhysd/actionlint/releases/tag/v1.6.1) - 16 Aug 2021

- [Problem Matchers](https://github.com/actions/toolkit/blob/master/docs/problem-matchers.md) is now officially supported by actionlint, which annotates errors from actionlint on GitHub as follows. The matcher definition is maintained at [`.github/actionlint-matcher.json`](https://github.com/rhysd/actionlint/blob/main/.github/actionlint-matcher.json) by [script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-actionlint-matcher). For the usage, see [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#problem-matchers).

<img src="https://github.com/rhysd/ss/blob/master/actionlint/problem-matcher.png?raw=true" alt="annotation by Problem Matchers" width="715" height="221"/>

- `runner_label` rule now checks conflicts in labels at `runs-on`. For example, there is no runner which meats both `ubuntu-latest` and `windows-latest`. This kind of misconfiguration sometimes happen when a beginner misunderstands the usage of `runs-on:`. To run a job on each runners, `matrix:` should be used. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-runner-labels) for more information.

```yaml
on: push
jobs:
  test:
    # These labels match to no runner
    runs-on: [ubuntu-latest, windows-latest]
    steps:
      - run: echo ...
```

- Reduce memory footprint (around 16%) on starting `actionlint` command by removing unnecessary data from `PopularActions` global variable. This also slightly reduces binary size (about 3.7% at `playground/main.wasm`).
- Fix accessing `steps.*` objects in job's `environment:` configuration caused a type error (#30).
- Fix checking that action's input names at `with:` were not in case insensitive (#31).
- Ignore outputs of [getsentry/paths-filter](https://github.com/getsentry/paths-filter). It is a fork of [dorny/paths-filter](https://github.com/dorny/paths-filter). actionlint cannot check the outputs statically because it sets outputs dynamically.
- Add [Azure/functions-action](https://github.com/Azure/functions-action) to popular actions.
- Update popular actions data set (#29).

[Changes][v1.6.1]


<a name="v1.6.0"></a>
# [v1.6.0](https://github.com/rhysd/actionlint/releases/tag/v1.6.0) - 11 Aug 2021

- Check potentially untrusted inputs to prevent [a script injection vulnerability](https://securitylab.github.com/research/github-actions-untrusted-input/) at `run:` and `script` input of [actions/github-script](https://github.com/actions/github-script). See [the rule document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#untrusted-inputs) for more explanations and workflow example. (thanks @azu for the feature request at #19)

Incorrect code

```yaml
- run: echo '${{ github.event.pull_request.title }}'
```

should be replaced with

```yaml
- run: echo "issue ${TITLE}"
  env:
    TITLE: ${{github.event.issue.title}}
```

- Add `-format` option to `actionlint` command. It allows to flexibly format error messages as you like with [Go template syntax](https://pkg.go.dev/text/template). See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format) for more details. (thanks @ybiquitous for the feature request at #20)

Simple example to output error messages as JSON:

```sh
actionlint -format '{{json .}}'
```

More compliated example to output error messages as markdown:

```sh
actionlint -format '{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\n\n{{$.Message}}\n\n```\n{{$.Snippet}}\n```\n\n{{end}}'
```

- Documents are reorganized. Long `README.md` is separated into several document files (#28)
  - [`README.md`](https://github.com/rhysd/actionlint/blob/main/README.md): Introduction, Quick start, Document links
  - [`docs/checks.md`](https://github.com/rhysd/actionlint/tree/main/docs/checks.md): Full list of all checks done by actionlint with example inputs, outputs, and playground links
  - [`docs/install.md`](https://github.com/rhysd/actionlint/tree/main/docs/install.md): Installation instruction
  - [`docs/usage.md`](https://github.com/rhysd/actionlint/tree/main/docs/usage.md): Advanced usage of `actionlint` command, usage of playground, integration with [reviewdog](https://github.com/reviewdog/reviewdog), [Problem Matchers](https://github.com/actions/toolkit/blob/master/docs/problem-matchers.md), [super-linter](https://github.com/github/super-linter)
  - [`docs/config.md`](https://github.com/rhysd/actionlint/tree/main/docs/config.md): About configuration file
  - [`doc/api.md`](https://github.com/rhysd/actionlint/tree/main/docs/api.md): Using actionlint as Go library
  - [`doc/reference.md`](https://github.com/rhysd/actionlint/tree/main/docs/reference.md): Links to resources
- Fix checking shell names was not case-insensitive, for example `PowerShell` was detected as invalid shell name
- Update popular actions data set to the latest
- Make lexer errors on checking `${{ }}` expressions more meaningful

[Changes][v1.6.0]


<a name="v1.5.3"></a>
# [v1.5.3](https://github.com/rhysd/actionlint/releases/tag/v1.5.3) - 04 Aug 2021

- Now actionlint allows to use any operators outside `${{ }}` on `if:` condition like `if: github.repository_owner == 'rhysd'` (#22). [The official document](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif) said that using any operator outside `${{ }}` was invalid even if it was on `if:` condition. However, [github/docs#8786](https://github.com/github/docs/pull/8786) clarified that the document was not correct.

[Changes][v1.5.3]


<a name="v1.5.2"></a>
# [v1.5.2](https://github.com/rhysd/actionlint/releases/tag/v1.5.2) - 02 Aug 2021

- Outputs of [dorny/paths-filter](https://github.com/dorny/paths-filter) are now not typed strictly because the action dynamically sets outputs which are not defined in its `action.yml`. actionlint cannot check such outputs statically (#18).
- [The table](https://github.com/rhysd/actionlint/blob/main/all_webhooks.go) for checking [Webhooks supported by GitHub Actions](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#webhook-events) is now generated from the official document automatically with [script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-webhook-events). The table continues to be updated weekly by [the CI workflow](https://github.com/rhysd/actionlint/actions/workflows/generate.yaml).
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


[v1.6.10]: https://github.com/rhysd/actionlint/compare/v1.6.9...v1.6.10
[v1.6.9]: https://github.com/rhysd/actionlint/compare/v1.6.8...v1.6.9
[v1.6.8]: https://github.com/rhysd/actionlint/compare/v1.6.7...v1.6.8
[v1.6.7]: https://github.com/rhysd/actionlint/compare/v1.6.6...v1.6.7
[v1.6.6]: https://github.com/rhysd/actionlint/compare/v1.6.5...v1.6.6
[v1.6.5]: https://github.com/rhysd/actionlint/compare/v1.6.4...v1.6.5
[v1.6.4]: https://github.com/rhysd/actionlint/compare/v1.6.3...v1.6.4
[v1.6.3]: https://github.com/rhysd/actionlint/compare/v1.6.2...v1.6.3
[v1.6.2]: https://github.com/rhysd/actionlint/compare/v1.6.1...v1.6.2
[v1.6.1]: https://github.com/rhysd/actionlint/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/rhysd/actionlint/compare/v1.5.3...v1.6.0
[v1.5.3]: https://github.com/rhysd/actionlint/compare/v1.5.2...v1.5.3
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

 <!-- Generated by https://github.com/rhysd/changelog-from-release -->
