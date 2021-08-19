LIBSRCS := $(filter-out ../%_test.go, $(wildcard ../*.go)) ../go.mod ../go.sum

serve: build
	npm run serve

build: node_modules lib main.wasm index.js test.js

node_modules:
	npm install

index.js test.js: index.ts test.ts lib.d.ts tsconfig.json
	npm run build

lib: node_modules index.js
	bash ./post-install.bash

install: node_modules lib

main.wasm: main.go $(LIBSRCS)
	GOOS=js GOARCH=wasm go build -o main.wasm

.testtimestamp: node_modules lib test.js main.wasm
	npm test
	touch .testtimestamp

test: .testtimestamp

clean:
	rm -f ./main.wasm ./index.js ./index.js.map .testtimestamp
	rm -rf ./lib

.PHONY: build install serve test clean
