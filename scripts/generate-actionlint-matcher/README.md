generate-actionlint-matcher
===========================

This script generates [`actionlint-matcher.json`](../../.github/actionlint-matcher.json).

## Usage

```sh
make .github/actionlint-matcher.json
```

or directly run the script

```sh
node ./scripts/generate-actionlint-matcher/main.js .github/actionlint-matcher.json
```

## Test

```sh
node ./scripts/generate-actionlint-matcher/test.js
```

The test uses test data at `./scripts/generate-actionlint-matcher/test/*.txt`. They should be updated when actionlint changes
the default error message format. To update them:

```sh
make ./scripts/generate-actionlint-matcher/test/escape.txt
make ./scripts/generate-actionlint-matcher/test/no_escape.txt
make ./scripts/generate-actionlint-matcher/test/want.json
```

or expand glob by your shell:

```sh
make ./scripts/generate-actionlint-matcher/test/*
```
