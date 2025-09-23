generate-webhook-events
=======================

This is a script for generating [`all_webhooks.go`](../../all_webhooks.go).

It does:

1. Fetch [the official markdown document](https://raw.githubusercontent.com/github/docs/refs/heads/main/content/actions/reference/workflows-and-actions/events-that-trigger-workflows.md)
2. Parse the markdown file and find Webhook names and their types from tables
3. Generate mappings from Webhook names to their types as Go map variable

## Usage

```
generate-webhook-events [[srcfile] dstfile]
```

Generate `all_webhooks.go` file:

```sh
go run ./scripts/generate-webhook-events ./all_webhooks.go
```

When the markdown file is in local:

```sh
go run ./scripts/generate-webhook-events ./events-that-trigger-workflows.md ./all_webhooks.go
```

For debugging, specifying `-` to `dstfile` outputs the generated source to stdout:

```sh
go run ./scripts/generate-webhook-events -
```

