SRCS := $(filter-out %_test.go, $(wildcard *.go cmd/actionlint/*.go))
TESTS := $(filter %_test.go, $(wildcard *.go))
GOTEST := $(shell command -v gotest 2>/dev/null)

all: clean build test

.testtimestamp: $(TESTS) $(SRCS)
ifdef GOTEST
	gotest  # https://github.com/rhysd/gotest
else
	go test -v
endif
	touch .testtimestamp

test: .testtimestamp

actionlint: $(SRCS)
	go build ./cmd/actionlint

build: actionlint

clean:
	rm -f ./actionlint ./.testtimestamp

b: build
t: test

.PHONY: all test clean build b t
