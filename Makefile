SRCS := $(filter-out %_test.go, $(wildcard *.go cmd/actionlint/*.go))
TESTS := $(filter %_test.go, $(wildcard *.go))
TESTDATA := $(wildcard testdata/examples/*.yaml testdata/examples/*.out)
GOTEST := $(shell command -v gotest 2>/dev/null)

all: clean build test

.testtimestamp: $(TESTS) $(SRCS) $(TESTDATA)
ifdef GOTEST
	gotest  # https://github.com/rhysd/gotest
else
	go test -v
endif
	touch .testtimestamp

test: .testtimestamp

actionlint: $(SRCS)
	CGO_ENABLED=0 go build ./cmd/actionlint

build: actionlint

actionlint_fuzz-fuzz.zip:
	go-fuzz-build ./fuzz

fuzz: actionlint_fuzz-fuzz.zip
	go-fuzz -bin ./actionlint_fuzz-fuzz.zip -func $(FUZZ_FUNC)

man/actionlint.1: man/actionlint.1.ronn
	ronn man/actionlint.1.ronn

man: man/actionlint.1

clean:
	rm -f ./actionlint ./.testtimestamp ./actionlint_fuzz-fuzz.zip ./man/actionlint.1 ./man/actionlint.1.html
	rm -rf ./corpus ./crashers

b: build
t: test
c: clean

.PHONY: all test clean build fuzz man b t c
