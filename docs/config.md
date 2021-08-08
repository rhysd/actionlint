Configuration
=============

This document describes how to configure [actionlint](..) behavior.

Note that configuration file is optional. The author tries to keep configuration file as minimal as possible not to
bother users to configure behavior of actionlint. Running actionlint without configuration file would work fine in most
cases.

## Configuration file

Configuration file `actionlint.yaml` or `actionlint.yml` can be put in `.github` directory.

You don't need to write the first configuration file by your hand. `actionlint` command can generate a default configuration
with `-init-config` flag.

```sh
actionlint -init-config
vim .github/actionlint.yaml
```

Currently only one item can be configured.

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


