#!/bin/bash

set -e -o pipefail

if [ ! -d .git ]; then
    echo 'This script must be run from root of repository: bash ./playground/deploy.bash' 1>&2
    exit 1
fi

echo "Ensuring gh-pages branch is up-to-date"
git fetch -u origin gh-pages:gh-pages

sha="$(git rev-parse HEAD)"
echo "Deploying playground from ${sha}"

echo 'Ensuring to install dependencies and building wasm'
(cd ./playground && make clean && make build && make test)

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

if [[ "$SKIP_BUILD_WASM" != "" ]]; then
    # Remove main.wasm from $files
    for i in "${!files[@]}"; do
        if [[ "${files[i]}" == "main.wasm" ]]; then
            unset 'files[i]'
            break
        fi
    done
fi

echo "Copying built assets from ./playground to ./playground-dist: " "${files[@]}"
for f in "${files[@]}"; do
    cp -R "./playground/${f}" "./playground-dist/${f}"
done

if [[ "$SKIP_BUILD_WASM" == "" ]]; then
    echo 'Applying wasm-opt to ./playground-dist/main.wasm'
    wasm-opt -O -o ./playground-dist/opt.wasm ./playground-dist/main.wasm --enable-bulk-memory
    mv ./playground-dist/opt.wasm ./playground-dist/main.wasm
else
    echo 'Skipped applying wasm-opt because SKIP_BUILD_WASM environment variable is set'
fi

echo 'Generating and copying manual'
make ./man/actionlint.1.html
cp ./man/actionlint.1.html ./playground-dist/man.html

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

echo 'Adding manual'
cp ./playground-dist/man.html ./man.html
git add ./man.html

echo 'Making commit for new deploy'
git commit -m "deploy from ${sha}"

rm -r ./playground-dist

echo "Successfully prepared deployment. Visit http://localhost:1234 and do the final check before deployment. If it looks good, stop the server with Ctrl+C and deploy it by 'git push'"
(trap '' INT; ./playground/node_modules/.bin/http-server . -p 1234 || true)
