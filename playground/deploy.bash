#!/bin/bash

set -e -o pipefail

if [ ! -d .git ]; then
    echo 'This script must be run from root of repository: bash ./playground/deploy.bash' 1>&2
    exit 1
fi

sha="$(git rev-parse HEAD)"
echo "Deploying playground from ${sha}"

echo 'Ensuring to install dependencies and building wasm'
(cd ./playground && make build)

echo 'Creating ./playground-dist'
rm -rf ./playground-dist
mkdir ./playground-dist

files=(
    index.html
    index.js
    index.js.map
    index.ts
    lib
    main.wasm
    style.css
)

echo 'Copying built assets from ./playground to ./playground-dist'
for f in "${files[@]}"; do
    cp -R "./playground/${f}" "./playground-dist/${f}"
done

echo 'Applying wasm-opt to ./playground-dist/main.wasm'
wasm-opt -O -o ./playground-dist/opt.wasm ./playground-dist/main.wasm
mv ./playground-dist/opt.wasm ./playground-dist/main.wasm

echo 'Switching to gh-pages branch'
git checkout gh-pages

echo 'Removing previous assets in branch'
for f in "${files[@]}"; do
    # This command fails when $f is new file
    git rm -rf "./${f}" || true
done

echo 'Adding new assets to branch'
for f in "${files[@]}"; do
    mv "./playground-dist/${f}" .
    git add "./${f}"
done

echo 'Making commit for new deploy'
git commit -m "deploy from ${sha}"

rm -r ./playground-dist
echo "Successfully prepared deployment. Run 'python3 -m http.server 1234' and do the final check before deployment. If it looks good, deploy it by 'git push'"
