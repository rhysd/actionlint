#!/bin/bash

set -o pipefail
set -e

if [ ! -d .git ]; then
    echo 'This script must be run from root of repository' >&2
    exit 1
fi

set -x

script="$(pwd)/scripts/download-actionlint.bash"
temp_dir="$(mktemp -d)"
trap 'popd && rm -rf $temp_dir' EXIT
pushd "$temp_dir"

# Normal cases
set -e

# No arguments
out="$(bash "$script")"
if [ -n "$GITHUB_ACTION" ]; then
    if [[ "$out" != *"::set-output name=executable::"* ]]; then
        echo "'executable' step output is not set: '${out}'" >&2
    fi
fi
out="$(./actionlint -version)"
if [[ "$out" != *'installed by downloading from release page'* ]]; then
    echo "Output from ./actionlint -version is unexpected: '${out}'" >&2
    exit 1
fi
rm -f ./actionlint

# Specify only version
bash "$script" '1.6.12'
out="$(./actionlint -version | head -n 1)"
if [[ "$out" != '1.6.12' ]]; then
    echo "Unexpected version: '${out}'" 1>&2
    exit 1
fi
rm -f ./actionlint

# Specify only a download directory
mkdir ./test1
bash "$script" ./test1
out="$(./test1/actionlint -version)"
if [[ "$out" != *'installed by downloading from release page'* ]]; then
    echo "Output from ./actionlint -version is unexpected: '${out}'" >&2
    exit 1
fi
rm -rf ./test1

# Specify both version and a download directory
mkdir ./test2
bash "$script" ./test2 '1.6.12'
out="$(./test2/actionlint -version | head -n 1)"
if [[ "$out" != '1.6.12' ]]; then
    echo "Unexpected version: '${out}'" 1>&2
    exit 1
fi
rm -rf ./test2

# Error cases
set +e

if bash "$script" 'v1.6.12'; then
    echo "Argument 'v1.6.12' at the first argument did not cause any error" >&2
fi
if bash "$script" . 'v1.6.12'; then
    echo "Argument 'v1.6.12' at the second argument did not cause any error" >&2
fi
if bash "$script" './this/dir/does/not/exist'; then
    echo "Argument './this/dir/does/not/exist' at the first argument did not cause any error" >&2
fi

echo 'SUCCESS'
