## Config file name

Lefthook supports the following file names for the main config:

| Format | File name |
|-------|-----------|
| YAML  | `lefthook.yml` |
| YAML  | `.lefthook.yml` |
| YAML  | `.config/lefthook.yml` |
|-------|-----------|
| YAML  | `lefthook.yaml` |
| YAML  | `.lefthook.yaml` |
| YAML  | `.config/lefthook.yaml` |
|-------|-----------|
| TOML  | `lefthook.toml` |
| TOML  | `.lefthook.toml` |
| TOML  | `.config/lefthook.toml` |
|-------|-----------|
| JSON  | `lefthook.json` |
| JSON  | `.lefthook.json` |
| JSON  | `.config/lefthook.json` |

If there are more than 1 file in the project, only one will be used, and you'll never know which one. So, please, use one format in a project.

Filenames without the leading dot will also be looked up from the [`.config` subdirectory](https://github.com/pi0/config-dir).

Lefthook also merges an extra config with the name `lefthook-local`. All supported formats can be applied to this `-local` config. If you name your main config with the leading dot, like `.lefthook.json`, the `-local` config also must be named with the leading dot: `.lefthook-local.json`.

## Options

- [`assert_lefthook_installed`](./assert_lefthook_installed.md)
- [`colors`](./colors.md)
- [`extends`](./extends.md)
- [`lefthook`](./lefthook.md)
- [`min_version`](./min_version.md)
- [`no_tty`](./no_tty.md)
- [`output`](./output.md)
- [`rc`](./rc.md)
- [`remotes`](./remotes.md)
  - [`git_url`](./git_url.md)
  - [`ref`](./ref.md)
  - [`refetch`](./refetch.md)
  - [`refetch_frequency`](./refetch_frequency.md)
  - [`configs`](./configs.md)
- [`skip_output`](./skip_output.md)
- [`source_dir`](./source_dir.md)
- [`source_dir_local`](./source_dir_local.md)
- [`skip_lfs`](./skip_lfs.md)
- [`templates`](./templates.md)
- [{Git hook name}](./Hook.md) (e.g. `pre-commit`)
  - [`files` (global)](./files-global.md)
  - [`parallel`](./parallel.md)
  - [`piped`](./piped.md)
  - [`follow`](./follow.md)
  - [`exclude_tags`](./exclude_tags.md)
  - [`skip`](./skip.md)
  - [`only`](./only.md)
  - [`jobs`](./jobs.md)
    - [`name`](./name.md)
    - [`run`](./run.md)
    - [`script`](./script.md)
    - [`runner`](./runner.md)
    - [`group`](./group.md)
      - [`parallel`](./parallel.md)
      - [`piped`](./piped.md)
      - [`jobs`](./jobs.md)
    - [`skip`](./skip.md)
    - [`only`](./only.md)
    - [`tags`](./tags.md)
    - [`glob`](./glob.md)
    - [`files`](./files.md)
    - [`file_types`](./file_types.md)
    - [`env`](./env.md)
    - [`root`](./root.md)
    - [`exclude`](./exclude.md)
    - [`fail_text`](./fail_text.md)
    - [`stage_fixed`](./stage_fixed.md)
    - [`interactive`](./interactive.md)
    - [`use_stdin`](./use_stdin.md)
  - [`commands`](./Commands.md)
    - [`run`](./run.md)
    - [`skip`](./skip.md)
    - [`only`](./only.md)
    - [`tags`](./tags.md)
    - [`glob`](./glob.md)
    - [`files`](./files.md)
    - [`file_types`](./file_types.md)
    - [`env`](./env.md)
    - [`root`](./root.md)
    - [`exclude`](./exclude.md)
    - [`fail_text`](./fail_text.md)
    - [`stage_fixed`](./stage_fixed.md)
    - [`interactive`](./interactive.md)
    - [`use_stdin`](./use_stdin.md)
    - [`priority`](./priority.md)
  - [`scripts`](./Scripts.md)
    - [`runner`](./runner.md)
    - [`skip`](./skip.md)
    - [`only`](./only.md)
    - [`tags`](./tags.md)
    - [`env`](./env.md)
    - [`fail_text`](./fail_text.md)
    - [`stage_fixed`](./stage_fixed.md)
    - [`interactive`](./interactive.md)
    - [`use_stdin`](./use_stdin.md)
    - [`priority`](./priority.md)
