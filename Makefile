COMMIT_HASH = $(shell git rev-parse HEAD)

build:
	go build -ldflags "-s -w -X github.com/evilmartians/lefthook/internal/version.commit=$(COMMIT_HASH)" -o lefthook

test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

bin/golangci-lint:
	@test -x $$(go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1

lint: bin/golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run
