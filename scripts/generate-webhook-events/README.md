generate-webhook-events
=======================

This is a script for generating [`all_webhooks.go`](../../all_webhooks.go).

It does:

1. Fetch [the GitHub Docs HTML page](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)
2. Parse the HTML and find webhook names and their activity types from tables
3. Generate mappings from webhook names to their activity types as a Go map variable

## Usage

```
generate-webhook-events [[srcfile] dstfile]
```

Generate `all_webhooks.go` file:

```sh
go run ./scripts/generate-webhook-events ./all_webhooks.go
```

When the HTML file is in local:

```sh
go run ./scripts/generate-webhook-events ./input.html ./all_webhooks.go
```

For debugging, specifying `-` to `dstfile` outputs the generated source to stdout:

```sh
go run ./scripts/generate-webhook-events ./input.html -
```

## Notes

- `Not applicable` activity types are generated as an empty slice: `[]string{}`
- `Custom` activity types are generated as `nil`
- The output is sorted by webhook name so the generated file is stable
