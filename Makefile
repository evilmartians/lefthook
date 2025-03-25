COMMIT_HASH = $(shell git rev-parse HEAD)

build:
	go build -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

build-with-coverage:
	go build -cover -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

jsonschema:
	go generate internal/gen/jsonschema.go > schema.json

install: build
ifeq ($(shell go env GOOS),windows)
	copy lefthook $(shell go env GOPATH)\bin\lefthook.exe
else
	cp lefthook $$(go env GOPATH)/bin
endif

test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

test-integrity: install
	go test -cpu 24 -race -count=1 -timeout=30s -tags=integrity integrity_test.go

bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

bin/golangci-lint:
	@test -x $$(go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8

lint: bin/golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run --fix

.ONESHELL:
version:
	@read -p "New version: " version
	sed -i "s/const version = .*/const version = \"$$version\"/" internal/version/version.go
	sed -i "s/VERSION = .*/VERSION = \"$$version\"/" packaging/pack.rb
	sed -i "s/lefthook-plugin.git\", exact: \".*\"/lefthook-plugin.git\", exact: \"$$version\"/" docs/install.md
	sed -i "s/lefthook-plugin.git\", exact: \".*\"/lefthook-plugin.git\", exact: \"$$version\"/" docs/mdbook/installation/swift.md
	ruby packaging/pack.rb clean set_version
	git add internal/version/version.go packaging/* docs/
