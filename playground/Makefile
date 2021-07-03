serve: build
	python3 -m http.server 1234

build: install main.wasm

node_modules:
	npm install

lib: node_modules
	bash ./post-install.bash

install: node_modules lib

main.wasm: main.go
	GOOS=js GOARCH=wasm go build -o main.wasm

.PHONY: build install serve
