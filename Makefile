build:
	go build -tags static,system_libgit2 -o lefthook cmd/lefthook/main.go

test:
	go test  -count=1 -timeout=30s -race ./...

lint:
	golangci-lint run
