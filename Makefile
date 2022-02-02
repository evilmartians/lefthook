build:
	go build -tags static,system_libgit2 -o lefthook cmd/lefthook/main.go

test:
	go test ./...

lint:
	golangci-lint run
