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

1. Ensure all changes were already pushed to remote by checking `git push origin master` outputs `Everything up-to-date`
2. `git tag v1.2.3 && git push origin v1.2.3`
3. Wait until [the CI release job](.github/workflows/release.yaml) completes successfully:
    - GoReleaser builds release binaries and make pre-release at GitHub and updates [Homebrew formula](./HomebrewFormula/actionlint.rb)
    - The CI job also updates version string in `./scripts/download-actionlint.bash`
4. Open the pre-release at [release page](https://github.com/rhysd/actionlint/releases) with browser
5. Write up release notes, uncheck pre-release checkbox and publish the new release
6. Run `git pull && changelog-from-release > CHANGELOG.md` locally to update [CHANGELOG.md](./CHANGELOG.md)

## How to generate manual

`actionlint.1` manual is generated from [`actionlint.1.ronn`](./man/actionlint.1.ronn) by [ronn](https://github.com/rtomayko/ronn).

```sh
ronn ./man/actionlint.1.ronn
```

or

```sh
make ./man/actionlint.1
```
