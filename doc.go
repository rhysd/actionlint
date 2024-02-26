/*
Package actionlint is the implementation of actionlint linter. It's a static checker for GitHub
Actions workflow files.

https://github.com/rhysd/actionlint

actionlint is a command line tool but it also provides Go API for Go programs. It includes a
workflow file parser built on top of go-yaml/yaml, lexer/parser/checker for expressions embedded by
${{ }} placeholder, popular actions data, available contexts information, etc.

To run the linter, Linter is the struct which manages the entire linter lifecycle. Please see the
first example.

actionlint also provides the flexibility to add your own rules by implementing Rule interface.
Please read the YourOwnRule example.

# Library versioning

The version is for the command line tool. So it does not represent the version of the library. It
means that the library does not follow sematinc versioning and any patch version bump may introduce
some breaking changes.

# Go version compatibility

Minimum supported Go version is written in go.mod file in this library. That said, older Go versions
are actually not tested on CI. Last two major Go versions are recommended because they're tested on
CI. For example, when the latest Go version is v1.22, v1.21 and v1.22 are nice to use.

# Other documentations

All documentations for actionlint can be found in the following page.

https://github.com/rhysd/actionlint/tree/main/docs
*/
package actionlint
