SRCS := $(filter-out %_test.go, $(wildcard *.go cmd/actionlint/*.go))
TESTS := $(filter %_test.go, $(wildcard *.go))
TOOL_SRCS := $(wildcard scripts/*/*.go)
TESTDATA := $(wildcard testdata/examples/*.yaml testdata/examples/*.out)
TOOL := $(wildcard scripts/actionlint-workflow-ast/*.go)
GOTEST := $(shell command -v gotest 2>/dev/null)

all: clean build test

.testtimestamp: $(TESTS) $(SRCS) $(TESTDATA) $(TOOL)
ifdef GOTEST
	gotest ./ ./scripts/... # https://github.com/rhysd/gotest
else
	go test -v ./ ./scripts/...
endif
	touch .testtimestamp

test: .testtimestamp

.staticchecktimestamp: $(TESTS) $(SRCS) $(TOOL_SRCS)
	staticcheck ./ ./cmd/... ./scripts/...
	GOOS=js GOARCH=wasm staticcheck ./playground
	touch .staticchecktimestamp

lint: .staticchecktimestamp

popular_actions.go: scripts/generate-popular-actions/main.go
	go generate

actionlint: $(SRCS) popular_actions.go
	CGO_ENABLED=0 go build ./cmd/actionlint

build: actionlint

actionlint_fuzz-fuzz.zip:
	go-fuzz-build ./fuzz

fuzz: actionlint_fuzz-fuzz.zip
	go-fuzz -bin ./actionlint_fuzz-fuzz.zip -func $(FUZZ_FUNC)

man/actionlint.1: man/actionlint.1.ronn
	ronn man/actionlint.1.ronn

man: man/actionlint.1

bench:
	go test -bench Lint -benchmem

actionlint-workflow-ast: ./scripts/actionlint-workflow-ast/main.go
	go build ./scripts/actionlint-workflow-ast/

clean:
	rm -f ./actionlint ./.testtimestamp ./.staticchecktimestamp ./actionlint_fuzz-fuzz.zip ./man/actionlint.1 ./man/actionlint.1.html ./actionlint-workflow-ast
	rm -rf ./corpus ./crashers

b: build
t: test
c: clean
l: lint

.PHONY: all test clean build lint fuzz man bench b t c l
