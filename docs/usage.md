Usage
=====

This document describes how to use [actionlint](..).

## `actionlint` command

With no argument, actionlint finds all workflow files in the current repository and checks them.

```sh
actionlint
```

When paths to YAML workflow files are given as arguments, actionlint checks them.

```sh
actionlint path/to/workflow1.yaml path/to/workflow2.yaml
```

For all flags and options, see an output of `actionlint -h` or [an online command manual][cmd-manual].

Note that actionlint focuses on catching mistakes in workflow files. If you want some general code style checks, please consider
using a general YAML checker like [yamllint][].

### Ignore some errors

TODO

### Exit status

`actionlint` command exits with one of the following exit statuses.

| Status | Description                                             |
|--------|---------------------------------------------------------|
| `0`    | The command ran successfully and no problem was found   |
| `1`    | The command ran successfully and some problem was found |
| `2`    | The command failed due to invalid command line option   |
| `3`    | The command failed due to some fatal error              |

## Online playground

Thanks to WebAssembly, actionlint playground is available on your browser. It never sends any data to outside of your browser.

https://rhysd.github.io/actionlint/

Paste your workflow content to the code editor at left pane. It automatically shows the results at right pane. When editing
the workflow content in the code editor, the results will be updated on the fly. Clicking an error message in the results
table moves a cursor to position of the error in the code editor.

## Tools integration

These tools have integration with actionlint:

- **[reviewdog/action-actionlint][reviewdog-actionlint]**: [reviewdog][] is an automated review tool for various code hosting
  services. It officially supports actionlint. You can check errors from actionlint easily in inline comments on code review.

[yamllint]: https://github.com/adrienverge/yamllint
[reviewdog-actionlint]: https://github.com/reviewdog/action-actionlint
[reviewdog]: https://github.com/reviewdog/reviewdog
[cmd-manual]: https://rhysd.github.io/actionlint/usage.html
