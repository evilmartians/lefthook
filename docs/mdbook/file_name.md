## Config file name

Lefthook supports the following file names for the main config:

- `lefthook.yml`
- `.lefthook.yml`
- `lefthook.yaml`
- `.lefthook.yaml`
- `lefthook.toml`
- `.lefthook.toml`
- `lefthook.json`
- `.lefthook.json`

If there are more than 1 file in the project, only one will be used, and you'll never know which one. So, please, use one format in a project.

Lefthook also merges an extra config with the name `lefthook-local`. All supported formats can be applied to this `-local` config. If you name your main config with the leading dot, like `.lefthook.json`, the `-local` config also must be named with the leading dot: `.lefthook-local.json`.

