PYTHON3 := $(shell command -v python3 2>/dev/null)

serve: build
ifdef PYTHON3
	python3 -m http.server 1234
else
	python -m http.server 1234
endif

build: install main.wasm

node_modules:
	npm install

lib: node_modules
	bash ./post-install.bash

install: node_modules lib

main.wasm: main.go
	GOOS=js GOARCH=wasm go build -o main.wasm

clean:
	rm -f ./main.wasm
	rm -rf ./lib

.PHONY: build install serve clean
