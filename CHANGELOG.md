<a id="v1.7.8"></a>
# [v1.7.8](https://github.com/rhysd/actionlint/releases/tag/v1.7.8) - 2025-10-11

- Support `models` permission in `permissions` section. ([#531](https://github.com/rhysd/actionlint/issues/531), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- Support `job.check_run_id` property. ([#576](https://github.com/rhysd/actionlint/issues/576), thanks [@muzimuzhi](https://github.com/muzimuzhi) for fixing the type at [#577](https://github.com/rhysd/actionlint/issues/577))
- Support `node24` runtime at `using` section in action metadata. ([#561](https://github.com/rhysd/actionlint/issues/561), thanks [@salmanmkc](https://github.com/salmanmkc))
- Add support for the following runner labels
  - [`macos-26` and `macos-26-large`](https://github.blog/changelog/2025-09-11-actions-macos-26-image-now-in-public-preview/) ([#572](https://github.com/rhysd/actionlint/issues/572), thanks [@muzimuzhi](https://github.com/muzimuzhi))
  - [`macos-15`](https://github.blog/changelog/2025-09-19-github-actions-macos-13-runner-image-is-closing-down/#what-you-need-to-do) ([#572](https://github.com/rhysd/actionlint/issues/572), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- Drop support for the following runner labels.
  - [`ubuntu-20.04`](https://github.com/actions/runner-images/issues/11101) ([#534](https://github.com/rhysd/actionlint/issues/534), thanks [@shogo82148](https://github.com/shogo82148))
  - [`windows-2019`](https://github.blog/changelog/2025-04-15-upcoming-breaking-changes-and-releases-for-github-actions/#windows-server-2019-is-closing-down) ([#572](https://github.com/rhysd/actionlint/issues/572), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- Support [`deprecationMessage`](https://docs.github.com/en/actions/reference/workflows-and-actions/metadata-syntax#inputsinput_iddeprecationmessage) in action inputs. ([#540](https://github.com/rhysd/actionlint/issues/540), thanks [@saansh45](https://github.com/saansh45))
- Support [`windows-11-arm` runner](https://github.blog/changelog/2025-04-14-windows-arm64-hosted-runners-now-available-in-public-preview/). ([#542](https://github.com/rhysd/actionlint/issues/542), thanks [@trim21](https://github.com/trim21))
- Handle `ubuntu-latest` runner label as `ubuntu-24.04` and `macos-latest` runner label as `macos-15`.
- Report mixing Intel Mac labels and Arm Mac labels as error.
- Add new types to `issues` and `pull_request_target` webhooks.
- Update the popular actions data set to the latest and add more actions to it (thanks [@sethvargo](https://github.com/sethvargo) for fixing the `go generate` scripts)
  - `actions/create-github-app-token`
  - `actions/attest-sbom`
  - `actions/ai-inference`
  - `peter-evans/create-or-update-comment`
  - `release-drafter/release-drafter`
  - `SamKirkland/FTP-Deploy-Action`
- Fix the version value in `actionlint -version` can be empty.
- Fix outdated URL links in some error messages and documents.
- [Homebrew formula in this repository](https://github.com/rhysd/actionlint/blob/main/HomebrewFormula/actionlint.rb) is deprecated and [Homebrew cask](https://github.com/rhysd/actionlint/blob/main/Casks/actionlint.rb) is newly added instead because [GoReleaser no longer supports Homebrew formula update](https://goreleaser.com/deprecations/#brews). Note that Homebrew's official `actionlint` formula is still maintained. Please read the [documentation](https://github.com/rhysd/actionlint/blob/main/docs/install.md#homebrew) for more details.
- Drop support for Go 1.23 and earlier because they are no longer maintained officially. Go 1.24 and later are supported to build actionlint.
- Replace [`go-yaml/yaml@v3`](https://github.com/go-yaml/yaml) package with [`yaml/go-yaml@v4`](https://github.com/yaml/go-yaml) package. `go-yaml/yaml` was used for parsing workflow files however it was unmaintained. `yaml/go-yaml` is a successor of the library officially maintained by YAML organization. ([#575](https://github.com/rhysd/actionlint/issues/575))
- Improve error messages on parsing workflow and action metadata files.

[Changes][v1.7.8]


<a id="v1.7.7"></a>
# [v1.7.7](https://github.com/rhysd/actionlint/releases/tag/v1.7.7) - 2025-01-19

- Support runner labels for [Linux arm64 hosted runners](https://github.blog/changelog/2025-01-16-linux-arm64-hosted-runners-now-available-for-free-in-public-repositories-public-preview/). ([#503](https://github.com/rhysd/actionlint/issues/503), [#504](https://github.com/rhysd/actionlint/issues/504), thanks [@martincostello](https://github.com/martincostello))
  - `ubuntu-24.04-arm`
  - `ubuntu-22.04-arm`
- Update Go dependencies to the latest.
- Update the popular actions data set to the latest.
- Add Linux arm64 job to our CI workflow. Now actionlint is tested on the platform. ([#507](https://github.com/rhysd/actionlint/issues/507), thanks [@cclauss](https://github.com/cclauss))

[Changes][v1.7.7]


<a id="v1.7.6"></a>
# [v1.7.6](https://github.com/rhysd/actionlint/releases/tag/v1.7.6) - 2025-01-04

- Fix using contexts at specific workflow keys is incorrectly reported as not allowed. Affected workflow keys are as follows. ([#495](https://github.com/rhysd/actionlint/issues/495), [#497](https://github.com/rhysd/actionlint/issues/497), [#498](https://github.com/rhysd/actionlint/issues/498), [#500](https://github.com/rhysd/actionlint/issues/500))
  - `jobs.<job_id>.steps.with.args`
  - `jobs.<job_id>.steps.with.entrypoint`
  - `jobs.<job_id>.services.<service_id>.env`
- Update Go dependencies to the latest.

[Changes][v1.7.6]


<a id="v1.7.5"></a>
# [v1.7.5](https://github.com/rhysd/actionlint/releases/tag/v1.7.5) - 2024-12-28

- Strictly check available contexts in `${{ }}` placeholders following the ['Context availability' table](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/accessing-contextual-information-about-workflow-runs#context-availability) in the official document.
  - For example, `jobs.<job_id>.defaults.run.shell` allows `env` context but `shell` workflow keys in other places allow no context.
    ```yaml
    defaults:
      run:
        # ERROR: No context is available here
        shell: ${{ env.SHELL }}

    jobs:
      test:
        runs-on: ubuntu-latest
        defaults:
          run:
            # OK: 'env' context is available here
            shell: ${{ env.SHELL }}
        steps:
          - run: echo hello
            # ERROR: No context is available here
            shell: ${{ env.SHELL}}
    ```
- Check a string literal passed to `fromJSON()` call. This pattern is [popular](https://github.com/search?q=fromJSON%28%27+lang%3Ayaml&type=code) to create array or object constants because GitHub Actions does not provide the literal syntax for them. See the [document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#contexts-and-built-in-functions) for more details. ([#464](https://github.com/rhysd/actionlint/issues/464))
  ```yaml
  jobs:
    test:
      # ERROR: Key 'mac' does not exist in the object returned by the fromJSON()
      runs-on: ${{ fromJSON('{"win":"windows-latest","linux":"ubuntul-latest"}')['mac'] }}
      steps:
        - run: echo This is a special branch!
          # ERROR: Broken JSON string passed to fromJSON.
          if: contains(fromJSON('["main","release","dev"'), github.ref_name)
  ```
- Allow passing command arguments to `-shellcheck` argument. ([#483](https://github.com/rhysd/actionlint/issues/483), thanks [@anuraaga](https://github.com/anuraaga))
  - This is useful when you want to use alternative build of shellcheck like [go-shellcheck](https://github.com/wasilibs/go-shellcheck/).
    ```sh
    actionlint -shellcheck="go run github.com/wasilibs/go-shellcheck/cmd/shellcheck@latest"
    ```
- Support undocumented `repository_visibility`, `artifact_cache_size_limit`, `step_summary`, `output`, `state` properties in `github` context. ([#489](https://github.com/rhysd/actionlint/issues/489), thanks [@rasa](https://github.com/rasa) for adding `repository_visibility` property)
- Remove `macos-12` runner label from known labels because it was [dropped](https://github.com/actions/runner-images/issues/10721) from GitHub-hosted runners on Dec. 3 and is no longer available.
- Add `windows-2025` runner label to the known labels. The runner is in [public preview](https://github.blog/changelog/2024-12-19-windows-server-2025-is-now-in-public-preview/). ([#491](https://github.com/rhysd/actionlint/issues/491), thanks [@ericcornelissen](https://github.com/ericcornelissen))
- Add `black` to the list of colors for `branding.color` action metadata. ([#485](https://github.com/rhysd/actionlint/issues/485), thanks [@eifinger](https://github.com/eifinger))
- Add `table` to the list of icons for `branding.icon` action metadata.
- Fix parsing escaped `{` in `format()` function call's first argument.
- Fix the incorrect `join()` function overload. `join(s1: string, s2: string)` was wrongly accepted.
- Update popular actions data set to the latest.
  - Add `download-artifact/v3-node20` to the data set. ([#468](https://github.com/rhysd/actionlint/issues/468))
  - Fix missing the `reviewdog/action-hadolint@v1` action input. ([#487](https://github.com/rhysd/actionlint/issues/487), thanks [@mi-wada](https://github.com/mi-wada))
- Link to the documents of the stable version in actionlint `man` page and `-help` output.
- Refactor `LintStdin()` API example and some unit tests. ([#472](https://github.com/rhysd/actionlint/issues/472), [#475](https://github.com/rhysd/actionlint/issues/475), thanks [@alexandear](https://github.com/alexandear))
- Improve the configuration example in `actionlint.yaml` document to explain glob patterns for `paths`. ([#481](https://github.com/rhysd/actionlint/issues/481))

[Changes][v1.7.5]


<a id="v1.7.4"></a>
# [v1.7.4](https://github.com/rhysd/actionlint/releases/tag/v1.7.4) - 2024-11-04

- Disallow the usage of popular actions that run on `node16` runner. The `node16` runner [will reach the end of life on November 12](https://github.blog/changelog/2024-09-25-end-of-life-for-actions-node16/).
  - In case of the error, please update your actions to the latest version so that they run on the latest `node20` runner.
  - If you're using self-hosted runner and you cannot upgrade your runner to `node20` soon, please consider to ignore the error by the `paths` configuration described below.
  - If you're using `actions/upload-artifact@v3` and `actions/download-artifact@v3` on GHES, please replace them with `actions/upload-artifact@v3-node20` and `actions/download-artifact@v3-node20`. ([#468](https://github.com/rhysd/actionlint/issues/468))
- Provide the configuration for ignoring errors by regular expressions in `actionlint.yml` (or `actionlint.yaml`). Please see the [document](https://github.com/rhysd/actionlint/blob/v1.7.4/docs/config.md) for more details. ([#217](https://github.com/rhysd/actionlint/issues/217), [#342](https://github.com/rhysd/actionlint/issues/342))
  - The `paths` is a mapping from the file path glob pattern to the corresponding configuration. The `ignore` configuration is a list of regular expressions to match error messages (similar to the `-ignore` command line option).
    ```yaml
    paths:
      # This pattern matches any YAML file under the '.github/workflows/' directory.
      .github/workflows/**/*.yaml:
        ignore:
          # Ignore the specific error from shellcheck
          - 'shellcheck reported issue in this script: SC2086:.+'
      # This pattern only matches '.github/workflows/release.yaml' file.
      .github/workflows/release.yaml:
        ignore:
          # Ignore errors from the old runner check. This may be useful for (outdated) self-hosted runner environment.
          - 'the runner of ".+" action is too old to run on GitHub Actions'
    ```
  - This configuration was not implemented initially because I wanted to keep the configuration as minimal as possible. However, due to several requests for it, the configuration has now been added.
- Untrusted inputs check is safely skipped inside specific function calls. ([#459](https://github.com/rhysd/actionlint/issues/459), thanks [@IlyaGulya](https://github.com/IlyaGulya))
  - For example, the following step contains the untrusted input `github.head_ref`, but it is safe because it's passed to the `contains()` argument.
    ```yaml
    - run: echo "is_release_branch=${{ contains(github.head_ref, 'release') }}" >> "$GITHUB_OUTPUT"
    ```
  - For more details, please read the [rule document](https://github.com/rhysd/actionlint/blob/v1.7.4/docs/checks.md#untrusted-inputs).
- Recognize `gcr.io` and `gcr.dev` as the correct container registry hosts. ([#463](https://github.com/rhysd/actionlint/issues/463), thanks [@takaidohigasi](https://github.com/takaidohigasi))
  - Note that it is recommended explicitly specifying the scheme like `docker://gcr.io/...`.
- Remove `macos-x.0` runner labels which are no longer available. ([#452](https://github.com/rhysd/actionlint/issues/452))
- Disable shellcheck [`SC2043`](https://www.shellcheck.net/wiki/SC2043) rule because it can cause false positives on checking `run:`. ([#355](https://github.com/rhysd/actionlint/issues/355))
  - The [rule document](https://github.com/rhysd/actionlint/blob/v1.7.4/docs/checks.md#check-shellcheck-integ) was updated as well. ([#466](https://github.com/rhysd/actionlint/issues/466), thanks [@risu729](https://github.com/risu729))
- Fix the error message was not deterministic when detecting cycles in `needs` dependencies.
- Fix the check for `format()` function was not applied when the function name contains upper case like `Format()`. Note that function names in `${{ }}` placeholders are case-insensitive.
- Update the popular actions data set to the latest.
  - This includes the [new `ref` and `commit` outputs](https://github.com/actions/checkout/pull/1180) of `actions/checkout`.
- Add [`actions/cache/save`](https://github.com/actions/cache/tree/main/save) and [`actions/cache/restore`](https://github.com/actions/cache/tree/main/restore) to the popular actions data set.
- Links in the [README.md](https://github.com/rhysd/actionlint/blob/main/README.md) now point to the document of the latest version tag instead of HEAD of `main` branch.
- Add [`Linter.LintStdin`](https://pkg.go.dev/github.com/rhysd/actionlint#Linter.LintStdin) method dedicated to linting STDIN instead of handling STDIN in `Command`.
- (Dev) Add new [`check-checks` script](https://github.com/rhysd/actionlint/tree/main/scripts/check-checks) to maintain the ['Checks' document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md). It automatically updates the outputs and playground links for example inputs in the document. It also checks the document is up-to-date on CI. Please read the [document](https://github.com/rhysd/actionlint/blob/main/scripts/check-checks/README.md) for more details.

[Documentation](https://github.com/rhysd/actionlint/tree/v1.7.4/docs)

[Changes][v1.7.4]


<a id="v1.7.3"></a>
# [v1.7.3](https://github.com/rhysd/actionlint/releases/tag/v1.7.3) - 2024-09-29

- Remove `macos-11` runner labels because [macOS 11 runner was dropped on 6/28/2024](https://github.blog/changelog/2024-05-20-actions-upcoming-changes-to-github-hosted-macos-runners/#macos-11-deprecation-and-removal). ([#451](https://github.com/rhysd/actionlint/issues/451), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- Support `macos-15`, `macos-15-large`, and `macos-15-xlarge` runner labels. The macOS 15 runner is not globally available yet, but [they are available in beta](https://github.com/actions/runner-images?tab=readme-ov-file#available-images). ([#453](https://github.com/rhysd/actionlint/issues/453), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- Release artifact includes checksums for the released binaries. The file name is `actionlint_{version}_checksums.txt`. ([#449](https://github.com/rhysd/actionlint/issues/449))
  - For example, the checksums for v1.7.3 can be found [here](https://github.com/rhysd/actionlint/releases/download/v1.7.3/actionlint_1.7.3_checksums.txt).
- Fix `download-path` output is missing in `actions/download-artifact@v3` action. ([#442](https://github.com/rhysd/actionlint/issues/442))
  - Note that the latest version `actions/download-artifact@v4` was not affected by this issue.
- Support Go 1.23.

[Documentation](https://github.com/rhysd/actionlint/blob/v1.7.3/docs/checks.md)

[Changes][v1.7.3]


<a id="v1.7.2"></a>
# [v1.7.2](https://github.com/rhysd/actionlint/releases/tag/v1.7.2) - 2024-09-23

- Fix child processes to run in parallel.
- Update the popular actions data set to the latest. ([#442](https://github.com/rhysd/actionlint/issues/442), [#445](https://github.com/rhysd/actionlint/issues/445), [#446](https://github.com/rhysd/actionlint/issues/446), [#447](https://github.com/rhysd/actionlint/issues/447), thanks [@maikelvdh](https://github.com/maikelvdh))
- Add support for checking branch filters on [`merge_group` event](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#merge_group). ([#448](https://github.com/rhysd/actionlint/issues/448), thanks [@muzimuzhi](https://github.com/muzimuzhi))
- [The playground](https://rhysd.github.io/actionlint/) now supports both light and dark modes and automatically applies the system's theme.
- Fix releasing a failure on making a new winget package. ([#438](https://github.com/rhysd/actionlint/issues/438), thanks [@vedantmgoyal9](https://github.com/vedantmgoyal9))

[Changes][v1.7.2]


<a id="v1.7.1"></a>
# [v1.7.1](https://github.com/rhysd/actionlint/releases/tag/v1.7.1) - 2024-05-28

- Support `ubuntu-24.04` runner label, which was [recently introduced as beta](https://github.blog/changelog/2024-05-14-github-hosted-runners-public-beta-of-ubuntu-24-04-is-now-available/). ([#425](https://github.com/rhysd/actionlint/issues/425), thanks [@bitcoin-tools](https://github.com/bitcoin-tools))
- Remove the support for `macos-10` runner label which was [officially dropped about 2 years ago](https://github.blog/changelog/2022-07-20-github-actions-the-macos-10-15-actions-runner-image-is-being-deprecated-and-will-be-removed-by-8-30-22/).
- Remove the support for `windows-2016` runner label which was [officially dropped about 2 years ago](https://github.blog/changelog/2021-10-19-github-actions-the-windows-2016-runner-image-will-be-removed-from-github-hosted-runners-on-march-15-2022/).
- Document URLs used in help output and links in the playground prefer specific version tag rather than `main` branch. For example,
  - Before: https://github.com/rhysd/actionlint/tree/main/docs
  - After: https://github.com/rhysd/actionlint/tree/v1.7.1/docs
- Fix actionlint wrongly reports an error when using `ghcr.io` or `docker.io` at `image` field of action metadata file of Docker action without `docker://` scheme. ([#428](https://github.com/rhysd/actionlint/issues/428))
  ```yaml
  runs:
    using: 'docker'
    # This should be OK
    image: 'ghcr.io/user/repo:latest'
  ```
- Fix checking `preactjs/compressed-size-action@v2` usage caused a false positive. ([#422](https://github.com/rhysd/actionlint/issues/422))
- Fix an error message when invalid escaping is found in globs.
- The design of the [playground page](https://rhysd.github.io/actionlint/) is overhauled following the upgrade of bulma package to v1.
  - Current actionlint version is shown in the heading.
  - The color theme is changed to the official dark theme.
  - The list of useful links is added to the bottom of the page as 'Resources' section.

[Changes][v1.7.1]


<a id="v1.7.0"></a>
# [v1.7.0](https://github.com/rhysd/actionlint/releases/tag/v1.7.0) - 2024-05-08

- From this version, actionlint starts to check action metadata file `action.yml` (or `action.yaml`). At this point, only very basic checks are implemented and contents of `steps:` are not checked yet.
  - It checks properties under `runs:` section (e.g. `main:` can be specified when it is a JavaScript action), `branding:` properties, and so on.
    ```yaml
    name: 'My action'
    author: '...'
    # ERROR: 'description' section is missing

    branding:
      # ERROR: Invalid icon name
      icon: dog

    runs:
      # ERROR: Node.js runtime version is too old
      using: 'node12'
      # ERROR: The source file being run by this action does not exist
      main: 'this-file-does-not-exist.js'
      # ERROR: 'env' configuration is only allowed for Docker actions
      env:
        SOME_VAR: SOME_VALUE
    ```
  - actionlint still focuses on checking workflow files. So there is no way to directly specify `action.yml` as an argument of `actionlint` command. actionlint checks all local actions which are used by given workflows. If you want to use actionlint for your action development, prepare a test/example workflow which uses your action, and check it with actionlint instead.
  - Checks for `steps:` contents are planned to be implemented. Since several differences are expected between `steps:` in workflow file and `steps:` in action metadata file (e.g. available contexts), the implementation is delayed to later version. And the current implementation of action metadata parser is ad hoc. I'm planning a large refactorying and breaking changes Go API around it are expected.
- Add `runner.environment` property. ([#412](https://github.com/rhysd/actionlint/issues/412))
  ```yaml
  - run: echo 'Run by GitHub-hosted runner'
    if: runner.environment == 'github-hosted'
  ```
- Using outdated popular actions is now detected at error. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#detect-outdated-popular-actions) for more details.
  - Here 'outdated' means actions which use runtimes no longer supported by GitHub-hosted runners such as `node12`.
    ```yaml
    # ERROR: actions/checkout@v2 is using the outdated runner 'node12'
    - uses: actions/checkout@v2
    ```
- Support `attestations` permission which was [recently added to GitHub Actions as beta](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds). ([#418](https://github.com/rhysd/actionlint/issues/418), thanks [@bdehamer](https://github.com/bdehamer))
  ```yaml
  permissions:
    id-token: write
    contents: read
    attestations: write
  ```
- Check comparison expressions more strictly. Arbitrary types of operands can be compared as [the official document](https://docs.github.com/en/actions/learn-github-actions/expressions#operators) explains. However, comparisons between some types are actually meaningless because the values are converted to numbers implicitly. actionlint catches such meaningless comparisons as errors. Please see [the check document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-comparison-types) for more details.
  ```yaml
  on:
    workflow_call:
      inputs:
        timeout:
          type: boolean

  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
        - run: echo 'called!'
          # ERROR: Comparing string to object is always evaluated to false
          if: ${{ github.event == 'workflow_call' }}
        - run: echo 'timeout is too long'
          # ERROR: Comparing boolean value with `>` doesn't make sense
          if: ${{ inputs.timeout > 60 }}
  ```
- Follow the update that `macos-latest` is now an alias to `macos-14` runner.
- Support a custom python shell by `pyflakes` rule.
- Add workaround actionlint reports that `dorny/paths-filter`'s `predicate-quantifier` input is not defined. ([#416](https://github.com/rhysd/actionlint/issues/416))
- Fix the type of a conditional expression by comparison operators is wider than expected by implementing type narrowing. ([#384](https://github.com/rhysd/actionlint/issues/384))
  - For example, the type of following expression should be `number` but it was actually `string | number` and actionlint complained that `timeout-minutes` must take a number value.
    ```yaml
    timeout-minutes: ${{ env.FOO && 10 || 60 }}
    ```
- Fix `${{ }}` placeholder is not available at `jobs.<job_id>.services`. ([#402](https://github.com/rhysd/actionlint/issues/402))
  ```yaml
  jobs:
    test:
      services: ${{ fromJSON('...') }}
      runs-on: ubuntu-latest
      steps:
        - run: ...
  ```
- Do not check outputs of `google-github-actions/get-secretmanager-secrets` because this action sets outputs dynamically. ([#404](https://github.com/rhysd/actionlint/issues/404))
- Fix `defaults.run` is ignored on detecting the shell used in `run:`. ([#409](https://github.com/rhysd/actionlint/issues/409))
  ```yaml
  defaults:
    run:
      shell: pwsh
  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
        # This was wrongly detected as bash script
        - run: $Env:FOO = "FOO"
  ```
- Fix parsing a syntax error reported from pyflakes when checking a Python script in `run:`. ([#411](https://github.com/rhysd/actionlint/issues/411))
  ```yaml
  - run: print(
    shell: python
  ```
- Skip checking `exclude:` items in `matrix:` when they are constructed from `${{ }}` dynamically. ([#414](https://github.com/rhysd/actionlint/issues/414))
  ```yaml
  matrix:
    foo: ['a', 'b']
    exclude:
      # actionlint complained this value didn't exist in matrix combinations
      - foo: ${{ env.EXCLUDE_FOO }}
  ```
- Fix checking `exclude:` items when `${{ }}` is used in nested arrays at matrix items.
  ```yaml
  matrix:
    foo:
      - ["${{ fromJSON('...') }}"]
    exclude:
      # actionlint complained this value didn't match to any matrix combinations
      - foo: ['foo']
  ```
- Update popular actions data set. New major versions are added and the following actions are newly added.
  - `peaceiris/actions-hugo`
  - `actions/attest-build-provenance`
  - `actions/add-to-project`
  - `octokit/graphql-action`
- Update Go dependencies to the latest.
- Reduce the size of `actionlint` executable by removing redundant data from popular actions data set.
  - x86_64 executable binary size was reduced from 6.9MB to 6.7MB (2.9% smaller).
  - Wasm binary size was reduced from 9.4MB to 8.9MB (5.3% smaller).
- Describe how to [integrate actionlint to Pulsar Edit](https://web.pulsar-edit.dev/packages/linter-github-actions) in [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pulsar-edit). ([#408](https://github.com/rhysd/actionlint/issues/408), thanks [@mschuchard](https://github.com/mschuchard))
- Update outdated action versions in the usage document. ([#413](https://github.com/rhysd/actionlint/issues/413), thanks [@naglis](https://github.com/naglis))

[Changes][v1.7.0]


<a id="v1.6.27"></a>
# [v1.6.27](https://github.com/rhysd/actionlint/releases/tag/v1.6.27) - 2024-02-24

- Add macOS 14 runner labels for [Apple Silicon support](https://github.blog/changelog/2024-01-30-github-actions-macos-14-sonoma-is-now-available/). The following labels are added. (thanks [@harryzcy](https://github.com/harryzcy), [#392](https://github.com/rhysd/actionlint/issues/392))
  - `macos-14`
  - `macos-14-xlarge`
  - `macos-14-large`
- Remove `ubuntu-18.04` runner label from runners list since [it is no longer supported](https://github.blog/changelog/2022-08-09-github-actions-the-ubuntu-18-04-actions-runner-image-is-being-deprecated-and-will-be-removed-by-12-1-22/). ([#363](https://github.com/rhysd/actionlint/issues/363))
- Allow glob patterns in `self-hosted-runner.labels` configuration. For example, the following configuration defines any runner labels prefixed with `private-linux-`. (thanks [@kishaningithub](https://github.com/kishaningithub), [#378](https://github.com/rhysd/actionlint/issues/378))
  ```yaml
  self-hosted-runner:
    labels:
      - private-linux-*
  ```
- Fix a race condition bug when `-format` option is used for linting multiple workflow files. Thanks [@ReinAchten-TomTom](https://github.com/ReinAchten-TomTom) for your help on the investigation. ([#370](https://github.com/rhysd/actionlint/issues/370))
- Fix a race condition due to conflicts between some goroutine which starts to run shellcheck process and other goroutine which starts to wait until all processes finish.
- The popular actions data set was updated to the latest and the following actions were newly added. (thanks [@jmarshall](https://github.com/jmarshall), [#380](https://github.com/rhysd/actionlint/issues/380))
  - `google-github-actions/auth`
  - `google-github-actions/get-secretmanager-secrets`
  - `google-github-actions/setup-gcloud`
  - `google-github-actions/upload-cloud-storage`
  - `pulumi/actions`
  - `pypa/gh-action-pypi-publish`
- Add support for larger runner labels. The following labels are added. (thanks [@therealdwright](https://github.com/therealdwright), [#371](https://github.com/rhysd/actionlint/issues/371))
  - `windows-latest-8-cores`
  - `ubuntu-latest-4-cores`
  - `ubuntu-latest-8-cores`
  - `ubuntu-latest-16-cores`
- The following WebHook types are supported for `pull_request` event.
  - `enqueued`
  - `dequeued`
  - `milestoned`
  - `demilestoned`
- Explain how to control shellckeck behavior in the [shellcheck rule document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-shellcheck-integ). Use `SHELLCHECK_OPTS` environment variable to pass arguments to shellcheck. See [the shellcheck's official document](https://github.com/koalaman/shellcheck/wiki/Integration#environment-variables) for more details.
  ```
  # Enable some optional rules
  SHELLCHECK_OPTS='--enable=avoid-nullary-conditions' actionlint
  # Disable some rules
  SHELLCHECK_OPTS='--exclude=SC2129' actionlint
  ```
- Explicitly specify `docker.io` host name in pre-commit hook. (thanks [@gotmax23](https://github.com/gotmax23), [#382](https://github.com/rhysd/actionlint/issues/382))
- Explain how to report issues and send patches in [CONTRIBUTING.md](https://github.com/rhysd/actionlint/blob/main/CONTRIBUTING.md).
- Fix the link to super-linter project. (thanks [@zkoppert](https://github.com/zkoppert), [#376](https://github.com/rhysd/actionlint/issues/376))
- Add the instruction to install actionlint via the Arch Linux's official repository. (thanks [@sorairolake](https://github.com/sorairolake), [#381](https://github.com/rhysd/actionlint/issues/381))
- Prefer fixed revisions in the pre-commit usage. (thanks [@corneliusroemer](https://github.com/corneliusroemer), [#354](https://github.com/rhysd/actionlint/issues/354))
- Add instructions to use actionlint with Emacs. (thanks [@tirimia](https://github.com/tirimia), [#341](https://github.com/rhysd/actionlint/issues/341))
- Add instructions to use actionlint with Vim and Neovim text editors.
- Add [`actionlint.RuleBase.Config`](https://pkg.go.dev/github.com/rhysd/actionlint#RuleBase.Config) method to get the actionlint configuration passed to rules. (thanks [@hugo-syn](https://github.com/hugo-syn), [#387](https://github.com/rhysd/actionlint/issues/387))
- Add [`actionlint.ContainsExpression`](https://pkg.go.dev/github.com/rhysd/actionlint#ContainsExpression) function to check if the given string contains `${{ }}` placeholders or not. (thanks [@hugo-syn](https://github.com/hugo-syn), [#388](https://github.com/rhysd/actionlint/issues/388))
- Support Go 1.22 and set the minimum supported Go version to 1.18 for `x/sys` package.
- Update Go dependencies to the latest.

[Changes][v1.6.27]


<a id="v1.6.26"></a>
# [v1.6.26](https://github.com/rhysd/actionlint/releases/tag/v1.6.26) - 2023-09-18

- Several template fields and template actions were added. All fields and actions are listed in [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format-error-messages). Please read it for more details. ([#311](https://github.com/rhysd/actionlint/issues/311))
  - By these additions, now actionlint can output the result in [the SARIF format](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html). SARIF is a format for the output of static analysis tools used by [GitHub CodeQL](https://codeql.github.com/). [the example Go template](https://github.com/rhysd/actionlint/blob/main/testdata/format/sarif_template.txt) to format actionlint output in SARIF.
    ```sh
    actionlint -format "$(cat /path/to/sarif_template.txt)" > output.json
    ```
  - `allKinds` returns the kinds (lint rules) information as an array. You can include what lint rules are defined in the command output.
  - `toPascalCase` converts snake case (`foo_bar`) or kebab case (`foo-bar`) into pascal case (`FooBar`).
- Report an error when the condition at `if:` is always evaluated to true. See [the check document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#if-cond-always-true) to know more details. ([#272](https://github.com/rhysd/actionlint/issues/272))
  ```yaml
  # ERROR: All the following `if:` conditions are always evaluated to true
  - run: echo 'Commit is pushed'
    if: |
      ${{ github.event_name == 'push' }}
  - run: echo 'Commit is pushed'
    if: "${{ github.event_name == 'push' }} "
  - run: echo 'Commit is pushed to main'
    if: ${{ github.event_name == 'push' }} && ${{ github.ref_name == 'main' }}
  ```
- Fix actionlint didn't understand `${{ }}` placeholders in environment variable names. ([#312](https://github.com/rhysd/actionlint/issues/312))
  ```yaml
  env:
    "${{ steps.x.outputs.value }}": "..."
  ```
- Fix type of matrix row when some expression is assigned to it with `${{ }}` ([#285](https://github.com/rhysd/actionlint/issues/285))
  ```yaml
  strategy:
    matrix:
      test:
        # Matrix rows are assigned from JSON string
        - ${{ fromJson(inputs.matrix) }}
  steps:
    - run: echo ${{ matrix.test.foo.bar }}
  ```
- Fix checking `exclude` of matrix was incorrect when some matrix row is dynamically constructed with `${{ }}`. ([#261](https://github.com/rhysd/actionlint/issues/261))
  ```yaml
  strategy:
    matrix:
      build-type:
        - debug
        - ${{ fromJson(inputs.custom-build-type) }}
      exclude:
        # 'release' is not listed in 'build-type' row, but it should not be reported as error
        # since the second row of 'build-type' is dynamically constructed with ${{ }}.
        - build-type: release
  ```
- Fix checking `exclude` of matrix was incorrect when object is nested at row of the matrix. ([#249](https://github.com/rhysd/actionlint/issues/249))
  ```yaml
  matrix:
    os:
      - name: Ubuntu
        matrix: ubuntu
      - name: Windows
        matrix: windows
    arch:
      - name: ARM
        matrix: arm
      - name: Intel
        matrix: intel
    exclude:
      # This should exclude { os: { name: Windows, matrix: windows }, arch: {name: ARM, matrix: arm } }
      - os:
          matrix: windows
        arch:
          matrix: arm
  ```
- Fix data race when `actionlint.yml` config file is used by multiple goroutines to check multiple workflow files. ([#333](https://github.com/rhysd/actionlint/issues/333))
- Check keys' case sensitivity. ([#302](https://github.com/rhysd/actionlint/issues/302))
  ```yaml
  steps:
    # ERROR: 'run:' is correct
    - ruN: echo "hello"
  ```
- Add `number` as [input type of `workflow_dispatch` event](https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context). ([#316](https://github.com/rhysd/actionlint/issues/316))
- Check max number of inputs of `workflow_dispatch` event is 10.
- Check numbers at `timeout-minutes` and `max-parallel` are greater than zero.
- Add Go APIs to define a custom rule. Please read [the code example](https://pkg.go.dev/github.com/rhysd/actionlint/#example_Linter_yourOwnRule) to know the usage.
  - Make some [`RuleBase`](https://pkg.go.dev/github.com/rhysd/actionlint#RuleBase) methods public which are useful to implement your own custom rule type. (thanks [@hugo-syn](https://github.com/hugo-syn), [#327](https://github.com/rhysd/actionlint/issues/327), [#331](https://github.com/rhysd/actionlint/issues/331))
  - `OnRulesCreated` field is added to [`LinterOptions`](https://pkg.go.dev/github.com/rhysd/actionlint#LinterOptions) struct. You can modify applied rules with the hook (add your own rule, remove some rule, ...).
- Add `NewProject()` Go API to create a [`Project`](https://pkg.go.dev/github.com/rhysd/actionlint#Project) instance.
- Fix tests failed when sources are downloaded from `.tar.gz` link. ([#307](https://github.com/rhysd/actionlint/issues/307))
- Improve [the pre-commit document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pre-commit) to explain all pre-commit hooks by this repository.
- Clarify the regular expression syntax of `-ignore` option is [RE2](https://github.com/google/re2/wiki/Syntax). ([#320](https://github.com/rhysd/actionlint/issues/320))
- Use ubuntu-latest runner to create winget release. (thanks [@sitiom](https://github.com/sitiom), [#308](https://github.com/rhysd/actionlint/issues/308))
- Update popular actions data set, available contexts, webhook types to the latest.
  - Fix typo in `watch` webhook's types (thanks [@suzuki-shunsuke](https://github.com/suzuki-shunsuke), [#334](https://github.com/rhysd/actionlint/issues/334))
  - Add `secret_source` property to [`github` context](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context). (thanks [@asml-mdroogle](https://github.com/asml-mdroogle), [#339](https://github.com/rhysd/actionlint/issues/339))
  - Many new major releases are added to the popular actions data set (including `actions/checkout@v4`).
- Use Go 1.21 to build release binaries.
- Update Go dependencies to the latest. (thanks [@harryzcy](https://github.com/harryzcy), [#322](https://github.com/rhysd/actionlint/issues/322))

[Changes][v1.6.26]


<a id="v1.6.25"></a>
# [v1.6.25](https://github.com/rhysd/actionlint/releases/tag/v1.6.25) - 2023-06-15

- Parse new syntax at `runs-on:`. Now `runs-on:` can have `group:` and `labels:` configurations. Please read [the official document](https://docs.github.com/en/actions/using-github-hosted-runners/using-larger-runners#running-jobs-on-your-runner) for more details. ([#280](https://github.com/rhysd/actionlint/issues/280))
  ```yaml
  runs-on:
    group: ubuntu-runners
    labels: ubuntu-20.04-16core
  ```
- Add support for macOS XL runners. `macos-latest-xl`, `macos-13-xl`, `macos-12-xl` labels are available at `runs-on:`. ([#299](https://github.com/rhysd/actionlint/issues/299), thanks [@woa7](https://github.com/woa7))
- Find Git project directory from `-stdin-filename` command line argument. Even if the workflow content is passed via stdin, actionlint can recognize reusable workflows depended by the workflow using file path passed at `-stdin-filename` argument. ([#283](https://github.com/rhysd/actionlint/issues/283))
- Fix order of errors is not deterministic when multiple errors happen at the same location (file name, line number, column number). It happens only when building actionlint with Go 1.20 or later.
- Fix type name of `watch` webhook.
- Fix type of matrix row (property of `matrix` context) when `${{ }}` is used in the row value. ([#294](https://github.com/rhysd/actionlint/issues/294))
- Fix `go install ./...` doesn't work. ([#297](https://github.com/rhysd/actionlint/issues/297))
- Update `actionlint` pre-commit hook to use Go toolchain. Now pre-commit automatically installs `actionlint` command so you don't need to install it manually. Note that this hook requires pre-commit v3.0.0 or later. For those who don't have Go toolchain, the previous hook is maintained as `actionlint-system` hook. Please read [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pre-commit) to know the usage details. ([#301](https://github.com/rhysd/actionlint/issues/301), thanks [@Freed-Wu](https://github.com/Freed-Wu) and [@dokempf](https://github.com/dokempf))
- Update Go dependencies to the latest.
- Update npm dependencies for playground to the latest and fix optimizing Wasm binary with `wasm-opt`.
- Update popular actions data set. New major versions and new inputs of many popular actions are now supported like `sparse-checkout` input of `actions/checkout` action.  ([#305](https://github.com/rhysd/actionlint/issues/305))
- Fix outdated document for Problem Matchers. ([#289](https://github.com/rhysd/actionlint/issues/289), thanks [@carlcsaposs-canonical](https://github.com/carlcsaposs-canonical))
- Fix outdated links in document for super-linter. ([#303](https://github.com/rhysd/actionlint/issues/303), thanks [@gmacario](https://github.com/gmacario))
- Automate releasing the Winget package with GitHub Actions. ([#276](https://github.com/rhysd/actionlint/issues/276), [#293](https://github.com/rhysd/actionlint/issues/293), thanks [@sitiom](https://github.com/sitiom))

[Changes][v1.6.25]


<a id="v1.6.24"></a>
# [v1.6.24](https://github.com/rhysd/actionlint/releases/tag/v1.6.24) - 2023-04-04

- Add support for [configuration variables](https://docs.github.com/en/actions/learn-github-actions/variables). However actionlint doesn't know what variables are defined in the repository on GitHub. To notify them, [you need to configure your variables in your repository](https://github.com/rhysd/actionlint/blob/main/docs/config.md).
  ```yaml
  config-variables:
    - DEFAULT_RUNNER
    - DEFAULT_TIMEOUT
  ```
- Fix type error when `inputs` context is shared by multiple events. ([#263](https://github.com/rhysd/actionlint/issues/263))
- Add document for [how to install actionlint with winget](https://github.com/rhysd/actionlint/blob/main/docs/install.md#winget). ([#267](https://github.com/rhysd/actionlint/issues/267), thanks [@sitiom](https://github.com/sitiom))
- Add document for [how to integrate actionlint to trunk.io](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#trunk). ([#269](https://github.com/rhysd/actionlint/issues/269), thanks [@dapirian](https://github.com/dapirian))
- Add document for [how to install actionlint with Nix package manager](https://github.com/rhysd/actionlint/blob/main/docs/install.md#nix). ([#273](https://github.com/rhysd/actionlint/issues/273), thanks [@diohabara](https://github.com/diohabara))
- Update popular actions data set to the latest
- Add support for Go 1.20 and build release binaries with Go 1.20


[Changes][v1.6.24]


<a id="v1.6.23"></a>
# [v1.6.23](https://github.com/rhysd/actionlint/releases/tag/v1.6.23) - 2023-01-19

- Fix using [`vars` context](https://docs.github.com/en/actions/learn-github-actions/contexts#vars-context) causes 'undefined context' error. This context is for ['Variables' feature](https://docs.github.com/en/actions/learn-github-actions/variables) which was recently added to GitHub Actions. ([#260](https://github.com/rhysd/actionlint/issues/260))
  ```yaml
  - name: Use variables
    run: |
      echo "repository variable : ${{ vars.REPOSITORY_VAR }}"
      echo "organization variable : ${{ vars.ORGANIZATION_VAR }}"
      echo "overridden variable : ${{ vars.OVERRIDE_VAR }}"
      echo "variable from shell environment : $env_var"
  ```
- Fix 'no property' error on accessing some `github` context's properties which were added recently. ([#259](https://github.com/rhysd/actionlint/issues/259))
- Update popular actions data set and add some new actions to it
  - [actions/dependency-review-action](https://github.com/actions/dependency-review-action)
  - [dtolnay/rust-toolchain](https://github.com/dtolnay/rust-toolchain)
- Playground is improved by making the right pane sticky. It is useful when many errors are reported. ([#253](https://github.com/rhysd/actionlint/issues/253), thanks [@ericcornelissen](https://github.com/ericcornelissen))
- Update Go modules dependencies and playground dependencies

[Changes][v1.6.23]


<a id="v1.6.22"></a>
# [v1.6.22](https://github.com/rhysd/actionlint/releases/tag/v1.6.22) - 2022-11-01

- Detect deprecated workflow commands such as [`set-output` or `save-state`](https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/) and suggest the alternative. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-deprecated-workflow-commands) for more details. ([#234](https://github.com/rhysd/actionlint/issues/234))
  ```yaml
  # ERROR: This format of 'set-output' workflow command was deprecated
  - run: echo '::set-output name=foo::bar'
  ```
- Fix that `${{ }}` expression at `on.workflow_call.inputs.<id>.default` caused an error. ([#235](https://github.com/rhysd/actionlint/issues/235))
  ```yaml
  on:
    workflow_call:
      inputs:
        project:
          type: string
          # OK: The default value is generated dynamically
          default: ${{ github.event.repository.name }}
  ```
- Improve type of `inputs` context to grow gradually while checking inputs in `workflow_call` event.
  ```yaml
  on:
    workflow_call:
      inputs:
        input1:
          type: string
          # ERROR: `input2` is not defined yet
          default: ${{ inputs.input2 }}
        input2:
          type: string
          # OK: `input1` was already defined above
          default: ${{ inputs.input1 }}
  ```
- Check types of default values of workflow call inputs even if `${{ }}` expression is used.
  ```yaml
  on:
    workflow_call:
      inputs:
        input1:
          type: boolean
        input2:
          type: number
          # ERROR: Boolean value cannot be assigned to number
          default: ${{ inputs.input1 }}
  ```
- Fix the download script is broken since GHE server does not support the new `set-output` format yet. ([#240](https://github.com/rhysd/actionlint/issues/240))
- Replace the deprecated `set-output` workflow command in our own workflows. ([#239](https://github.com/rhysd/actionlint/issues/239), thanks [@Mrtenz](https://github.com/Mrtenz))
- Popular actions data set was updated to the latest as usual.

[Changes][v1.6.22]


<a id="v1.6.21"></a>
# [v1.6.21](https://github.com/rhysd/actionlint/releases/tag/v1.6.21) - 2022-10-09

- [Check contexts availability](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#ctx-spfunc-availability). Some contexts limit where they can be used. For example, `jobs.<job_id>.env` workflow key does not allow accessing `env` context, but `jobs.<job_id>.steps.env` allows. See [the official document](https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability) for the complete list of contexts availability. ([#180](https://github.com/rhysd/actionlint/issues/180))
  ```yaml
  ...

  env:
    TOPLEVEL: ...

  jobs:
    test:
      runs-on: ubuntu-latest
      env:
        # ERROR: 'env' context is not available here
        JOB_LEVEL: ${{ env.TOPLEVEL }}
      steps:
        - env:
            # OK: 'env' context is available here
            STEP_LEVEL: ${{ env.TOPLEVEL }}
          ...
  ```
  actionlint reports the context is not available and what contexts are available as follows:
  ```
  test.yaml:11:22: context "env" is not allowed here. available contexts are "github", "inputs", "matrix", "needs", "secrets", "strategy". see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details [expression]
     |
  11 |       JOB_LEVEL: ${{ env.TOPLEVEL }}
     |                      ^~~~~~~~~~~~
  ```
- [Check special functions availability](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#ctx-spfunc-availability). Some functions limit where they can be used. For example, status functions like `success()` or `failure()` are only available in conditions of `if:`. See [the official document](https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability) for the complete list of special functions availability. ([#214](https://github.com/rhysd/actionlint/issues/214))
  ```yaml
  ...

  steps:
    # ERROR: 'success()' function is not available here
    - run: echo 'Success? ${{ success() }}'
      # OK: 'success()' function is available here
      if: success()
  ```
  actionlint reports `success()` is not available and where the function is available as follows:
  ```
  test.yaml:8:33: calling function "success" is not allowed here. "success" is only available in "jobs.<job_id>.if", "jobs.<job_id>.steps.if". see https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability for more details [expression]
    |
  8 |       - run: echo 'Success? ${{ success() }}'
    |                                 ^~~~~~~~~
  ```
- Fix `inputs` context is not available in `run-name:` section. ([#223](https://github.com/rhysd/actionlint/issues/223))
- Allow dynamic shell configuration like `shell: ${{ env.SHELL }}`.
- Fix no error is reported when `on:` does not exist at toplevel. ([#232](https://github.com/rhysd/actionlint/issues/232))
- Fix an error position is not correct when the error happens at root node of workflow AST.
- Fix an incorrect empty event is parsed when `on:` section is empty.
- Fix the error message when parsing an unexpected key on toplevel. (thanks [@norwd](https://github.com/norwd), [#231](https://github.com/rhysd/actionlint/issues/231))
- Add `in_progress` type to `workflow_run` webhook event trigger.
- Describe [the actionlint extension](https://extensions.panic.com/extensions/org.netwrk/org.netwrk.actionlint/) for [Nova.app](https://nova.app) in [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#nova). (thanks [@jbergstroem](https://github.com/jbergstroem), [#222](https://github.com/rhysd/actionlint/issues/222))
- Note [Super-Linter](https://github.com/github/super-linter) uses a different place for configuration file. (thanks [@per-oestergaard](https://github.com/per-oestergaard), [#227](https://github.com/rhysd/actionlint/issues/227))
- Add `actions/setup-dotnet@v3` to popular actions data set.
- [`generate-availability` script](https://github.com/rhysd/actionlint/tree/main/scripts/generate-availability) was created to scrape the information about contexts and special functions availability from the official document. The information can be used through [`actionlint.WorkflowKeyAvailability()`](https://pkg.go.dev/github.com/rhysd/actionlint#WorkflowKeyAvailability) Go API. This script is run once a week on CI to keep the information up-to-date.



[Changes][v1.6.21]


<a id="v1.6.20"></a>
# [v1.6.20](https://github.com/rhysd/actionlint/releases/tag/v1.6.20) - 2022-09-30

- Support `run-name` which [GitHub introduced recently](https://github.blog/changelog/2022-09-26-github-actions-dynamic-names-for-workflow-runs/). It is a name of workflow run dynamically configured. See [the official document](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#run-name) for more details. ([#220](https://github.com/rhysd/actionlint/issues/220))
  ```yaml
  on: push
  run-name: Deploy by @${{ github.actor }}

  jobs:
    ...
  ```
- Add `end_column` property to JSON representation of error. The property indicates a column of the end position of `^~~~~~~` indicator in snippet. Note that `end_column` is equal to `column` when the indicator cannot be shown. ([#219](https://github.com/rhysd/actionlint/issues/219))
  ```console
  $ actionlint -format '{{json .}}' test.yaml | jq
  [
    {
      "message": "property \"unknown_prop\" is not defined in object type {arch: string; debug: string; name: string; os: string; temp: string; tool_cache: string; workspace: string}",
      "filepath": "test.yaml",
      "line": 7,
      "column": 23,
      "kind": "expression",
      "snippet": "      - run: echo ${{ runner.unknown_prop }}\n                      ^~~~~~~~~~~~~~~~~~~",
      "end_column": 41
    }
  ]
  ```
- Overhaul the workflow parser to parse workflow keys in case-insensitive. This is a work derived from the fix of [#216](https://github.com/rhysd/actionlint/issues/216). Now the parser parses all workflow keys in case-insensitive way correctly. Note that permission names at `permissions:` are exceptionally case-sensitive.
  - This fixes properties of `inputs` for `workflow_dispatch` were not case-insensitive.
  - This fixes inputs and outputs of local actions were not handled in case-insensitive way.
- Update popular actions data set. `actions/stale@v6` was newly added.

[Changes][v1.6.20]


<a id="v1.6.19"></a>
# [v1.6.19](https://github.com/rhysd/actionlint/releases/tag/v1.6.19) - 2022-09-22

- Fix inputs, outputs, and secrets of reusable workflow should be case-insensitive. ([#216](https://github.com/rhysd/actionlint/issues/216))
  ```yaml
  # .github/workflows/reusable.yaml
  on:
    workflow_call:
      inputs:
        INPUT_UPPER:
          type: string
        input_lower:
          type: string
      secrets:
        SECRET_UPPER:
        secret_lower:
  ...

  # .github/workflows/test.yaml
  ...

  jobs:
    caller:
      uses: ./.github/workflows/reusable.yaml
      # Inputs and secrets are case-insensitive. So all the followings should be OK
      with:
        input_upper: ...
        INPUT_LOWER: ...
      secrets:
        secret_upper: ...
        SECRET_LOWER: ...
  ```
- Describe [how to install specific version of `actionlint` binary with the download script](https://github.com/rhysd/actionlint/blob/main/docs/install.md#download-script). ([#218](https://github.com/rhysd/actionlint/issues/218))

[Changes][v1.6.19]


<a id="v1.6.18"></a>
# [v1.6.18](https://github.com/rhysd/actionlint/releases/tag/v1.6.18) - 2022-09-17

- This release much enhances checks for local reusable workflow calls. Note that these checks are done for local reusable workflows (starting with `./`). ([#179](https://github.com/rhysd/actionlint/issues/179)).
  - Detect missing required inputs/secrets and undefined inputs/secrets at `jobs.<job_id>.with` and `jobs.<job_id>.secrets`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-inputs-and-secrets-in-workflow-call) for more details.
    ```yaml
    # .github/workflows/reusable.yml
    on:
      workflow_call:
        inputs:
          name:
            type: string
            required: true
        secrets:
          password:
            required: true
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      missing-required:
        uses: ./.github/workflows/reusable.yml
        with:
          # ERROR: Undefined input "user"
          user: rhysd
          # ERROR: Required input "name" is missing
        secrets:
          # ERROR: Undefined secret "credentials"
          credentials: my-token
          # ERROR: Required secret "password" is missing
    ```
  - Type check for reusable workflow inputs at `jobs.<job_id>.with`. Types are defined at `on.workflow_call.inputs.<name>.type` in reusable workflow. actionlint checks types of expressions in workflow calls. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-inputs-and-secrets-in-workflow-call) for more details.
    ```yaml
    # .github/workflows/reusable.yml
    on:
      workflow_call:
        inputs:
          id:
            type: number
          message:
            type: string
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      type-checks:
        uses: ./.github/workflows/reusable.yml
        with:
          # ERROR: Cannot assign string value to number input. format() returns string value
          id: ${{ format('runner name is {0}', runner.name) }}
          # ERROR: Cannot assign null to string input. If you want to pass string "null", use ${{ 'null' }}
          message: null
    ```
  - Detect local reusable workflow which does not exist at `jobs.<job_id>.uses`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-workflow-call-syntax) for more details.
    ```yaml
    jobs:
      test:
        # ERROR: This workflow file does not exist
        with: ./.github/workflows/does-not-exist.yml
    ```
  - Check `needs.<job_id>.outputs.<output_id>` in downstream jobs of workflow call jobs. The outputs object is now typed strictly based on `on.workflow_call.outputs.<name>` in the called reusable workflow. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-outputs-of-workflow-call-in-downstream-jobs) for more details.
    ```yaml
    # .github/workflows/get-build-info.yml
    on:
      workflow_call:
        outputs:
          version:
            value: ...
            description: version of software
    ...

    # .github/workflows/test.yml
    ...

    jobs:
      # This job's outputs object is typed as {version: string}
      get_build_info:
        uses: ./.github/workflows/get-build-info.yml
      downstream:
        needs: [get_build_info]
        runs-on: ubuntu-latest
        steps:
          # OK. `version` is defined in the reusable workflow
          - run: echo '${{ needs.get_build_info.outputs.version }}'
          # ERROR: `tag` is not defined in the reusable workflow
          - run: echo '${{ needs.get_build_info.outputs.tag }}'
    ```
- Add missing properties in contexts and improve types of some properties looking at [the official contexts document](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context).
  - `github.action_status`
  - `runner.debug`
  - `services.<service_id>.ports`
- Fix `on.workflow_call.inputs.<name>.description` and `on.workflow_call.secrets.<name>.description` were incorrectly mandatory. They are actually optional.
- Report parse errors when parsing `action.yml` in local actions. They were ignored in previous versions.
- Sort the order of properties in an object type displayed in error message. In previous versions, actionlint sometimes displayed `{a: true, b: string}`, or it displayed `{b: string, a: true}` for the same object type. This randomness was caused by random iteration of map values in Go.
- Update popular actions data set to the latest.

[Changes][v1.6.18]


<a id="v1.6.17"></a>
# [v1.6.17](https://github.com/rhysd/actionlint/releases/tag/v1.6.17) - 2022-08-28

- Allow workflow calls are available in matrix jobs. See [the official announcement](https://github.blog/changelog/2022-08-22-github-actions-improvements-to-reusable-workflows-2/) for more details. ([#197](https://github.com/rhysd/actionlint/issues/197))
  ```yaml
  jobs:
    ReusableMatrixJobForDeployment:
      strategy:
        matrix:
          target: [dev, stage, prod]
      uses: octocat/octo-repo/.github/workflows/deployment.yml@main
      with:
        target: ${{ matrix.target }}
  ```
- Allow nested workflow calls. See [the official announcement](https://github.blog/changelog/2022-08-22-github-actions-improvements-to-reusable-workflows-2/) for more details. ([#201](https://github.com/rhysd/actionlint/issues/201))
  ```yaml
  on: workflow_call

  jobs:
    call-another-reusable:
      uses: path/to/another-reusable.yml@v1
  ```
- Fix job outputs should be passed to `needs.*.outputs` of only direct children. Until v1.6.16, they are passed to any downstream jobs. ([#151](https://github.com/rhysd/actionlint/issues/151))
  ```yaml
  jobs:
    first:
      runs-on: ubuntu-latest
      outputs:
        first: 'output from first job'
      steps:
        - run: echo 'first'

    second:
      needs: [first]
      runs-on: ubuntu-latest
      outputs:
        second: 'output from second job'
      steps:
        - run: echo 'second'

    third:
      needs: [second]
      runs-on: ubuntu-latest
      steps:
        - run: echo '${{ toJSON(needs.second.outputs) }}'
        # ERROR: `needs.first` does not exist, but v1.6.16 reported no error
        - run: echo '${{ toJSON(needs.first.outputs) }}'
  ```
  When you need both `needs.first` and `needs.second`, add the both to `needs:`.
  ```yaml
    third:
      needs: [first, second]
      runs-on: ubuntu-latest
      steps:
        # OK
        -  echo '${{ toJSON(needs.first.outputs) }}'
  ```
- Fix `}}` in string literals are detected as end marker of placeholder `${{ }}`. ([#205](https://github.com/rhysd/actionlint/issues/205))
  ```yaml
  jobs:
    test:
      runs-on: ubuntu-latest
      strategy:
        # This caused an incorrect error until v1.6.16
        matrix: ${{ fromJSON('{"foo":{}}') }}
  ```
- Fix `working-directory:` should not be available with `uses:` in steps. `working-directory:` is only available with `run:`. ([#207](https://github.com/rhysd/actionlint/issues/207))
  ```yaml
  steps:
    - uses: actions/checkout@v3
      # ERROR: `working-directory:` is not available here
      working-directory: ./foo
  ```
- The working directory for running `actionlint` command can be set via [`WorkingDir` field of `LinterOptions` struct](https://pkg.go.dev/github.com/rhysd/actionlint#LinterOptions). When it is empty, the return value from `os.Getwd` will be used.
- Update popular actions data set. `actions/configure-pages@v2` was added.
- Use Go 1.19 on CI by default. It is used to build release binaries.
- Update dependencies (go-yaml/yaml v3.0.1).
- Update playground dependencies (except for CodeMirror v6).

[Changes][v1.6.17]


<a id="v1.6.16"></a>
# [v1.6.16](https://github.com/rhysd/actionlint/releases/tag/v1.6.16) - 2022-08-19

- Allow an empty object at `permissions:`. You can use it to disable permissions for all of the available scopes. ([#170](https://github.com/rhysd/actionlint/issues/170), [#171](https://github.com/rhysd/actionlint/issues/171), thanks [@peaceiris](https://github.com/peaceiris))
  ```yaml
  permissions: {}
  ```
- Support `github.triggering_actor` context value. ([#190](https://github.com/rhysd/actionlint/issues/190), thanks [@stefreak](https://github.com/stefreak))
- Rename `step-id` rule to `id` rule. Now the rule checks both job IDs and step IDs. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#id-naming-convention) for more details. ([#182](https://github.com/rhysd/actionlint/issues/182))
  ```yaml
  jobs:
    # ERROR: '.' cannot be contained in ID
    v1.2.3:
      runs-on: ubuntu-latest
      steps:
        - run: echo 'job ID with version'
          # ERROR: ID cannot contain spaces
          id: echo for test
    # ERROR: ID cannot start with numbers
    2d-game:
      runs-on: ubuntu-latest
      steps:
        - run: echo 'oops'
  ```
- Accessing `env` context in `jobs.<id>.if` is now reported as error. ([#155](https://github.com/rhysd/actionlint/issues/155))
  ```yaml
  jobs:
    test:
      runs-on: ubuntu-latest
      # ERROR: `env` is not available here
      if: ${{ env.DIST == 'arch' }}
      steps:
        - run: ...
  ```
- Fix actionlint wrongly typed some matrix value when the matrix is expanded with `${{ }}`. For example, `matrix.foo` in the following code is typed as `{x: string}`, but it should be `any` because it is initialized with the value from `fromJSON`. ([#145](https://github.com/rhysd/actionlint/issues/145))
  ```yaml
  strategy:
    matrix:
      foo: ${{ fromJSON(...) }}
      exclude:
        - foo:
            x: y
  ```
- Fix incorrect type check when multiple runner labels are set to `runs-on:` via expanding `${{ }}` for selecting self-hosted runners. ([#164](https://github.com/rhysd/actionlint/issues/164))
  ```yaml
  jobs:
    test:
      strategy:
        matrix:
          include:
            - labels: ["self-hosted", "macOS", "X64"]
            - labels: ["self-hosted", "linux"]
      # actionlint incorrectly reported type error here
      runs-on: ${{ matrix.labels }}
  ```
- Fix usage of local actions (`uses: ./path/to/action`) was not checked when multiple workflow files were passed to `actionlint` command. ([#173](https://github.com/rhysd/actionlint/issues/173))
- Allow `description:` is missing in `secrets:` of reusable workflow call definition since it is optional. ([#174](https://github.com/rhysd/actionlint/issues/174))
- Fix type of property of `github.event.inputs` is string unlike `inputs` context. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#workflow-dispatch-event-validation) for more details. ([#181](https://github.com/rhysd/actionlint/issues/181))
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        is-valid:
          # Type of `inputs.is-valid` is bool
          # Type of `github.event.inputs.is-valid` is string
          type: boolean
  ```
- Fix crash when a value is expanded with `${{ }}` at `continue-on-error:`. ([#193](https://github.com/rhysd/actionlint/issues/193))
- Fix some error was caused by some other error. For example, the following code reported two errors. '" is not available for string literal' error caused another 'one placeholder should be included in boolean value string' error. This was caused because the `${{ x == "foo" }}` placeholder was not counted due to the previous type error.
  ```yaml
  if: ${{ x == "foo" }}
  ```
- Add support for [`merge_group` workflow trigger](https://github.blog/changelog/2022-08-18-merge-group-webhook-event-and-github-actions-workflow-trigger/).
- Add official actions to manage GitHub Pages to popular actions data set.
  - `actions/configure-pages@v1`
  - `actions/deploy-pages@v1`
  - `actions/upload-pages-artifact@v1`
- Update popular actions data set to the latest. Several new major versions and new inputs of actions were added to it.
- Describe how to install actionlint via [Chocolatey](https://chocolatey.org/), [scoop](https://scoop.sh/), and [AUR](https://aur.archlinux.org/) in [the installation document](https://github.com/rhysd/actionlint/blob/main/docs/install.md). ([#167](https://github.com/rhysd/actionlint/issues/167), [#168](https://github.com/rhysd/actionlint/issues/168), thanks [@sitiom](https://github.com/sitiom))
- [VS Code extension for actionlint](https://marketplace.visualstudio.com/items?itemName=arahata.linter-actionlint) was created by [@arahatashun](https://github.com/arahatashun). See [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#vs-code) for more details.
- Describe how to use [the Docker image](https://hub.docker.com/r/rhysd/actionlint) at step of GitHub Actions workflow. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#use-actionlint-on-github-actions) for the details. ([#146](https://github.com/rhysd/actionlint/issues/146))
  ```yaml
  - uses: docker://rhysd/actionlint:latest
    with:
      args: -color
  ```
- Clarify the behavior if empty strings are set to some command line options in documents. `-shellcheck=` disables shellcheck integration and `-pyflakes=` disables pyflakes integration. ([#156](https://github.com/rhysd/actionlint/issues/156))
- Update Go module dependencies.

[Changes][v1.6.16]


<a id="v1.6.15"></a>
# [v1.6.15](https://github.com/rhysd/actionlint/releases/tag/v1.6.15) - 2022-06-28

- Fix referring `env` context from `env:` at step level caused an error. `env:` at toplevel and job level cannot refer `env` context, but `env:` at step level can. ([#158](https://github.com/rhysd/actionlint/issues/158))
  ```yaml
  on: push

  env:
    # ERROR: 'env:' at toplevel cannot refer 'env' context
    ERROR1: ${{ env.PATH }}

  jobs:
    my_job:
      runs-on: ubuntu-latest
      env:
        # ERROR: 'env:' at job level cannot refer 'env' context
        ERROR2: ${{ env.PATH }}
      steps:
        - run: echo "$THIS_IS_OK"
          env:
            # OK: 'env:' at step level CAN refer 'env' context
            THIS_IS_OK: ${{ env.PATH }}
  ```
- [Docker image for linux/arm64](https://hub.docker.com/layers/rhysd/actionlint/1.6.15/images/sha256-f63ee59f1846abce86ca9de1d41a1fc22bc7148d14b788cb455a9594d83e73f7?context=repo) is now provided. It is useful for M1 Mac users. ([#159](https://github.com/rhysd/actionlint/issues/159), thanks [@politician](https://github.com/politician))
- Fix [the download script](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) did not respect the version specified via the first argument. ([#162](https://github.com/rhysd/actionlint/issues/162), thanks [@mateiidavid](https://github.com/mateiidavid))

[Changes][v1.6.15]


<a id="v1.6.14"></a>
# [v1.6.14](https://github.com/rhysd/actionlint/releases/tag/v1.6.14) - 2022-06-26

- Some filters are exclusive in events at `on:`. Now actionlint checks the exclusive filters are used in the same event. `paths` and `paths-ignore`, `branches` and `branches-ignore`, `tags` and `tags-ignore` are exclusive. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#webhook-events-validation) for the details.
  ```yaml
  on:
    push:
      # ERROR: Both 'paths' and 'paths-ignore' filters cannot be used for the same event
      paths: ...
      paths-ignore: ...
  ```
- Some event filters are checked more strictly. Some filters are only available with specific events. Now actionlint checks the limitation. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#webhook-events-validation) for complete list of such filters.
  ```yaml
  on:
    release:
      # ERROR: 'tags' filter is only available for 'push' event
      tags: v*.*.*
  ```
- Paths starting/ending with spaces are now reported as error.
- Inputs of workflow which specify both `default` and `required` are now reported as error. When `required` is specified at input of workflow call, a caller of it must specify value of the input. So the default value will never be used. ([#154](https://github.com/rhysd/actionlint/issues/154), thanks [@sksat](https://github.com/sksat))
  ```yaml
  on:
    workflow_call:
      inputs:
        my_input:
          description: test
          type: string
          # ERROR: The default value 'aaa' will never be used
          required: true
          default: aaa
  ```
- Fix inputs of `workflow_dispatch` are set to `inputs` context as well as `github.event.inputs`. This was added by [the recent change of GitHub Actions](https://github.blog/changelog/2022-06-10-github-actions-inputs-unified-across-manual-and-reusable-workflows/). ([#152](https://github.com/rhysd/actionlint/issues/152))
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        my_input:
          type: string
          required: true
  jobs:
    my_job:
      runs-on: ubuntu-latest
      steps:
        - run: echo ${{ github.event.inputs.my_input }}
        # Now the input is also set to `inputs` context
        - run: echo ${{ inputs.my_input }}
  ```
- Improve that `env` context is now not defined in values of `env:`, `id:` and `uses:`. actionlint now reports usage of `env` context in such places as type errors. ([#158](https://github.com/rhysd/actionlint/issues/158))
  ```yaml
  runs-on: ubuntu-latest
  env:
    FOO: aaa
  steps:
    # ERROR: 'env' context is not defined in values of 'env:', 'id:' and 'uses:'
    - uses: test/${{ env.FOO }}@main
      env:
        BAR: ${{ env.FOO }}
      id: foo-${{ env.FOO }}
  ```
- `actionlint` command gains `-stdin-filename` command line option. When it is specified, the file name is used on reading input from stdin instead of `<stdin>`. ([#157](https://github.com/rhysd/actionlint/issues/157), thanks [@arahatashun](https://github.com/arahatashun))
  ```sh
  # Error message shows foo.yml as file name where the error happened
  ... | actionlint -stdin-filename foo.yml -
  ```
- [The download script](https://github.com/rhysd/actionlint/blob/main/docs/install.md#download-script) allows to specify a directory path to install `actionlint` executable with the second argument of the script. For example, the following command downloads `/path/to/bin/actionlint`:
  ```sh
  # Downloads the latest stable version at `/path/to/bin/actionlint`
  bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash) latest /path/to/bin
  # Downloads actionlint v1.6.14 at `/path/to/bin/actionlint`
  bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash) 1.6.14 /path/to/bin
  ```
- Update popular actions data set including `goreleaser-action@v3`, `setup-python@v4`, `aks-set-context@v3`.
- Update Go dependencies including go-yaml/yaml v3.

[Changes][v1.6.14]


<a id="v1.6.13"></a>
# [v1.6.13](https://github.com/rhysd/actionlint/releases/tag/v1.6.13) - 2022-05-18

- [`secrets: inherit` in reusable workflow](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretsinherit) is now supported ([#138](https://github.com/rhysd/actionlint/issues/138))
  ```yaml
  on:
    workflow_dispatch:

  jobs:
    pass-secrets-to-workflow:
      uses: ./.github/workflows/called-workflow.yml
      secrets: inherit
  ```
  This means that actionlint cannot know the workflow inherits secrets or not when checking a reusable workflow. To support `secrets: inherit` without giving up on checking `secrets` context, actionlint assumes the followings. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-types-of-inputs-and-secrets-in-reusable-workflow) for the details.
  - when `secrets:` is omitted in a reusable workflow, the workflow inherits secrets from a caller
  - when `secrets:` exists in a reusable workflow, the workflow inherits no other secret
- [`macos-12` runner](https://github.blog/changelog/2022-04-25-github-actions-public-beta-of-macos-12-for-github-hosted-runners-is-now-available/) is now supported ([#134](https://github.com/rhysd/actionlint/issues/134), thanks [@shogo82148](https://github.com/shogo82148))
- [`ubuntu-22.04` runner](https://github.blog/changelog/2022-05-10-github-actions-beta-of-ubuntu-22-04-for-github-hosted-runners-is-now-available/) is now supported ([#142](https://github.com/rhysd/actionlint/issues/142), thanks [@shogo82148](https://github.com/shogo82148))
- `concurrency` is available on reusable workflow call ([#136](https://github.com/rhysd/actionlint/issues/136))
  ```yaml
  jobs:
    checks:
      concurrency:
        group: ${{ github.ref }}-${{ github.workflow }}
        cancel-in-progress: true
      uses: ./path/to/workflow.yaml
  ```
- [pre-commit](https://pre-commit.com/) hook now uses a fixed version of actionlint. For example, the following configuration continues to use actionlint v1.6.13 even if v1.6.14 is released. ([#116](https://github.com/rhysd/actionlint/issues/116))
  ```yaml
  repos:
    - repo: https://github.com/rhysd/actionlint
      rev: v1.6.13
      hooks:
        - id: actionlint-docker
  ```
- Update popular actions data set including new versions of `docker/*`, `haskell/actions/setup`,  `actions/setup-go`, ... ([#140](https://github.com/rhysd/actionlint/issues/140), thanks [@bflad](https://github.com/bflad))
- Update Go module dependencies


[Changes][v1.6.13]


<a id="v1.6.12"></a>
# [v1.6.12](https://github.com/rhysd/actionlint/releases/tag/v1.6.12) - 2022-04-14

- Fix `secrets.ACTIONS_RUNNER_DEBUG` and `secrets.ACTIONS_STEP_DEBUG` are not pre-defined in a reusable workflow. ([#130](https://github.com/rhysd/actionlint/issues/130))
- Fix checking permissions is outdated. `pages` and `discussions` permissions were added and `metadata` permission was removed. ([#131](https://github.com/rhysd/actionlint/issues/131), thanks [@suzuki-shunsuke](https://github.com/suzuki-shunsuke))
- Disable [SC2157](https://github.com/koalaman/shellcheck/wiki/SC2157) shellcheck rule to avoid a false positive due to [the replacement of `${{ }}`](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#shellcheck-integration-for-run) in script. For example, in the below script `-z ${{ env.FOO }}` was replaced with `-z ______________` and it caused 'always false due to literal strings' error. ([#113](https://github.com/rhysd/actionlint/issues/113))
  ```yaml
  - run: |
      if [[ -z ${{ env.FOO }} ]]; then
        echo "FOO is empty"
      fi
  ```
- Add codecov-action@v3 to popular actions data set.

[Changes][v1.6.12]


<a id="v1.6.11"></a>
# [v1.6.11](https://github.com/rhysd/actionlint/releases/tag/v1.6.11) - 2022-04-05

- Fix crash on making [outputs in JSON format](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format-error-messages) with `actionlint -format '{{json .}}'`. ([#128](https://github.com/rhysd/actionlint/issues/128))
- Allow any outputs from `actions/github-script` action because it allows to set arbitrary outputs via calling `core.setOutput()` in JavaScript. ([#104](https://github.com/rhysd/actionlint/issues/104))
  ```yaml
  - id: test
    uses: actions/github-script@v5
    with:
      script: |
        core.setOutput('answer', 42);
  - run: |
      echo "The answer is ${{ steps.test.outputs.answer }}"
  ```
- Add support for Go 1.18. All released binaries were built with Go 1.18 compiler. The bottom supported version is Go 1.16 and it's not been changed.
- Update popular actions data set (`actions/cache`, `code-ql-actions/*`, ...)
- Update some Go module dependencies

[Changes][v1.6.11]


<a id="v1.6.10"></a>
# [v1.6.10](https://github.com/rhysd/actionlint/releases/tag/v1.6.10) - 2022-03-11

- Support outputs in reusable workflow call. See [the official document](https://docs.github.com/en/actions/using-workflows/reusing-workflows#using-outputs-from-a-reusable-workflow) for the usage of the outputs syntax. ([#119](https://github.com/rhysd/actionlint/issues/119), [#121](https://github.com/rhysd/actionlint/issues/121))
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
  Example of reusable workflow call:
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
- Check job IDs. They must start with a letter or `_` and contain only alphanumeric characters, `-` or `_`. See [the document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#job-id-naming-convention) for more details. ([#80](https://github.com/rhysd/actionlint/issues/80))
  ```yaml
  on: push
  jobs:
    # ERROR: '.' cannot be contained in job ID
    foo-v1.2.3:
      runs-on: ubuntu-latest
      steps:
        - run: 'job ID with version'
  ```
- Fix `windows-latest` now means `windows-2022` runner. See [virtual-environments#4856](https://github.com/actions/virtual-environments/issues/4856) for the details. ([#120](https://github.com/rhysd/actionlint/issues/120))
- Update [the playground](https://rhysd.github.io/actionlint/) dependencies to the latest.
- Update Go module dependencies

[Changes][v1.6.10]


<a id="v1.6.9"></a>
# [v1.6.9](https://github.com/rhysd/actionlint/releases/tag/v1.6.9) - 2022-02-24

- Support [`runner.arch` context value](https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context). (thanks [@shogo82148](https://github.com/shogo82148), [#101](https://github.com/rhysd/actionlint/issues/101))
  ```yaml
  steps:
    - run: ./do_something_64bit.sh
      if: ${{ runner.arch == 'x64' }}
  ```
- Support [calling reusable workflows in local directories](https://docs.github.com/en/actions/using-workflows/reusing-workflows#calling-a-reusable-workflow). (thanks [@jsok](https://github.com/jsok), [#107](https://github.com/rhysd/actionlint/issues/107))
  ```yaml
  jobs:
    call-workflow-in-local-repo:
      uses: ./.github/workflows/useful_workflow.yml
  ```
- Add [a document](https://github.com/rhysd/actionlint/blob/main/docs/install.md#asdf) to install actionlint via [asdf](https://asdf-vm.com/) version manager. (thanks [@crazy-matt](https://github.com/crazy-matt), [#99](https://github.com/rhysd/actionlint/issues/99))
- Fix using `secrets.GITHUB_TOKEN` caused a type error when some other secret is defined. (thanks [@mkj-is](https://github.com/mkj-is), [#106](https://github.com/rhysd/actionlint/issues/106))
- Fix nil check is missing on parsing `uses:` step. (thanks [@shogo82148](https://github.com/shogo82148), [#102](https://github.com/rhysd/actionlint/issues/102))
- Fix some documents including broken links. (thanks [@ohkinozomu](https://github.com/ohkinozomu), [#105](https://github.com/rhysd/actionlint/issues/105))
- Update popular actions data set to the latest. More arguments are added to many actions. And a few actions had new major versions.
- Update webhook payload data set to the latest. `requested_action` type was added to `check_run` hook. `requested` and `rerequested` types were removed from `check_suite` hook. `updated` type was removed from `project` hook.


[Changes][v1.6.9]


<a id="v1.6.8"></a>
# [v1.6.8](https://github.com/rhysd/actionlint/releases/tag/v1.6.8) - 2021-11-15

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
  - `github.ref_name` (thanks [@dihmandrake](https://github.com/dihmandrake), [#72](https://github.com/rhysd/actionlint/issues/72))
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


<a id="v1.6.7"></a>
# [v1.6.7](https://github.com/rhysd/actionlint/releases/tag/v1.6.7) - 2021-11-08

- Fix missing property `name` in `runner` context object (thanks [@ioanrogers](https://github.com/ioanrogers), [#67](https://github.com/rhysd/actionlint/issues/67)).
- Fix a false positive on type checking at `x.*` object filtering syntax where the receiver is an object. actionlint previously only allowed arrays as receiver of object filtering ([#66](https://github.com/rhysd/actionlint/issues/66)).
  ```ruby
  fromJSON('{"a": "from a", "b": "from b"}').*
  # => ["from a", "from b"]

  fromJSON('{"a": {"x": "from a.x"}, "b": {"x": "from b.x"}}').*.x
  # => ["from a.x", "from b.x"]
  ```
- Add [rust-cache](https://github.com/Swatinem/rust-cache) as new popular action.
- Remove `bottle: unneeded` from Homebrew formula (thanks [@oppara](https://github.com/oppara), [#63](https://github.com/rhysd/actionlint/issues/63)).
- Support `branch_protection_rule` webhook again.
- Update popular actions data set to the latest ([#64](https://github.com/rhysd/actionlint/issues/64), [#70](https://github.com/rhysd/actionlint/issues/70)).

[Changes][v1.6.7]


<a id="v1.6.6"></a>
# [v1.6.6](https://github.com/rhysd/actionlint/releases/tag/v1.6.6) - 2021-10-17

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
- `id-token` is added to permissions (thanks [@cmmarslender](https://github.com/cmmarslender), [#62](https://github.com/rhysd/actionlint/issues/62)).
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
- Popular actions data set was updated to the latest ([#61](https://github.com/rhysd/actionlint/issues/61)).
- Dependencies of playground were updated to the latest (including eslint v8).

[Changes][v1.6.6]


<a id="v1.6.5"></a>
# [v1.6.5](https://github.com/rhysd/actionlint/releases/tag/v1.6.5) - 2021-10-08

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
- Support `github.run_attempt` property in `${{ }}` expression ([#57](https://github.com/rhysd/actionlint/issues/57)).
- Add support for `windows-2022` runner which is now in [public beta](https://github.com/actions/virtual-environments/issues/3949).
- Remove support for `ubuntu-16.04` runner which was [removed from GitHub Actions at the end of September](https://github.com/actions/virtual-environments/issues/3287).
- Ignore [SC2154](https://github.com/koalaman/shellcheck/wiki/SC2154) shellcheck rule which can cause false positive ([#53](https://github.com/rhysd/actionlint/issues/53)).
- Fix error position was not correct when required keys are not existing in job configuration.
- Update popular actions data set. New major versions of github-script and lock-threads actions are supported ([#59](https://github.com/rhysd/actionlint/issues/59)).
- Fix document (thanks [@fornwall](https://github.com/fornwall) at [#52](https://github.com/rhysd/actionlint/issues/52), thanks [@equal-l2](https://github.com/equal-l2) at [#56](https://github.com/rhysd/actionlint/issues/56)).
  - Now actionlint is [an official package of Homebrew](https://formulae.brew.sh/formula/actionlint). Simply executing `brew install actionlint` can install actionlint.

[Changes][v1.6.5]


<a id="v1.6.4"></a>
# [v1.6.4](https://github.com/rhysd/actionlint/releases/tag/v1.6.4) - 2021-09-21

- Implement 'map' object types `{ string => T }`, where all properties of the object are typed as `T`. Since a key of object is always string, left hand side of `=>` is fixed to `string`. For example, `env` context only has string properties so it is typed as `{ string => string}`. Previously its properties were typed `any`.
  ```yaml
  # typed as string (previously any)
  env.FOO

  # typed as { id: string; network: string; ports: object; } (previously any)
  job.services.redis
  ```
- `github.event.discussion.title` and `github.event.discussion.body` are now checked as untrusted inputs.
- Update popular actions data set. ([#50](https://github.com/rhysd/actionlint/issues/50), [#51](https://github.com/rhysd/actionlint/issues/51))
- Update webhooks payload data set. `branch_protection_rule` hook was dropped from the list due to [github/docs@179a6d3](https://github.com/github/docs/commit/179a6d334e92b9ade8626ef42a546dae66b49951). ([#50](https://github.com/rhysd/actionlint/issues/50), [#51](https://github.com/rhysd/actionlint/issues/51))

[Changes][v1.6.4]


<a id="v1.6.3"></a>
# [v1.6.3](https://github.com/rhysd/actionlint/releases/tag/v1.6.3) - 2021-09-04

- Improve guessing a type of matrix value. When a matrix contains numbers and strings, previously the type fell back to `any`. Now it is deduced as string.
  ```yaml
  strategy:
    matrix:
      # matrix.node is now deduced as `string` instead of `any`
      node: [14, 'latest']
  ```
- Fix types of `||` and `&&` expressions. Previously they were typed as `bool` but it was not correct. Correct type is sum of types of both sides of the operator like TypeScript. For example, type of `'foo' || 'bar'` is a string, and `github.event && matrix` is an object.
- actionlint no longer reports an error when a local action does not exist in the repository. It is a popular pattern that a local action directory is cloned while a workflow running. ([#25](https://github.com/rhysd/actionlint/issues/25), [#40](https://github.com/rhysd/actionlint/issues/40))
- Disable [SC2050](https://github.com/koalaman/shellcheck/wiki/SC2050) shellcheck rule since it causes some false positive. ([#45](https://github.com/rhysd/actionlint/issues/45))
- Fix `-version` did not work when running actionlint via [the Docker image](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) ([#47](https://github.com/rhysd/actionlint/issues/47)).
- Fix pre-commit hook file name. (thanks [@xsc27](https://github.com/xsc27), [#38](https://github.com/rhysd/actionlint/issues/38))
- [New `branch_protection_rule` event](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#branch_protection_rule) is supported. ([#48](https://github.com/rhysd/actionlint/issues/48))
- Update popular actions data set. ([#41](https://github.com/rhysd/actionlint/issues/41), [#48](https://github.com/rhysd/actionlint/issues/48))
- Update Go library dependencies.
- Update playground dependencies.

[Changes][v1.6.3]


<a id="v1.6.2"></a>
# [v1.6.2](https://github.com/rhysd/actionlint/releases/tag/v1.6.2) - 2021-08-23

- actionlint now checks evaluated values at `${{ }}` are not an object nor an array since they are not useful. See [the check document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#check-type-check-expression) for more details.
```yaml
# ERROR: This will always be replaced with `echo 'Object'`
- run: echo '${{ runner }}'
# OK: Serialize an object into JSON to check the content
- run: echo '${{ toJSON(runner) }}'
```
- Add [pre-commit](https://pre-commit.com/) support. pre-commit is a framework for managing Git `pre-commit` hooks. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#pre-commit) for more details. (thanks [@xsc27](https://github.com/xsc27) for adding the integration at [#33](https://github.com/rhysd/actionlint/issues/33)) ([#23](https://github.com/rhysd/actionlint/issues/23))
- Add [an official Docker image](https://hub.docker.com/repository/docker/rhysd/actionlint). The Docker image contains shellcheck and pyflakes as dependencies. Now actionlint can be run with `docker run` command easily. See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#docker) for more details. (thanks [@xsc27](https://github.com/xsc27) for the help at [#34](https://github.com/rhysd/actionlint/issues/34))
```sh
docker run --rm -v $(pwd):/repo --workdir /repo rhysd/actionlint:latest -color
```
- Go 1.17 is now a default compiler to build actionlint. Built binaries are faster than before by 2~7% when the process is CPU-bound. Sizes of built binaries are about 2% smaller. Note that Go 1.16 continues to be supported.
- `windows/arm64` target is added to released binaries thanks to Go 1.17.
- Now any value can be converted into bool implicitly. Previously this was not permitted as actionlint provides stricter type check. However it is not useful that a condition like `if: github.event.foo` causes a type error.
- Fix a prefix operator cannot be applied repeatedly like `!!42`.
- Fix a potential crash when type checking on expanding an object with `${{ }}` like `matrix: ${{ fromJSON(env.FOO) }}`
- Update popular actions data set ([#36](https://github.com/rhysd/actionlint/issues/36))

[Changes][v1.6.2]


<a id="v1.6.1"></a>
# [v1.6.1](https://github.com/rhysd/actionlint/releases/tag/v1.6.1) - 2021-08-16

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
- Fix accessing `steps.*` objects in job's `environment:` configuration caused a type error ([#30](https://github.com/rhysd/actionlint/issues/30)).
- Fix checking that action's input names at `with:` were not in case insensitive ([#31](https://github.com/rhysd/actionlint/issues/31)).
- Ignore outputs of [getsentry/paths-filter](https://github.com/getsentry/paths-filter). It is a fork of [dorny/paths-filter](https://github.com/dorny/paths-filter). actionlint cannot check the outputs statically because it sets outputs dynamically.
- Add [Azure/functions-action](https://github.com/Azure/functions-action) to popular actions.
- Update popular actions data set ([#29](https://github.com/rhysd/actionlint/issues/29)).

[Changes][v1.6.1]


<a id="v1.6.0"></a>
# [v1.6.0](https://github.com/rhysd/actionlint/releases/tag/v1.6.0) - 2021-08-11

- Check potentially untrusted inputs to prevent [a script injection vulnerability](https://securitylab.github.com/research/github-actions-untrusted-input/) at `run:` and `script` input of [actions/github-script](https://github.com/actions/github-script). See [the rule document](https://github.com/rhysd/actionlint/blob/main/docs/checks.md#untrusted-inputs) for more explanations and workflow example. (thanks [@azu](https://github.com/azu) for the feature request at [#19](https://github.com/rhysd/actionlint/issues/19))

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

- Add `-format` option to `actionlint` command. It allows to flexibly format error messages as you like with [Go template syntax](https://pkg.go.dev/text/template). See [the usage document](https://github.com/rhysd/actionlint/blob/main/docs/usage.md#format) for more details. (thanks [@ybiquitous](https://github.com/ybiquitous) for the feature request at [#20](https://github.com/rhysd/actionlint/issues/20))

Simple example to output error messages as JSON:

```sh
actionlint -format '{{json .}}'
```

More compliated example to output error messages as markdown:

```sh
actionlint -format '{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\n\n{{$.Message}}\n\n```\n{{$.Snippet}}\n```\n\n{{end}}'
```

- Documents are reorganized. Long `README.md` is separated into several document files ([#28](https://github.com/rhysd/actionlint/issues/28))
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


<a id="v1.5.3"></a>
# [v1.5.3](https://github.com/rhysd/actionlint/releases/tag/v1.5.3) - 2021-08-04

- Now actionlint allows to use any operators outside `${{ }}` on `if:` condition like `if: github.repository_owner == 'rhysd'` ([#22](https://github.com/rhysd/actionlint/issues/22)). [The official document](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif) said that using any operator outside `${{ }}` was invalid even if it was on `if:` condition. However, [github/docs#8786](https://github.com/github/docs/pull/8786) clarified that the document was not correct.

[Changes][v1.5.3]


<a id="v1.5.2"></a>
# [v1.5.2](https://github.com/rhysd/actionlint/releases/tag/v1.5.2) - 2021-08-02

- Outputs of [dorny/paths-filter](https://github.com/dorny/paths-filter) are now not typed strictly because the action dynamically sets outputs which are not defined in its `action.yml`. actionlint cannot check such outputs statically ([#18](https://github.com/rhysd/actionlint/issues/18)).
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


<a id="v1.5.1"></a>
# [v1.5.1](https://github.com/rhysd/actionlint/releases/tag/v1.5.1) - 2021-07-29

- Improve checking the intervals of scheduled events ([#14](https://github.com/rhysd/actionlint/issues/14), [#15](https://github.com/rhysd/actionlint/issues/15)). Since GitHub Actions [limits the interval to once every 5 minutes](https://github.blog/changelog/2019-11-01-github-actions-scheduled-jobs-maximum-frequency-is-changing/), actionlint now reports an error when a workflow is configured to be run once per less than 5 minutes.
- Skip checking inputs of [octokit/request-action](https://github.com/octokit/request-action) since it allows to specify arbitrary inputs though they are not defined in its `action.yml` ([#16](https://github.com/rhysd/actionlint/issues/16)).
  - Outputs of the action are still be typed strictly. Only its inputs are not checked.
- The help text of `actionlint` is now hosted online: https://rhysd.github.io/actionlint/usage.html
- Add new fuzzing target for parsing glob patterns.

[Changes][v1.5.1]


<a id="v1.5.0"></a>
# [v1.5.0](https://github.com/rhysd/actionlint/releases/tag/v1.5.0) - 2021-07-26

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


<a id="v1.4.3"></a>
# [v1.4.3](https://github.com/rhysd/actionlint/releases/tag/v1.4.3) - 2021-07-21

- Support new Webhook events [`discussion` and `discussion_comment`](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#discussion) ([#8](https://github.com/rhysd/actionlint/issues/8)).
- Read file concurrently with limiting concurrency to number of CPUs. This improves performance when checking many files and disabling shellcheck/pyflakes integration.
- Support Linux based on musl libc by [the download script](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) ([#5](https://github.com/rhysd/actionlint/issues/5)).
- Reduce number of goroutines created while running shellcheck/pyflakes processes. This has small impact on memory usage when your workflows have many `run:` steps.
- Reduce built binary size by splitting an external library which is only used for debugging into a separate command line tool.
- Introduce several micro benchmark suites to track performance.
- Enable code scanning for Go/TypeScript/JavaScript sources in actionlint repository.

[Changes][v1.4.3]


<a id="v1.4.2"></a>
# [v1.4.2](https://github.com/rhysd/actionlint/releases/tag/v1.4.2) - 2021-07-16

- Fix executables in the current directory may be used unexpectedly to run `shellcheck` or `pyflakes` on Windows. This behavior could be security vulnerability since an attacker might put malicious executables in shared directories. actionlint searched an executable with [`exec.LookPath`](https://pkg.go.dev/os/exec#LookPath), but it searched the current directory on Windows as [golang/go#43724](https://github.com/golang/go/issues/43724) pointed. Now actionlint uses [`execabs.LookPath`](https://pkg.go.dev/golang.org/x/sys/execabs#LookPath) instead, which does not have the issue. (ref: [sharkdp/bat#1724](https://github.com/sharkdp/bat/pull/1724))
- Fix issue caused by running so many processes concurrently. Since checking workflows by actionlint is highly parallelized, checking many workflow files makes too many `shellcheck` processes and opens many files in parallel. This hit OS resources limitation (issue [#3](https://github.com/rhysd/actionlint/issues/3)). Now reading files is serialized and number of processes run concurrently is limited for fixing the issue. Note that checking workflows is still done in parallel so this fix does not affect actionlint's performance.
- Ensure cleanup processes even if actionlint stops due to some fatal issue while visiting a workflow tree.
- Improve fatal error message to know which workflow file caused the error.
- [Playground](https://rhysd.github.io/actionlint/) improvements
  - "Permalink" button was added to make permalink directly linked to the current workflow source code. The source code is embedded in hash of the URL.
  - "Check" button and URL input form was added to check workflow files on https://github.com or https://gist.github.com easily. Visit a workflow file on GitHub, copy the URL, paste it to the input form and click the button. It instantly fetches the workflow file content and checks it with actionlint.
  - `u=` URL parameter was added to specify GitHub or Gist URL like https://rhysd.github.io/actionlint/?u=https://github.com/rhysd/actionlint/blob/main/.github/workflows/ci.yaml

[Changes][v1.4.2]


<a id="v1.4.1"></a>
# [v1.4.1](https://github.com/rhysd/actionlint/releases/tag/v1.4.1) - 2021-07-12

- A pre-built executable for `darwin/arm64` (Apple M1) was added to CI ([#1](https://github.com/rhysd/actionlint/issues/1))
  - Managing `actionlint` command with Homebrew on M1 Mac is now available. See [the instruction](https://github.com/rhysd/actionlint#homebrew-on-macos) for more details
  - Since the author doesn't have M1 Mac and GitHub Actions does not support M1 Mac yet, the built binary is not tested
- Pre-built executables are now built with Go 1.16 compiler (previously it was 1.15)
- Fix error message is sometimes not in one line when the error message was caused by go-yaml/yaml parser
- Fix playground does not work on Safari browsers on both iOS and Mac since they don't support `WebAssembly.instantiateStreaming()` yet
- Make URLs in error messages clickable on playground
- Code base of playground was migrated from JavaScript to Typescript along with improving error handlings

[Changes][v1.4.1]


<a id="v1.4.0"></a>
# [v1.4.0](https://github.com/rhysd/actionlint/releases/tag/v1.4.0) - 2021-07-09

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


<a id="v1.3.2"></a>
# [v1.3.2](https://github.com/rhysd/actionlint/releases/tag/v1.3.2) - 2021-07-04

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


<a id="v1.3.1"></a>
# [v1.3.1](https://github.com/rhysd/actionlint/releases/tag/v1.3.1) - 2021-06-30

- Files are checked in parallel. This made actionlint around 1.3x faster with 3 workflow files in my environment
- Manual for `man` command was added. `actionlint.1` is included in released archives. If you installed actionlint via Homebrew, the manual is also installed automatically
- `-version` now reports how the binary was built (Go version, arch, os, ...)
- Added [`Command`](https://pkg.go.dev/github.com/rhysd/actionlint#Command) struct to manage entire command lifecycle
- Order of checked files is now stable. When all the workflows in the current repository are checked, the order is sorted by file names
- Added fuzz target for rule checkers

[Changes][v1.3.1]


<a id="v1.3.0"></a>
# [v1.3.0](https://github.com/rhysd/actionlint/releases/tag/v1.3.0) - 2021-06-26

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


<a id="v1.2.0"></a>
# [v1.2.0](https://github.com/rhysd/actionlint/releases/tag/v1.2.0) - 2021-06-25

- [pyflakes](https://github.com/PyCQA/pyflakes) integration was added. If `pyflakes` is installed on your system, actionlint checks Python scripts in `run:` (when `shell: python`) with it. See [the rule document](https://github.com/rhysd/actionlint#check-pyflakes-integ) for more details.
- Error handling while running rule checkers was improved. When some internal error occurs while applying rules, actionlint stops correctly due to the error. Previously, such errors were only shown in debug logs and actionlint continued checks.
- Fixed sanitizing `${{ }}` expressions in scripts before passing them to shellcheck or pyflakes. Previously expressions were not correctly sanitized when `}}` came before `${{`.

[Changes][v1.2.0]


<a id="v1.1.2"></a>
# [v1.1.2](https://github.com/rhysd/actionlint/releases/tag/v1.1.2) - 2021-06-21

- Run `shellcheck` command for scripts at `run:` in parallel. Since executing an external process is heavy and running shellcheck was bottleneck of actionlint, this brought better performance. In my environment, it was **more than 3x faster** than before.
- Sort errors by their positions in the source file.

[Changes][v1.1.2]


<a id="v1.1.1"></a>
# [v1.1.1](https://github.com/rhysd/actionlint/releases/tag/v1.1.1) - 2021-06-20

- [`download-actionlint.yaml`](https://github.com/rhysd/actionlint/blob/main/scripts/download-actionlint.bash) now sets `executable` output when it is run in GitHub Actions environment. Please see [instruction in 'Install' document](https://github.com/rhysd/actionlint#ci-services) for the usage.
- Redundant type `ArrayDerefType` was removed. Instead, [`Deref` field](https://pkg.go.dev/github.com/rhysd/actionlint#ArrayType) is now provided in `ArrayType`.
- Fix crash on broken YAML input.
- `actionlint -version` returns correct version string even if the `actionlint` command was installed via `go install`.

[Changes][v1.1.1]


<a id="v1.1.0"></a>
# [v1.1.0](https://github.com/rhysd/actionlint/releases/tag/v1.1.0) - 2021-06-19

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


<a id="v1.0.0"></a>
# [v1.0.0](https://github.com/rhysd/actionlint/releases/tag/v1.0.0) - 2021-06-16

First release :tada:

See documentation for more details:

- [Installation](https://github.com/rhysd/actionlint#install)
- [Usage](https://github.com/rhysd/actionlint#usage)
- [Checks done by actionlint](https://github.com/rhysd/actionlint#checks)

[Changes][v1.0.0]


[v1.7.8]: https://github.com/rhysd/actionlint/compare/v1.7.7...v1.7.8
[v1.7.7]: https://github.com/rhysd/actionlint/compare/v1.7.6...v1.7.7
[v1.7.6]: https://github.com/rhysd/actionlint/compare/v1.7.5...v1.7.6
[v1.7.5]: https://github.com/rhysd/actionlint/compare/v1.7.4...v1.7.5
[v1.7.4]: https://github.com/rhysd/actionlint/compare/v1.7.3...v1.7.4
[v1.7.3]: https://github.com/rhysd/actionlint/compare/v1.7.2...v1.7.3
[v1.7.2]: https://github.com/rhysd/actionlint/compare/v1.7.1...v1.7.2
[v1.7.1]: https://github.com/rhysd/actionlint/compare/v1.7.0...v1.7.1
[v1.7.0]: https://github.com/rhysd/actionlint/compare/v1.6.27...v1.7.0
[v1.6.27]: https://github.com/rhysd/actionlint/compare/v1.6.26...v1.6.27
[v1.6.26]: https://github.com/rhysd/actionlint/compare/v1.6.25...v1.6.26
[v1.6.25]: https://github.com/rhysd/actionlint/compare/v1.6.24...v1.6.25
[v1.6.24]: https://github.com/rhysd/actionlint/compare/v1.6.23...v1.6.24
[v1.6.23]: https://github.com/rhysd/actionlint/compare/v1.6.22...v1.6.23
[v1.6.22]: https://github.com/rhysd/actionlint/compare/v1.6.21...v1.6.22
[v1.6.21]: https://github.com/rhysd/actionlint/compare/v1.6.20...v1.6.21
[v1.6.20]: https://github.com/rhysd/actionlint/compare/v1.6.19...v1.6.20
[v1.6.19]: https://github.com/rhysd/actionlint/compare/v1.6.18...v1.6.19
[v1.6.18]: https://github.com/rhysd/actionlint/compare/v1.6.17...v1.6.18
[v1.6.17]: https://github.com/rhysd/actionlint/compare/v1.6.16...v1.6.17
[v1.6.16]: https://github.com/rhysd/actionlint/compare/v1.6.15...v1.6.16
[v1.6.15]: https://github.com/rhysd/actionlint/compare/v1.6.14...v1.6.15
[v1.6.14]: https://github.com/rhysd/actionlint/compare/v1.6.13...v1.6.14
[v1.6.13]: https://github.com/rhysd/actionlint/compare/v1.6.12...v1.6.13
[v1.6.12]: https://github.com/rhysd/actionlint/compare/v1.6.11...v1.6.12
[v1.6.11]: https://github.com/rhysd/actionlint/compare/v1.6.10...v1.6.11
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

<!-- Generated by https://github.com/rhysd/changelog-from-release v3.9.0 -->
