[Introduction](./intro.md)

# User guide

- [Installation](./installation/README.md)
  - [Ruby](./installation/ruby.md)
  - [Node.js](./installation/node.md)
  - [Swift](./installation/swift.md)
  - [Go](./installation/go.md)
  - [Python](./installation/python.md)
  - [Scoop](./installation/scoop.md)
  - [Homebrew](./installation/homebrew.md)
  - [Winget](./installation/winget.md)
  - [Snap](./installation/snap.md)
  - [Debian-based distro](./installation/deb.md)
  - [RPM-based distro](./installation/rpm.md)
  - [Alpine](./installation/alpine.md)
  - [Arch Linux](./installation/arch.md)
  - [Mise](./installation/mise.md)
  - [Manual](./installation/manual.md)
- [Commands](./usage/commands.md)
  - [lefthook install](./usage/commands/install.md)
  - [lefthook uninstall](./usage/commands/uninstall.md)
  - [lefthook run](./usage/commands/run.md)
  - [lefthook add](./usage/commands/add.md)
  - [lefthook validate](./usage/commands/validate.md)
  - [lefthook dump](./usage/commands/dump.md)
  - [lefthook check-install](./usage/commands/check-install.md)
  - [lefthook self-update](./usage/commands/self-update.md)
- [Features](./usage/features.md)
  - [Local config](./usage/features/local.md)
  - [Pass Git args](./usage/features/git-args.md)
  - [Git LFS](./usage/features/git-lfs.md)
  - [Interactive commands](./usage/features/interactive.md)
  - [Pass STDIN](./usage/features/pass-stdin.md)

# Reference guide

- [Configuration](./configuration/README.md)
  - [`assert_lefthook_installed`](./configuration/assert_lefthook_installed.md)
  - [`colors`](./configuration/colors.md)
  - [`extends`](./configuration/extends.md)
  - [`lefthook`](./configuration/lefthook.md)
  - [`min_version`](./configuration/min_version.md)
  - [`no_tty`](./configuration/no_tty.md)
  - [`output`](./configuration/output.md)
  - [`rc`](./configuration/rc.md)
  - [`remotes`](./configuration/remotes.md)
    - [`git_url`](./configuration/git_url.md)
    - [`ref`](./configuration/ref.md)
    - [`refetch`](./configuration/refetch.md)
    - [`refetch_frequency`](./configuration/refetch_frequency.md)
    - [`configs`](./configuration/configs.md)
  - [`skip_output`](./configuration/skip_output.md)
  - [`source_dir`](./configuration/source_dir.md)
  - [`source_dir_local`](./configuration/source_dir_local.md)
  - [`skip_lfs`](./configuration/skip_lfs.md)
  - [`templates`](./configuration/templates.md)
  - [{Git hook name}](./configuration/Hook.md)
    - [`files`](./configuration/files-global.md)
    - [`parallel`](./configuration/parallel.md)
    - [`piped`](./configuration/piped.md)
    - [`follow`](./configuration/follow.md)
    - [`exclude_tags`](./configuration/exclude_tags.md)
    - [`skip`](./configuration/skip.md)
    - [`only`](./configuration/only.md)
    - [`jobs`](./configuration/jobs.md)
      - [`name`](./configuration/name.md)
      - [`run`](./configuration/run.md)
      - [`script`](./configuration/script.md)
      - [`runner`](./configuration/runner.md)
      - [`group`](./configuration/group.md)
        - [`parallel`](./configuration/parallel.md)
        - [`piped`](./configuration/piped.md)
        - [`jobs`](./configuration/jobs.md)
      - [`skip`](./configuration/skip.md)
      - [`only`](./configuration/only.md)
      - [`tags`](./configuration/tags.md)
      - [`glob`](./configuration/glob.md)
      - [`files`](./configuration/files.md)
      - [`file_types`](./configuration/file_types.md)
      - [`env`](./configuration/env.md)
      - [`root`](./configuration/root.md)
      - [`exclude`](./configuration/exclude.md)
      - [`fail_text`](./configuration/fail_text.md)
      - [`stage_fixed`](./configuration/stage_fixed.md)
      - [`interactive`](./configuration/interactive.md)
      - [`use_stdin`](./configuration/use_stdin.md)
    - [`commands`](./configuration/Commands.md)
      - [`run`](./configuration/run.md)
      - [`skip`](./configuration/skip.md)
      - [`only`](./configuration/only.md)
      - [`tags`](./configuration/tags.md)
      - [`glob`](./configuration/glob.md)
      - [`files`](./configuration/files.md)
      - [`file_types`](./configuration/file_types.md)
      - [`env`](./configuration/env.md)
      - [`root`](./configuration/root.md)
      - [`exclude`](./configuration/exclude.md)
      - [`fail_text`](./configuration/fail_text.md)
      - [`stage_fixed`](./configuration/stage_fixed.md)
      - [`interactive`](./configuration/interactive.md)
      - [`use_stdin`](./configuration/use_stdin.md)
      - [`priority`](./configuration/priority.md)
    - [`scripts`](./configuration/Scripts.md)
      - [`runner`](./configuration/runner.md)
      - [`skip`](./configuration/skip.md)
      - [`only`](./configuration/only.md)
      - [`tags`](./configuration/tags.md)
      - [`env`](./configuration/env.md)
      - [`fail_text`](./configuration/fail_text.md)
      - [`stage_fixed`](./configuration/stage_fixed.md)
      - [`interactive`](./configuration/interactive.md)
      - [`use_stdin`](./configuration/use_stdin.md)
      - [`priority`](./configuration/priority.md)
- [ENV variables](./usage/env.md)
  - [LEFTHOOK](./usage/envs/LEFTHOOK.md)
  - [LEFTHOOK_VERBOSE](./usage/envs/LEFTHOOK_VERBOSE.md)
  - [LEFTHOOK_OUTPUT](./usage/envs/LEFTHOOK_OUTPUT.md)
  - [LEFTHOOK_CONFIG](./usage/envs/LEFTHOOK_CONFIG.md)
  - [LEFTHOOK_EXCLUDE](./usage/envs/LEFTHOOK_EXCLUDE.md)
  - [CLICOLOR_FORCE](./usage/envs/CLICOLOR_FORCE.md)
  - [NO_COLOR](./usage/envs/NO_COLOR.md)
  - [CI](./usage/envs/CI.md)

# Examples

- [Using local only config](./examples/lefthook-local.md)
- [Wrap commands locally](./examples/wrap-commands.md)
- [Auto add linter fixes to commit](./examples/stage_fixed.md)
- [Filter files](./examples/filters.md)
- [Skip or run on condition](./examples/skip.md)
- [Remote configs](./examples/remotes.md)
- [With commitlint](./examples/commitlint.md)

---

[Contributors](./misc/contributors.md)
