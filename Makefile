SRCS := $(filter-out %_test.go, $(wildcard *.go cmd/actionlint/*.go)) go.mod go.sum .git-hooks/.timestamp
TESTS := $(filter %_test.go, $(wildcard *.go))
TOOL := $(filter %_test.go, $(wildcard scripts/*/*.go))
TESTDATA := $(wildcard \
		testdata/examples/* \
		testdata/err/* \
		testdata/ok/* \
		testdata/config/* \
		testdata/format/* \
		testdata/projects/* \
		testdata/reusable_workflow_metadata/* \
	)
GO_GEN_SRCS := scripts/generate-popular-actions/main.go \
				scripts/generate-popular-actions/popular_actions.json \
				scripts/generate-webhook-events/main.go \
				scripts/generate-availability/main.go

ifeq ($(OS),Windows_NT)
	SHELL := powershell.exe
	.SHELLFLAGS := -NoProfile -ExecutionPolicy Bypass -Command
	TARGET = actionlint.exe
	TOUCH = powershell -NoProfile -ExecutionPolicy Bypass scripts/touch.ps1
	# It's hard to prepare C toolchain for CGO on Windows
	RACE =
else
	TARGET = actionlint
	TOUCH = touch
	RACE = -race
endif


all: build test lint

.testtimestamp: $(TESTS) $(SRCS) $(TESTDATA) $(TOOL)
	go test $(RACE) ./...
	$(TOUCH) .testtimestamp

t test: .testtimestamp

coverage.out: $(TESTS) $(SRCS) $(TESTDATA) $(TOOL)
	go test $(RACE) -coverprofile coverage.out -covermode=atomic ./...
	$(TOUCH) .testtimestamp

coverage.html: coverage.out
	go tool cover -html=coverage.out -o coverage.html

cov: coverage.out coverage.html
	go tool cover -func=coverage.out

.linttimestamp: $(TESTS) $(SRCS) $(TOOL) docs/checks.md
	go vet ./...
	staticcheck ./...
	govulncheck ./...
ifneq ($(OS),Windows_NT)
	GOOS=js GOARCH=wasm staticcheck ./playground
	go run ./scripts/check-checks -quiet ./docs/checks.md
endif
	$(TOUCH) .linttimestamp

l lint: .linttimestamp

popular_actions.go all_webhooks.go availability.go: $(GO_GEN_SRCS)
ifdef SKIP_GO_GENERATE
	$(TOUCH) popular_actions.go all_webhooks.go availability.go
else
	go generate
endif

$(TARGET): $(SRCS)
ifeq ($(OS),Windows_NT)
	go build ./cmd/actionlint
else
	CGO_ENABLED=0 go build ./cmd/actionlint
endif

b build: $(TARGET)

actionlint_fuzz-fuzz.zip:
	go-fuzz-build ./fuzz

fuzz: actionlint_fuzz-fuzz.zip
	go-fuzz -bin ./actionlint_fuzz-fuzz.zip -func $(FUZZ_FUNC)

man/actionlint.1 man/actionlint.1.html: man/actionlint.1.ronn
	ronn man/actionlint.1.ronn

man: man/actionlint.1

bench:
	go test -bench Lint -benchmem

.github/actionlint-matcher.json: scripts/generate-actionlint-matcher/object.mjs
	node ./scripts/generate-actionlint-matcher/main.mjs .github/actionlint-matcher.json

scripts/generate-actionlint-matcher/test/escape.txt: $(TARGET)
	./actionlint -color ./testdata/err/one_error.yaml > ./scripts/generate-actionlint-matcher/test/escape.txt || true
scripts/generate-actionlint-matcher/test/no_escape.txt: $(TARGET)
	./actionlint -no-color ./testdata/err/one_error.yaml > ./scripts/generate-actionlint-matcher/test/no_escape.txt || true
scripts/generate-actionlint-matcher/test/want.json: $(TARGET)
	./actionlint -format '{{json .}}' ./testdata/err/one_error.yaml > scripts/generate-actionlint-matcher/test/want.json || true

CHANGELOG.md: .bumptimestamp
	changelog-from-release > CHANGELOG.md

c clean:
	rm -f ./$(TARGET) ./.testtimestamp ./.linttimestamp ./actionlint_fuzz-fuzz.zip ./man/actionlint.1 ./man/actionlint.1.html ./actionlint-workflow-ast
	rm -rf ./corpus ./crashers

.git-hooks/.timestamp: .git-hooks/pre-push
ifneq ($(OS),Windows_NT)
	[ -z "${CI}" ] && git config core.hooksPath .git-hooks || true
endif
	$(TOUCH) .git-hooks/.timestamp

.PHONY: all test clean build lint fuzz man bench cov b t c l
