generate-availability
=====================

This is a script for generating [`availability.go`](../../availability.go).

It does:

1. Fetch [the official contexts document](https://raw.githubusercontent.com/github/docs/main/content/actions/writing-workflows/choosing-what-your-workflow-does/accessing-contextual-information-about-workflow-runs.md)
2. Parse the markdown file and find "Context availability" table
3. Extract contexts and special functions from the table
4. Generate Go function and variable to map from workflow keys to available contexts and special functions

## Background

GitHub Actions limits contexts and functions in certain places. For example:

- limited workflow keys can access `secrets` context
- `jobs.<job_id>.if` and `jobs.<job_id>.steps.if` can use `always()` function.

To check these limitations by actionlint, we maintain a table to map workflow keys to available contexts and special functions.

## Usage

```
generate-availability [[srcfile] dstfile]
```

For generating the source at root directory of this repository:

```sh
go run ./scripts/generate-availability ./availability.go
```

Read local file instead of fetching it from remote:

```sh
go run ./scripts/generate-availability /path/to/contexts.md ./availability.go
```

For debugging, specifying `-` to `dstfile` outputs the generated source to stdout:

```sh
go run ./scripts/generate-availability -
```
