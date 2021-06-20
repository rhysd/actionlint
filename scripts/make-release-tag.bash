#!/bin/bash

set -e -o pipefail

case "$OSTYPE" in
    linux-gnu*) sed='sed' ;;
    darwin*) sed='gsed' ;;
    *)
        echo "OS '${OSTYPE}' is not supported. Run this script on Linux or macOS" >&2
        exit 1
    ;;
esac

if [ ! -d .git ]; then
    echo 'Run this script at root of repository' >&2
    exit 1
fi

if [[ ! "$1" =~ ^[0-9]+\.[0-9]\.[0-9]$ ]]; then
    echo "First argument '$1' does not match to regular expression" '^[0-9]+\.[0-9]\.[0-9]$' >&2
    echo 'Usage: ./scripts/make-release-tag.bash x.y.z' >&2
    exit 1
fi

ver="$1"
echo "Update to version $ver"

set -x
"${sed}" -i -E "s/const version = \"[^\"]+\"/const version = \"$ver\"/" ./cmd/actionlint/main.go
git diff
git add ./cmd/actionlint/main.go
git commit -m "update version to $ver by make-release-tag.bash"
git push
git tag "v${ver}"
git push origin "v${ver}"

echo "Done successfully"
