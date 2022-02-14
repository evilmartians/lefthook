build:
	go build -tags static,system_libgit2 -o lefthook cmd/lefthook/main.go

test:
	go test -count=1 -timeout=30s -race ./...

bench:
	go test -run=Bench -bench=. ./...

bin/golangci-lint:
	@test -x $$(go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.43.0

lint: bin/golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run
