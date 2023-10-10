Files in this directory are used for testing `-format` option of actionlint. Test cases are in `linter_test.go`.

How to generate `test.sarif`:

```sh
cd /path/to/actionlint
make build

# Generate output formatting with jq
./actionlint -pyflakes= -shellcheck= -format "$(cat testdata/format/sarif_template.txt)" testdata/format/test.yaml | jq . > test.sarif

# Remove version because actionlint.version is empty string while running unit tests
sed -i 's/(devel)//' test.sarif

mv test.sarif testdata/format/
```

How to generate other files:

```sh
./actionlint -pyflakes= -shellcheck= -format '{{json .}}' testdata/format/test.yaml > testdata/format/test.json
./actionlint -pyflakes= -shellcheck= -format '{{range $err := .}}{{json $err}}{{end}}' testdata/format/test.yaml > testdata/format/test.jsonl
./actionlint -pyflakes= -shellcheck= -format '{{range $ := .}}### Error at line {{$.Line}}, col {{$.Column}} of `{{$.Filepath}}`\n\n{{$.Message}}\n\n```\n{{$.Snippet}}\n```\n\n{{end}}' testdata/format/test.yaml > testdata/format/test.md
```
