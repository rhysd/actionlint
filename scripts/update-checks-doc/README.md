update-checks-doc
=================

This is a script to maintain [the 'Checks' document](../../docs/checks.md).

This script does:

- update the outputs of the example inputs; the code blocks after `Output:` header
- update the links to the [playground](https://rhysd.github.io/actionlint/) for the example inputs

For making the implementation simple, this script does not support Windows.

## Prerequisites

- Go
- Linux or macOS
- `shellcheck` command
- `pyflakes` command

## Usage

```
go run ./scripts/update-checks-doc [-check] FILE
```

Update the document. This command directly modifies the file.

```sh
go run ./scripts/update-checks-doc ./docs/checks.md
```

Check the document is up-to-date.

```sh
go run ./scripts/update-checks-doc -check ./docs/checks.md
```

The check is run on the [CI workflow](../../.github/workflows/ci.yaml).

## Format

The format of the check document is:

````markdown
<a id="some-id"></a>
## This is title of the check

Example input:

```yaml
This section is referred to generate the output and the playground link
```

Output:

```
This section will be auto-generated
```

[Playground](URL_WILL_BE_AUTO_GENERATED)

Explanation for the check...
````

When you don't want to update the output by this script, put the comment as follows. This script
will ignore the code block. Instead you need to write the output in the code blcok manually.

```yaml
Output:
<!-- Skip update output -->
```

When you don't want to put a playground link for your example input, put the comment as follows
instead of the link. This script will not generate the hash for the playground link.

```yaml
<!-- Skip playground link -->
```
