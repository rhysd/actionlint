This directory is used for testing that actionlint can detect a repository root.

- `TestLintFindProjectFromPath` in `linter_test.go`
- `project_test.go`

`.git` directory is dynamically created when the test case is run because Git doesn't allow committing `.git` directory.
