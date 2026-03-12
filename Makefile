COMMIT_HASH = $(shell git rev-parse HEAD)

.PHONY: build
build:
	go build -ldflags "-s -w -X github.com/evilmartians/lefthook/v2/internal/version.commit=$(COMMIT_HASH)" -o lefthook

.PHONY: build-with-coverage
build-with-coverage:
	go build -cover -ldflags "-s -w -X github.com/evilmartians/lefthook/v2/internal/version.commit=$(COMMIT_HASH)" -o lefthook

.PHONY: jsonschema
jsonschema:
	go generate gen/jsonschema.go > schema.json
	go generate gen/jsonschema.go > internal/config/jsonschema.json

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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin/ v$$(awk '/^golangci-lint[[:space:]]/ { print $2 }' .tool-versions)

.ONESHELL:
version:
	@read -p "New version: " version
	sed -i "s/const version = .*/const version = \"$$version\"/" internal/version/version.go
	sed -i "s/VERSION = .*/VERSION = \"$$version\";/" packaging/scripts/lib/Constants.rakumod
	sed -i "s/lefthook-plugin.git\", exact: \".*\"/lefthook-plugin.git\", exact: \"$$version\"/" docs/installation/swift.md
	sed -i "s/go install github.com\/evilmartians\/lefthook\/v2.*/go install github.com\/evilmartians\/lefthook\/v2@v$$version/" docs/installation/go.md
	sed -i "s/go install github.com\/evilmartians\/lefthook\/v2.*/go install github.com\/evilmartians\/lefthook\/v2@v$$version/" README.md
	sed -i "s/go get -tool github.com\/evilmartians\/lefthook\/v2.*/go get -tool github.com\/evilmartians\/lefthook\/v2@v$$version/" README.md
	raku packaging/scripts/set-version.raku
	git add internal/version/version.go packaging/* docs/ README.md
