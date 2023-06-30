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
