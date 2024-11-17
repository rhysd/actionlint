#!/bin/bash

set -e -o pipefail

if [ ! -d .git ]; then
    echo 'This script must be run from repository root' 1>&2
    echo 'Usage: bash ./scripts/bump-version.bash 1.2.3' 1>&2
    exit 1
fi

if ! git diff --quiet; then
    echo 'Working tree is dirty! Ensure all changes are committed and working tree is clean' >&2
    exit 1
fi

if ! git diff --cached --quiet; then
    echo 'Git index is dirty! Ensure all changes are committed and Git index is clean' >&2
    exit 1
fi

if [[ "$(git rev-parse --abbrev-ref HEAD)" != main ]]; then
    echo "This script must be run on 'main' branch" >&2
    exit 1
fi

version="$1"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "The first argument must match to '^\d+\.\d+\.\d+$': ${version}" 1>&2
    echo 'Usage: bash ./scripts/bump-version.bash 1.2.3' 1>&2
    exit 1
fi

function sed_() {
    case "$OSTYPE" in
        darwin*)
            /usr/bin/sed -i '' -E "$@"
            ;;
        *)
            sed -i -E "$@"
            ;;
    esac
}

pre_commit_hook='./.pre-commit-hooks.yaml'
usage_doc='./docs/usage.md'
tag="v${version}"
job_url='https://github.com/rhysd/actionlint/actions/workflows/release.yaml'
playground_html='./playground/index.html'
readme_doc='./README.md'
man_ronn='./man/actionlint.1.ronn'

echo "Bumping up version to ${version} (tag: ${tag})"

# Update container image tag in pre-commit hook (See #116 for more details)
echo "Updating $pre_commit_hook"
sed_ "s/entry: docker\\.io\\/rhysd\\/actionlint:.*/entry: docker.io\\/rhysd\\/actionlint:${version}/" "$pre_commit_hook"

echo "Updating $usage_doc"
sed_ "\
    s/    rev: v[0-9]+\.[0-9]+\.[0-9]+/    rev: v${version}/; \
    s/ actionlint@[0-9]+\.[0-9]+\.[0-9]+/ actionlint@${version}/g; \
    s/\`actionlint:[0-9]+\.[0-9]+\.[0-9]+\`/\`actionlint:${version}\`/g; \
    " "$usage_doc"

echo "Updating $playground_html"
sed_ "s/id=\"version\">v[0-9]+\.[0-9]+\.[0-9]+/id=\"version\">v${version}/" "$playground_html"

for f in "$readme_doc" "$man_ronn" "$playground_html"; do
    echo "Updating document links in $f"
    sed_ "s/\/rhysd\/actionlint\/blob\/v[0-9]+\.[0-9]+\.[0-9]+\/docs\//\/rhysd\/actionlint\/blob\/v${version}\/docs\//g" "$f"
done

echo 'Creating a version bump commit and a version tag'
git add "$pre_commit_hook" "$usage_doc" "$playground_html" "$readme_doc" "$man_ronn"
git commit -m "bump up version to ${tag}"
git tag "$tag"

# This is necessary since docker/build-push-action assumes the tagged commit was also pushed to main branch
echo "Pushing bump commit to main"
git push origin main

echo "Pushing the version tag '${tag}'"
git push origin "$tag"

echo "Open release job page to check release progress ${job_url}"
if [[ "$OSTYPE" == darwin* ]]; then
    open "$job_url"
fi

echo "Update version bump timestamp"
touch .bumptimestamp

echo 'Done.'
