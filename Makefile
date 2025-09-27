COMMIT_HASH = $(shell git rev-parse HEAD)

.PHONY: build
build:
	go build -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

.PHONY: build-with-coverage
build-with-coverage:
	go build -cover -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

.PHONY: jsonschema
jsonschema:
	go generate gen/jsonschema.go > schema.json

install: build
ifeq ($(shell go env GOOS),windows)
	copy lefthook $(shell go env GOPATH)\bin\lefthook.exe
else
	cp lefthook $$(go env GOPATH)/bin
endif

.PHONY: test
test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

.PHONY: test-integration
test-integration: install
	go test -cpu 24 -race -count=1 -timeout=30s -tags=integration integration_test.go

.PHONY: bench
bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint run --fix

bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin/ v2.5.0

.ONESHELL:
version:
	@read -p "New version: " version
	sed -i "s/const version = .*/const version = \"$$version\"/" internal/version/version.go
	sed -i "s/VERSION = .*/VERSION = \"$$version\"/" packaging/pack.rb
	sed -i "s/lefthook-plugin.git\", exact: \".*\"/lefthook-plugin.git\", exact: \"$$version\"/" docs/mdbook/installation/swift.md
	sed -i "s/go install github.com\/evilmartians\/lefthook.*/go install github.com\/evilmartians\/lefthook@v$$version/" docs/mdbook/installation/go.md
	sed -i "s/go install github.com\/evilmartians\/lefthook.*/go install github.com\/evilmartians\/lefthook@v$$version/" README.md
	ruby packaging/pack.rb clean set_version
	git add internal/version/version.go packaging/* docs/ README.md
