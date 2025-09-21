#!/bin/bash

set -e -o pipefail

rm -rf ./lib
mkdir -p ./lib/css/fonts
mkdir -p ./lib/js

cp node_modules/codemirror/lib/codemirror.css ./lib/css/
cp node_modules/codemirror/theme/material-darker.css ./lib/css/
cp node_modules/bulma/css/bulma.min.css ./lib/css/
cp node_modules/devicon/devicon.min.css ./lib/css/

cp node_modules/devicon/fonts/* ./lib/css/fonts/

cp node_modules/codemirror/lib/codemirror.js ./lib/js/
cp node_modules/codemirror/addon/selection/active-line.js ./lib/js/
cp node_modules/codemirror/mode/yaml/yaml.js ./lib/js/
cp node_modules/ismobilejs/dist/isMobile.min.js ./lib/js/
cp node_modules/pako/dist/pako.min.js ./lib/js/

cat "$(go env GOROOT)/lib/wasm/wasm_exec.js" >> ./lib/js/wasm_exec.js
