build:
	go build -o lefthook cmd/lefthook/*.go

test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

bin/golangci-lint:
	@test -x $$(go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.43.0

lint: bin/golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run
