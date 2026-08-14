# AGENTS.md

Lefthook is a CLI-first Git hooks manager. Contributions must be predictable,
backwards-compatible, and dependency-light.

## Requirements

- Go 1.26+ (respect `go.mod` toolchain)
- Git, Make
```
make build            # compile
make test             # unit tests
make test-integration # integration tests
make lint             # golangci-lint
make jsonschema       # regenerate schema.json after config changes
```

## Codebase map

| Path | Purpose |
|---|---|
| `cmd/` | CLI commands |
| `internal/config/` | Config parsing, validation, JSON schema |
| `internal/run/` | Hook runner, parallelism |
| `internal/command/` | Top-level orchestrator |
| `internal/git/` | Git utilities |
| `docs/` | documentation source → lefthook.dev |
| `tests/` | Integration/fixture tests |

## Rules

**Errors** — always wrap with context; never silently ignore; no panic in production paths.

**Concurrency** — no goroutine leaks; use `context.Context`; deterministic output when order matters.

**CLI** — preserve exit codes, flag names, and output format. Update docs and tests for any behavior change.

**Config** — edit structs in `internal/config/`, then run `make jsonschema`. Both `schema.json` and `internal/config/jsonschema.json` must be committed.

**Security** — treat user input as untrusted; no unsafe shell concatenation; sanitize paths.

## Testing

Prefer table-driven unit tests. Integration tests should validate CLI behavior and real git interaction — not internal implementation details.

## PR checklist

- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] Docs updated if behavior changed or new config option added

When in doubt, follow existing patterns. Consistency over cleverness.
