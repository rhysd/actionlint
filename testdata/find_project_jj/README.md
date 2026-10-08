This directory is used for testing that actionlint can detect a repository root of a [Jujutsu](https://github.com/jj-vcs/jj)
workspace, which has `.jj` directory but no `.git`.

- `project_test.go`

`.jj` directory is dynamically created when the test case is run.
