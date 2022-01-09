generate-popular-actions
========================

This is a script for generating [`popular_actions.go`](../../popular_actions.go).

It does:

- Fetchs metadata of popular actions
  - from https://github.com
  - from JSONL file in local
- Generates the fetched data set of metadata
  - as Go source file
  - as JSONL file

## Usage

Generate Go source:

```sh
go run ./scripts/generate-popular-actions ./popular_actions.go
```

Detect new releases on GitHub:

```sh
go run ./scripts/generate-popular-actions -d
```

Please see output of `-help` flag for more details.
