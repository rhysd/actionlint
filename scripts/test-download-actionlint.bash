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

bash "$script" '1.6.12'
out="$(./actionlint -version | head -n 1)"
if [[ "$out" != '1.6.12' ]]; then
    echo "Unexpected version: '${out}'" 1>&2
    exit 1
fi
rm -f ./actionlint

# Error cases
set +e
if bash "$script" 'v1.6.12'; then
    echo "Argument 'v1.6.12' did not cause any error" >&2
fi
