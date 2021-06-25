# Development

`make` (3.81 or later) is useful to run each tasks and reduce redundant builds/tests.

## How to build

```sh
go build ./cmd/actionlint
./actionlint -h
```

or

```sh
make build
```

## How to run tests

```sh
go test ./...
```

or

```sh
make test
```

## How to run fuzzer

Fuzz tests use [go-fuzz](https://github.com/dvyukov/go-fuzz). Install `go-fuzz` and `go-fuzz-build` in your system.

Since there are multiple fuzzing targets, `-func` argument is necessary. Specify a target which you want to run.

```sh
# Create first corpus
go-fuzz-build ./fuzz

# Run fuzzer
go-fuzz -bin ./actionlint_fuzz-fuzz.zip -func FuzzParse
```

or

```sh
make fuzz FUZZ_FUNC=FuzzParse
```

## How to release

When releasing v1.2.3 as example:

1. Run `./scripts/make-release-tag.bash 1.2.3`. It modifies version string in `cmd/actionlint/main.go`, make a new tag for release, and pushes them to remote
2. Wait until [the CI release job](.github/workflows/release.yaml) completes successfully:
    - GoReleaser builds release binaries and make pre-release at GitHub and updates [Homebrew formula](./HomebrewFormula/actionlint.rb)
    - The CI job also updates version string in `./scripts/download-actionlint.bash`
3. Open the pre-release at [release page](https://github.com/rhysd/actionlint/releases) with browser
4. Write up release notes, uncheck pre-release checkbox and publish the new release
5. Run `git pull && changelog-from-release > CHANGELOG.md` locally to update [CHANGELOG.md](./CHANGELOG.md)
