This directory contains project directories tested by `TestLinterLintProject` in `linter_test.go`.

Each directory represents one project. And `<project>.out` file describes all errors when linting the project (one error per
line). Error messages in the file must be sorted by file paths and error positions. An empty file means no error.

Each project must contain `workflows` directory and it must contain at least one workflow file.

Working directory is set to `projects/<project>` when running the tests. File paths in `.out` files should be relative to the
project directory.

```
├── some_project
│   ├── action
│   │   └── action1.yaml
│   └── workflows
│       ├── workflow1.yaml
│       └── workflow2.yaml
└── some_project.out
```
