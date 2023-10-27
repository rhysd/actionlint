### Error at line 3, col 5 of `testdata/format/test.yaml`

unexpected key "branch" for "push" section. expected one of "branches", "branches-ignore", "paths", "paths-ignore", "tags", "tags-ignore", "types", "workflows"

```
    branch: main
    ^~~~~~~
```

### Error at line 6, col 14 of `testdata/format/test.yaml`

label "linux-latest" is unknown. available labels are "windows-latest", "windows-latest-8-cores", "windows-2022", "windows-2019", "windows-2016", "ubuntu-latest", "ubuntu-latest-4-cores", "ubuntu-latest-8-cores", "ubuntu-latest-16-cores", "ubuntu-22.04", "ubuntu-20.04", "macos-latest", "macos-latest-xl", "macos-latest-xlarge", "macos-latest-large", "macos-13-xl", "macos-13-xlarge", "macos-13-large", "macos-13", "macos-13.0", "macos-12-xl", "macos-12-xlarge", "macos-12-large", "macos-12", "macos-12.0", "macos-11", "macos-11.0", "macos-10.15", "self-hosted", "x64", "arm", "arm64", "linux", "macos", "windows". if it is a custom label for self-hosted runner, set list of labels in actionlint.yaml config file

```
    runs-on: linux-latest
             ^~~~~~~~~~~~
```

