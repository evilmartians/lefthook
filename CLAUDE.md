# CLAUDE.md

See [AGENTS.md](AGENTS.md) for build commands, directory map, and contribution rules.

## Patterns

**Tests** — table-driven with `map[string]struct{ ... }` keyed by description string; use `testify/assert`.

**Errors** — wrap with `fmt.Errorf("context: %w", err)`; use typed errors (structs implementing `error`) when callers need `errors.As`.

**Config structs** — every field needs all four tags: `json:"..." yaml:"..." toml:"..." mapstructure:"..."`. Add `jsonschema` tags for documented options. Run `make jsonschema` after any struct change.

**CLI commands** — return `*cli.Command` from a factory function; action signature is `func(ctx context.Context, cmd *cli.Command) error` (urfave/cli/v3).

**Filesystem** — use `afero.Fs` (never `os` directly) so tests can swap in a MemMapFs.

**Key libs** — `koanf` (config loading/merging), `afero` (FS), `lipgloss`/`spinner` (output), `doublestar` (glob).
