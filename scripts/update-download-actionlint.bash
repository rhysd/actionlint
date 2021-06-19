#!/bin/bash

set -e -o pipefail

# if [[ "$OSTYPE" != linux-gnu* ]]; then
#     echo "This script must be run on Linux: ${OSTYPE}" >&2
#     exit 1
# fi

if [ ! -d .git ]; then
    echo 'Run this script at root of repository' >&2
    exit 1
fi

# if [[ ! "$1" =~ ^[1-9]+\.[0-9]\.[0-9]$ ]]; then
if [[ "$1" != test ]]; then
    echo "First argument must be in format of 'x.y.z': \$1='$1'" >&2
    exit 1
fi

ver="$1"
echo "Update to version $ver"

sed -i -E "s/const version = \"[^\"]+\"/const version = \"$ver\"/" ./cmd/actionlint/main.go
echo 'Updated ./cmd/actionlint/main.go'

sed -i -E "s/version=\"[^\"]+\"/version=\"$ver\"/" ./scripts/download-actionlint.bash
echo 'Updated ./scripts/download-actionlint.bash'

set -x
git diff
git add ./cmd/actionlint/main.go ./scripts/download-actionlint.bash
git -c user.email='github@users.noreply.github.com' -c user.name='github-actions' commit -m "update versions in sources to $ver by update-download-actionlint.bash"
git log -n 1
git push origin HEAD:refs/remotes/origin/update-ver-script
