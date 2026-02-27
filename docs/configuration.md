---
title: "Configuration"
---

# Config file name

Lefthook supports the following file names for the main config:

| Format | Acceptable config names |
|-------|-----------|
| YAML  |`lefthook.yml`<br />`lefthook.yaml`<br />`.lefthook.yml`<br />`.lefthook.yaml`<br />`.config/lefthook.yml`<br />`.config/lefthook.yaml` |
| TOML  | `lefthook.toml` <br />`.lefthook.toml` <br />`.config/lefthook.toml` |
| JSON  | `lefthook.json` <br />`.lefthook.json` <br />`.config/lefthook.json` |
| JSONC | `lefthook.jsonc` <br />`.lefthook.jsonc` <br />`.config/lefthook.jsonc` |

If there are more than 1 file in the project, only one will be used, and you'll never know which one. So, please, use one format in a project.

Filenames without the leading dot will also be looked up from the [`.config` subdirectory](https://github.com/pi0/config-dir).

Lefthook also merges an extra config with the name `lefthook-local`. All supported formats can be applied to this `-local` config. If you name your main config with the leading dot, like `.lefthook.json`, the `-local` config also must be named with the leading dot: `.lefthook-local.json`.

The `-local` config can be used without a main config file. This is useful when you want to use lefthook locally without imposing it on your teammates – just create a `lefthook-local.yml` file and add it to your global `.gitignore`.

