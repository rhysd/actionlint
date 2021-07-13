serve: build
	npm run serve

build: install main.wasm

node_modules:
	npm install

index.js: index.ts lib.d.ts tsconfig.json
	npm run build

lib: node_modules index.js
	bash ./post-install.bash

install: node_modules lib

main.wasm: main.go
	GOOS=js GOARCH=wasm go build -o main.wasm

clean:
	rm -f ./main.wasm ./index.js ./index.js.map
	rm -rf ./lib

.PHONY: build install serve clean
