COMMIT_HASH = $(shell git rev-parse HEAD)

build:
	go build -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

build-with-coverage:
	go build -cover -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

install: build
	cp lefthook $(GOPATH)/bin/

test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

test-integrity: install
	go test -cpu 24 -race -count=1 -timeout=30s -tags=integrity integrity_test.go

bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

bin/golangci-lint:
	@test -x $$(go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.1

lint: bin/golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run --fix

.ONESHELL:
version:
	@read -p "New version: " version
	sed -i "s/const version = .*/const version = \"$$version\"/" internal/version/version.go
	sed -i "s/VERSION := .*/VERSION := $$version/" packaging/Makefile
	make -C packaging clean set-version
	git add internal/version/version.go packaging/*
