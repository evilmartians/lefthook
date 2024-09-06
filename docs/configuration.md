# Configure lefthook

Lefthook [supports](#config-file) YAML, JSON, and TOML configuration. In this document `lefthook.yml` is used for simplicity.

- [Config file](#config-file)
- [Top level options](#top-level-options)
  - [`assert_lefthook_installed`](#assert_lefthook_installed)
  - [`colors`](#colors)
  - [`no_tty`](#no_tty)
  - [`extends`](#extends)
  - [`min_version`](#min_version)
  - [`skip_output`](#skip_output)
  - [`source_dir`](#source_dir)
  - [`source_dir_local`](#source_dir_local)
  - [`rc`](#rc)
- [`remote`](#remote--deprecated-show-remotes-instead) :warning: **DEPRECATED** use [`remotes`](#remotes)
  - [`git_url`](#git_url)
  - [`ref`](#ref)
  - [`config`](#config)
- [`remotes`](#remotes)
  - [`git_url`](#git_url-1)
  - [`ref`](#ref-1)
  - [`refetch`](#refetch)
  - [`configs`](#configs)
- [Git hook](#git-hook)
  - [`files` (global)](#files-global)
  - [`parallel`](#parallel)
  - [`piped`](#piped)
  - [`follow`](#follow)
  - [`exclude_tags`](#exclude_tags)
  - [`commands`](#commands)
  - [`scripts`](#scripts)
- [Command](#command)
  - [`run`](#run)
    - [`{files}` template](#files-template)
    - [`{staged_files}` template](#staged_files-template)
    - [`{push_files}` template](#push_files-template)
    - [`{all_files}` template](#all_files-template)
    - [`{cmd}` template](#cmd-template)
  - [`skip`](#skip)
  - [`only`](#only)
  - [`tags`](#tags)
  - [`glob`](#glob)
  - [`files`](#files)
  - [`file_types`](#file_types)
  - [`env`](#env)
  - [`root`](#root)
  - [`exclude`](#exclude)
  - [`fail_text`](#fail_text)
  - [`stage_fixed`](#stage_fixed)
  - [`interactive`](#interactive)
  - [`use_stdin`](#use_stdin)
  - [`priority`](#priority)
- [Script](#script)
  - [`use_stdin`](#use_stdin)
  - [`runner`](#runner)
  - [`skip`](#skip)
  - [`only`](#only)
  - [`tags`](#tags)
  - [`env`](#env)
  - [`fail_text`](#fail_text)
  - [`stage_fixed`](#stage_fixed)
  - [`interactive`](#interactive)
  - [`use_stdin`](#use_stdin)
  - [`priority`](#priority)
- [Examples](#examples)
- [More info](#more-info)

----

## Config file

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

## Top level options

These options are not related to git hooks, and they only control lefthook behavior.

### `assert_lefthook_installed`

**Default: `false`**

When set to `true`, fail (with exit status 1) if `lefthook` executable can't be found in $PATH, under node_modules/, as a Ruby gem, or other supported method. This makes sure git hook won't omit `lefthook` rules if `lefthook` ever was installed.

### `colors`

**Default: `auto`**

Whether enable or disable colorful output of Lefthook. This option can be overwritten with `--colors` option. You can also provide your own color codes.

**Example**

Disable colors.

```yml
# lefthook.yml

colors: false
```

Custom color codes. Can be hex or ANSI codes.

```yml
# lefthook.yml

colors:
  cyan: 14
  gray: 244
  green: '#32CD32'
  red: '#FF1493'
  yellow: '#F0E68C'
```

### `no_tty`

**Default: `false`**

Whether hide spinner and other interactive things. This can be also controlled with `--no-tty` option for `lefthook run` command.

**Example**

```yml
# lefthook.yml

no_tty: true
```

### `extends`

You can extend your config with another one YAML file. Its content will be merged. Extends for `lefthook.yml`, `lefthook-local.yml`, and [`remote`](#remote) configs are handled separately, so you can have different extends in these files.

**Example**

```yml
# lefthook.yml

extends:
  - /home/user/work/lefthook-extend.yml
  - /home/user/work/lefthook-extend-2.yml
  - lefthook-extends/file.yml
  - ../extend.yml
```

### `min_version`

If you want to specify a minimum version for lefthook binary (e.g. if you need some features older versions don't have) you can set this option.

**Example**

```yml
# lefthook.yml

min_version: 1.1.3
```

### `output`

You can manage verbosity using the `output` config. You can specify what to print in your output by setting these values, which you need to have

Possible values are `meta,summary,success,failure,execution,execution_out,execution_info,skips`.
By default, all output values are enabled

You can also disable all output with setting `output: false`. In this case only errors will be printed.

This config quiets all outputs except for errors.

`output` is enabled if there is no `skip_output` and `LEFTHOOK_QUIET`.

**Example**

```yml
# lefthook.yml

output:
  - meta           # Print lefthook version
  - summary        # Print summary block (successful and failed steps)
  - empty_summary  # Print summary heading when there are no steps to run
  - success        # Print successful steps
  - failure        # Print failed steps printing
  - execution      # Print any execution logs (but prints if the execution failed)
  - execution_out  # Print execution output (but still prints failed commands output)
  - execution_info # Print `EXECUTE > ...` logging
  - skips          # Print "skip" (i.e. no files matched)
```

You can also *extend* this list with an environment variable `LEFTHOOK_OUTPUT`:

```bash
LEFTHOOK_OUTPUT="meta,success,summary" lefthook run pre-commit
```

### `skip_output`

> [!IMPORTANT]
> **DEPRECATED** This feature is deprecated and might be removed in future versions. Please, use `[output]` instead for managing verbosity.

You can manage the verbosity using the `skip_output` config. You can set whether lefthook should print some parts of its output.

Possible values are `meta,summary,success,failure,execution,execution_out,execution_info,skips`.

You can also disable all output with setting `skip_output: true`. In this case only errors will be printed.

This config quiets all outputs except for errors.

**Example**

```yml
# lefthook.yml

skip_output:
  - meta           # Skips lefthook version printing
  - summary        # Skips summary block (successful and failed steps) printing
  - empty_summary  # Skips summary heading when there are no steps to run
  - success        # Skips successful steps printing
  - failure        # Skips failed steps printing
  - execution      # Skips printing any execution logs (but prints if the execution failed)
  - execution_out  # Skips printing execution output (but still prints failed commands output)
  - execution_info # Skips printing `EXECUTE > ...` logging
  - skips          # Skips "skip" printing (i.e. no files matched)
```

You can also *extend* this list with an environment variable `LEFTHOOK_QUIET`:

```bash
LEFTHOOK_QUIET="meta,success,summary" lefthook run pre-commit
```

### `source_dir`

**Default: `.lefthook/`**

Change a directory for script files. Directory for script files contains folders with git hook names which contain script files.

Example of directory tree:

```
.lefthook/
├── pre-commit/
│   ├── lint.sh
│   └── test.py
└── pre-push/
    └── check-files.rb
```

### `source_dir_local`

**Default: `.lefthook-local/`**

Change a directory for *local* script files (not stored in VCS).

This option is useful if you have a `lefthook-local.yml` config file and want to reference different scripts there.

### `rc`

Provide an [**rc**](https://www.baeldung.com/linux/rc-files) file, which is actually a simple `sh` script. Currently it can be used to set ENV variables that are not accessible from non-shell programs.

**Example**

Use cases:

- You have a GUI program that runs git hooks (e.g., VSCode)
- You reference executables that are accessible only from a tweaked $PATH environment variable (e.g., when using rbenv or nvm)
- Or even if your GUI program cannot locate the `lefthook` executable :scream:
- Or if you want to use ENV variables that control the executables behavior in `lefthook.yml`

```bash
# An npm executable which is managed by nvm
$ which npm
/home/user/.nvm/versions/node/v15.14.0/bin/npm
```

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: npm run eslint {staged_files}
```

Provide a tweak to access `npm` executable the same way you do it in your ~/<shell>rc.

```yml
# lefthook-local.yml

# You can choose whatever name you want.
# You can share it between projects where you use lefthook.
# Make sure the path is absolute.
rc: ~/.lefthookrc
```

Or

```yml
# lefthook-local.yml

# If the path contains spaces, you need to quote it.
rc: '"${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc"'
```

In the rc file, export any new environment variables or modify existing ones.

```bash
# ~/.lefthookrc

# An nvm way
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Or maybe just
PATH=$PATH:$HOME/.nvm/versions/node/v15.14.0/bin
```

```bash
# Make sure you updated git hooks. This is important.
$ lefthook install -f
```

Now any program that runs your hooks will have a tweaked PATH environment variable and will be able to get `nvm` :wink:

## `remote`

> [!WARNING]
> **DEPRECATED** Use [`remotes`](#remotes) setting

You can provide a remote config if you want to share your lefthook configuration across many projects. Lefthook will automatically download and merge the configuration into your local `lefthook.yml`.

You can use [`extends`](#extends) related to the config file (not absolute paths).

If you provide [`scripts`](#scripts) in a remote file, the [scripts](#source_dir) folder must be in the **root of the repository**.

**Note**

Configuration in `remote` will be merged to configuration in `lefthook.yml`, so the priority will be the following:

- `lefthook.yml`
- `remote`
- `lefthook-local.yml`

This can be changed in the future. For convenience, please use `remote` configuration without any hooks configuration in `lefthook.yml`.

### `git_url`

> [!WARNING]
> **DEPRECATED** Use [`remotes`](#remotes) setting

A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on.

**Example**

```yml
# lefthook.yml

remote:
  git_url: git@github.com:evilmartians/lefthook
```

Or

```yml
# lefthook.yml

remote:
  git_url: https://github.com/evilmartians/lefthook
```

### `ref`

> [!WARNING]
> **DEPRECATED** Use [`remotes`](#remotes) setting

An optional *branch* or *tag* name.

**Example**

```yml
# lefthook.yml

remote:
  git_url: git@github.com:evilmartians/lefthook
  ref: v1.0.0
```

> [!CAUTION]
>
> If you initially had `ref` option, ran `lefthook install`, and then removed it, lefthook won't decide which branch/tag to use as a ref. So, if you added it once, please, use it always to avoid issues in local setups.

### `config`

> [!WARNING]
> **DEPRECATED**. Use [`remotes`](#remotes) setting

**Default:** `lefthook.yml`

An optional config path from remote's root.

**Example**

```yml
# lefthook.yml

remote:
  git_url: git@github.com:evilmartians/remote
  ref: v1.0.0
  config: examples/ruby-linter.yml
```

## `remotes`

> [!IMPORTANT]
> :test_tube: This feature is in **Beta** version

You can provide multiple remote configs if you want to share yours lefthook configurations across many projects. Lefthook will automatically download and merge configurations into your local `lefthook.yml`.

You can use [`extends`](#extends) but the paths must be relative to the remote repository root.

If you provide [`scripts`](#scripts) in a remote config file, the [scripts](#source_dir) folder must also be in the **root of the repository**.

**Note**

The configuration from `remotes` will be merged to the local config using the following priority:

1. Local main config (`lefthook.yml`)
1. Remote configs (`remotes`)
1. Local overrides (`lefthook-local.yml`)

This priority may be changed in the future. For convenience, if you use `remotes`, please don't configure any hooks.

### `git_url`

A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
```

Or

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
```

### `ref`

An optional *branch* or *tag* name.

> [!NOTE]
>
> If you initially had `ref` option, ran `lefthook install`, and then removed it, lefthook won't decide which branch/tag to use as a ref. So, if you added it once, please, use it always to avoid issues in local setups.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
```

### `refetch`

**Default:** `false`

Force remote config refetching on every run. Lefthook will be refetching the specified remote every time it is called.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch: true
```

### `configs`

**Default:** `[lefthook.yml]`

An optional array of config paths from remote's root.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
      - examples/test.yml
```

Example with multiple remotes merging multiple configurations.

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:org/lefthook-configs
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
      - examples/test.yml
  - git_url: https://github.com/org2/lefthook-configs
    configs:
      - lefthooks/pre_commit.yml
      - lefthooks/post_merge.yml
  - git_url: https://github.com/org3/lefthook-configs
    ref: feature/new
    configs:
      - configs/pre-push.yml

```

## Git hook

Commands and scripts are defined for git hooks. You can defined a hook for all hooks listed in [this file](../internal/config/available_hooks.go).

### `files` (global)

A custom git command for files to be referenced in `{files}` template. See [`run`](#run) and [`files`](#files).

If the result of this command is empty, the execution of commands will be skipped.

**Example**

```yml
# lefthook.yml

pre-commit:
  files: git diff --name-only master # custom list of files
  commands:
    ...
```

### `parallel`

**Default: `false`**

> [!NOTE]
>
> Lefthook runs commands and scripts **sequentially** by default.

Run commands and scripts concurrently.

### `piped`

**Default: `false`**

> [!NOTE]
>
> Lefthook will return an error if both `piped: true` and `parallel: true` are set.

Stop running commands and scripts if one of them fail.

**Example**

```yml
# lefthook.yml

database:
  piped: true # Stop if one of the steps fail
  commands:
    1_create:
      run: rake db:create
    2_migrate:
      run: rake db:migrate
    3_seed:
      run: rake db:seed
```

### `follow`

**Default: `false`**

Follow the STDOUT of the running commands and scripts.

**Example**

```yml
# lefthook.yml

pre-push:
  follow: true
  commands:
    backend-tests:
      run: bundle exec rspec
    frontend-tests:
      run: yarn test
```

> [!NOTE]
>
> If used with [`parallel`](#parallel) the output can be a mess, so please avoid setting both options to `true`.

### `exclude_tags`

[Tags](#tags) or command names that you want to exclude. This option can be overwritten with `LEFTHOOK_EXCLUDE` env variable.

**Example**

```yml
# lefthook.yml

pre-commit:
  exclude_tags: frontend
  commands:
    lint:
      tag: frontend
      ...
    test:
      tag: frontend
      ...
    check-syntax:
      tag: documentation
```

```bash
lefthook run pre-commit # will only run check-syntax command
```

**Notes**

This option is good to specify in `lefthook-local.yml` when you want to skip some execution locally.

```yml
# lefthook.yml

pre-push:
  commands:
    packages-audit:
      tags: frontend security
      run: yarn audit
    gems-audit:
      tags: backend security
      run: bundle audit
```

You can skip commands by tags:

```yml
# lefthook-local.yml

pre-push:
  exclude_tags:
    - frontend
```

### `commands`

Commands to be executed for the hook. Each command has a name and associated run [options](#command).

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      ... # command options
```

### `scripts`

Scripts to be executed for the hook. Each script has a name (filename in scripts dir) and associated run [options](#script).

> [!IMPORTANT]
> Script must exist under `<source_dir>/<git-hook-name>/` folder. See [`source_dir`](#source_dir).

**Example**

```yml
# lefthook.yml

pre-commit:
  scripts:
    "lint.sh":
      ... # script options
```

Correct folders structure:
```
.lefthook/
└── pre-commit/
    └── lint.sh
```

## Command

### `run`

This is a mandatory option for a command. This is actually a command that is executed for the hook.

You can use files templates that will be substituted with the appropriate files on execution:

- `{files}` - custom [`files`](#files) command result.
- `{staged_files}` - staged files which you try to commit.
- `{push_files}` - files that are committed but not pushed.
- `{all_files}` - all files tracked by git.
- `{cmd}` - shorthand for the command from `lefthook.yml`.
- `{0}` - shorthand for the single space-joint string of git hook arguments.
- `{N}` - shorthand for the N-th git hook argument.

**Example**

Run `yarn lint` on `pre-commit` hook.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
```

#### `{files}` template

Run `go vet` only on files listed with `git ls-files -m` command with `.go` extension.

```yml
# lefthook.yml

pre-commit:
  commands:
    govet:
      files: git ls-files -m
      glob: "*.go"
      run: go vet {files}
```

#### `{staged_files}` template

Run `yarn eslint` only on staged files with `.js`, `.ts`, `.jsx`, and `.tsx` extensions.

```yml
# lefthook.yml

pre-commit:
  commands:
    eslint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {staged_files}
```

#### `{push_files}` template

If you want to lint files only before pushing them.

```yml
# lefthook.yml

pre-push:
  commands:
    eslint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {push_files}
```

#### `{all_files}` template

Simply run `bundle exec rubocop` on all files with `.rb` extension excluding `application.rb` and `routes.rb` files.

> [!NOTE]
>
> `--force-exclusion` will apply `Exclude` configuration setting of Rubocop.

```yml
# lefthook.yml

pre-commit:
  commands:
    rubocop:
      tags: backend style
      glob: "*.rb"
      exclude:
        - config/application.rb
        - config/routes.rb
      run: bundle exec rubocop --force-exclusion {all_files}
```

#### `{cmd}` template

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
  scripts:
    "good_job.js":
      runner: node
```

You can wrap it in docker runner locally:

```yml
# lefthook-local.yml

pre-commit:
  commands:
    lint:
      run: docker run -it --rm <container_id_or_name> {cmd}
  scripts:
    "good_job.js":
      runner: docker run -it --rm <container_id_or_name> {cmd}
```

#### Git arguments

Make sure commits are signed.

```yml
# lefthook.yml

# Note: commit-msg hook takes a single parameter,
#       the name of the file that holds the proposed commit log message.
# Source: https://git-scm.com/docs/githooks#_commit_msg
commit-msg:
  commands:
    multiple-sign-off:
      run: 'test $(grep -c "^Signed-off-by: " {1}) -lt 2'
```

**Notes**

#### Rubocop

If using `{all_files}` with RuboCop, it will ignore RuboCop's `Exclude` configuration setting. To avoid this, pass `--force-exclusion`.

#### Quotes

If you want to have all your files quoted with double quotes `"` or single quotes `'`, quote the appropriate shorthand:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.js"
      # Quoting with double quotes `"` might be helpful for Windows users
      run: yarn eslint "{staged_files}" # will run `yarn eslint "file1.js" "file2.js" "[strange name].js"`
    test:
      glob: "*.{spec.js}"
      run: yarn test '{staged_files}' # will run `yarn eslint 'file1.spec.js' 'file2.spec.js' '[strange name].spec.js'`
    format:
      glob: "*.js"
      # Will quote where needed with single quotes
      run: yarn test {staged_files} # will run `yarn eslint file1.js file2.js '[strange name].spec.js'`
```

### `skip`

You can skip all or specific commands and scripts using `skip` option. You can also skip when merging, rebasing, or being on a specific branch. Globs are available for branches.

**Example**

Always skipping a command:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: true
      run: yarn lint
```

Skipping on merging and rebasing:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip:
        - merge
        - rebase
      run: yarn lint
```

Or

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: merge
      run: yarn lint
```

Skipping the whole hook on `main` branch:

```yml
# lefthook.yml

pre-commit:
  skip:
    - ref: main
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

Skipping hook for all `dev/*` branches:

```yml
# lefthook.yml

pre-commit:
  skip:
    - ref: dev/*
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

Skipping hook by running a command:

```yml
# lefthook.yml

pre-commit:
  skip:
    - run: test "${NO_HOOK}" -eq 1
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

> [!TIP]
>
> Always skipping is useful when you have a `lefthook-local.yml` config and you don't want to run some commands locally. So you just overwrite the `skip` option for them to be `true`.
>
> ```yml
> # lefthook.yml
>
> pre-commit:
>   commands:
>     lint:
>       run: yarn lint
> ```
>
> ```yml
> # lefthook-local.yml
>
> pre-commit:
>   commands:
>     lint:
>       skip: true
> ```

### `only`

You can force a command, script, or the whole hook to execute only in certain conditions. This option acts like the opposite of [`skip`](#skip). It accepts the same values but skips execution only if the condition is not satisfied.

> [!NOTE]
>
> `skip` option takes precedence over `only` option, so if you have conflicting conditions the execution will be skipped.

**Example**

Execute a hook only for `dev/*` branches.

```yml
# lefthook.yml

pre-commit:
  only:
    - ref: dev/*
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

When rebasing execute quick linter but skip usual linter and tests.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: rebase
      run: yarn lint
    test:
      skip: rebase
      run: yarn test
    lint-on-rebase:
      only: rebase
      run: yarn lint-quickly
```

### `tags`

You can specify tags for commands and scripts. This is useful for [excluding](#exclude_tags). You can specify more than one tag using comma.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      tags: frontend,js
      run: yarn lint
    test:
      tags: backend,ruby
      run: bundle exec rspec
```

### `glob`

You can set a glob to filter files for your command. This is only used if you use a file template in [`run`](#run) option or provide your custom [`files`](#files) command.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {staged_files}
```

**Notes**

For patterns that you can use see [this](https://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm) reference. We use [glob](https://github.com/gobwas/glob) library.

If you've specified `glob` but don't have a files template in [`run`](#run) option, lefthook will check `{staged_files}` for `pre-commit` hook and `{push_files}` for `pre-push` hook and apply filtering. If no files left, the command will be skipped.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.js"
      run: npm run lint # skipped if no .js files staged
```

### `files`

A custom git command for files or directories to be referenced in `{files}` template for [`run`](#run) setting.

If the result of this command is empty, the execution of commands will be skipped.

This option overwrites the [hook-level `files`](#files-global) option.

**Example**

Provide a git command to list files.

```yml
# lefthook.yml

pre-push:
  commands:
    stylelint:
      tags: frontend style
      files: git diff --name-only master
      glob: "*.js"
      run: yarn stylelint {files}
```

Call a custom script for listing files.

```yml
# lefthook.yml

pre-push:
  commands:
    rubocop:
      tags: backend
      glob: "**/*.rb"
      files: node ./lefthook-scripts/ls-files.js # you can call your own scripts
      run: bundle exec rubocop --force-exclusion --parallel {files}
```

### `file_types`

Filter files in a [`run`](#run) templates by their type. Supported types:

|File type| Exlanation|
|---------|-----------|
|`text`   | Any file that contains text. Symlinks are not followed. |
|`binary` | Any file that contains non-text bytes. Symlinks are not followed. |
|`executable` | Any file that has executable bits set. Symlinks are not followed. |
|`not executable` | Any file without executable bits in file mode. Symlinks included. |
|`symlink` | A symlink file. |
|`not symlink` | Any non-symlink file. |

> [!IMPORTANT]
> When passed multiple file types all constraints will be applied to the resulting list of files.

**Examples**

Apply some different linters on text and binary files.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint-code:
      run: yarn lint {staged_files}
      file_types: text
    check-hex-codes:
      run: yarn check-hex {staged_files}
      file_types: binary
```

Skip symlinks.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix {staged_files}
      file_types:
        - not symlink
```

Lint executable scripts.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix {staged_files}
      file_types:
        - executable
        - text
```

### `env`

You can specify some ENV variables for the command or script.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      env:
        RAILS_ENV: test
      run: bundle exec rspec
```

#### Extending PATH

If your hook is run by GUI program, and you use some PATH tweaks in your ~/.<shell>rc, you might see an error saying *executable not found*. In that case You can extend the **$PATH** variable with `lefthook-local.yml` configuration the following way.

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      run: yarn test
```

```yml
# lefthook-local.yml

pre-commit:
  commands:
    test:
      env:
        PATH: $PATH:/home/me/path/to/yarn
```

**Notes**

This option is useful when using lefthook on different OSes or shells where ENV variables are set in different ways.

### `root`

You can change the CWD for the command you execute using `root` option.

This is useful when you execute some `npm` or `yarn` command but the `package.json` is in another directory.

For `pre-push` and `pre-commit` hooks and for the custom `files` command `root` option is used to filter file paths. If all files are filtered the command will be skipped.

**Example**

Format and stage files from a `client/` folder.

```bash
# Folders structure

$ tree .
.
├── client/
│   ├── package.json
│   ├── node_modules/
|   ├── ...
├── server/
|   ...
```

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      root: "client/"
      glob: "*.{js,ts}"
      run: yarn eslint --fix {staged_files} && git add {staged_files}
```

### `exclude`

For the `exclude` option two variants are supported:

- A list of globs to be excluded
- A single regular expression (deprecated)


> [!NOTE]
>
> The regular expression is matched against full paths to files in the repo,
> relative to the repo root, using `/` as the directory separator on all platforms.
> File paths do not begin with the separator or any other prefix.

**Example**

Run Rubocop on staged files with `.rb` extension except for `application.rb`, `routes.rb`, `rails_helper.rb`, and all Ruby files in `config/initializers/`.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.rb"
      exclude:
        - config/routes.rb
        - config/application.rb
        - config/initializers/*.rb
        - spec/rails_helper.rb
      run: bundle exec rubocop --force-exclusion {staged_files}
```

The same example using a regular expression.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.rb"
      exclude: '(^|/)(application|routes|rails_helper|initializers/\w+)\.rb$'
      run: bundle exec rubocop --force-exclusion {staged_files}
```

**Notes**

Be careful with the config file format's string quoting and escaping rules when writing regexps in it. For YAML, single quotes are often the simplest choice.

If you've specified `exclude` but don't have a files template in [`run`](#run) option, lefthook will check `{staged_files}` for `pre-commit` hook and `{push_files}` for `pre-push` hook and apply filtering. If no files left, the command will be skipped.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      exclude: '(^|/)application\.rb$'
      run: bundle exec rubocop # skipped if only application.rb was staged
```

### `fail_text`

You can specify a text to show when the command or script fails.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
      fail_text: Add node executable to $PATH
```

```bash
$ git commit -m 'fix: Some bug'

Lefthook v1.1.3
RUNNING HOOK: pre-commit

  EXECUTE > lint

SUMMARY: (done in 0.01 seconds)
🥊  lint: Add node executable to $PATH env
```

### `stage_fixed`

**Default: `false`**

> Used **only for `pre-commit`** hook. Is ignored for other hooks.

When set to `true` lefthook will automatically call `git add` on files after running the command or script. For a command if [`files`](#files) option was specified, the specified command will be used to retrieve files for `git add`. For scripts and commands without [`files`](#files) option `{staged_files}` template will be used. All filters ([`glob`](#glob), [`exclude`](#exclude)) will be applied if specified.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: npm run lint --fix {staged_files}
      stage_fixed: true
```

### `interactive`

**Default: `false`**

> [!NOTE]
>
> If you want to pass stdin to your command or script but don't need to get the input from CLI, use [`use_stdin`](#use_stdin) option instead.


Whether to use interactive mode. This applies the certain behavior:
- All `interactive` commands/scripts are executed after non-interactive. Exception: [`piped`](#piped) option is set to `true`.
- When executing, lefthook tries to open /dev/tty (Linux/Unix only) and use it as stdin.
- When [`no_tty`](#no_tty) option is set, `interactive` is ignored.

### `priority`

**Default: `0`**

> [!NOTE]
>
> This option makes sense only when `parallel: false` or `piped: true` is set.
>
> Value `0` is considered an `+Infinity`, so commands or scripts with `priority: 0` or without this setting will be run at the very end.

Set priority from 1 to +Infinity. This option can be used to configure the order of the sequential steps.

**Example**

```yml
# lefthook.yml

post-checkout:
  piped: true
  commands:
    db-create:
      priority: 1
      run: rails db:create
    db-migrate:
      priority: 2
      run: rails db:migrate
    db-seed:
      priority: 3
      run: rails db:seed

  scripts:
    "check-spelling.sh":
      runner: bash
      priority: 1
    "check-grammar.rb":
      runner: ruby
      priority: 2
```

## Script

Scripts are stored under `<source_dir>/<hook-name>/` folder. These scripts are your own executables which are being run in the project root (if you don't specify a [`root`](#root) option).

To add a script for a `pre-commit` hook:

1. Run `lefthook add -d pre-commit`
1. Edit `.lefthook/pre-commit/my-script.sh`
1. Add an entry to `lefthook.yml`
   ```yml
   # lefthook.yml

   pre-commit:
     scripts:
       "my-script.sh":
         runner: bash
   ```

**Example**

Let's create a bash script to check commit templates `.lefthook/commit-msg/template_checker`:

```bash
INPUT_FILE=$1
START_LINE=`head -n1 $INPUT_FILE`
PATTERN="^(TICKET)-[[:digit:]]+: "
if ! [[ "$START_LINE" =~ $PATTERN ]]; then
  echo "Bad commit message, see example: TICKET-123: some text"
  exit 1
fi
```

Now we can ask lefthook to run our bash script by adding this code to
`lefthook.yml` file:

```yml
# lefthook.yml

commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

When you try to commit `git commit -m "bad commit text"` script `template_checker` will be executed. Since commit text doesn't match the described pattern the commit process will be interrupted.

### `use_stdin`

> [!NOTE]
>
> With many commands or scripts having `use_stdin: true`, only one will receive the data. The others will have nothing. If you need to pass the data from stdin to every command or script, please, submit a [feature request](https://github.com/evilmartians/lefthook/issues/new?assignees=&labels=feature+request&projects=&template=feature_request.md).

Pass the stdin from the OS to the command/script.

**Example**

Use this option for the `pre-push` hook when you have a script that does `while read ...`. Without this option lefthook will hang: lefthook uses [pseudo TTY](https://github.com/creack/pty) by default, and it doesn't close stdin when all data is read.

```bash
# .lefthook/pre-push/do-the-magic.sh

remote="$1"
url="$2"

while read local_ref local_oid remote_ref remote_oid; do
  # ...
done
```

```yml
# lefthook.yml
pre-push:
  scripts:
    "do-the-magic.sh":
      runner: bash
      use_stdin: true
```

### `runner`

You should specify a runner for the script. This is a command that should execute a script file. It will be called the following way: `<runner> <path-to-script>` (e.g. `ruby .lefthook/pre-commit/lint.rb`).

**Example**

```yml
# lefthook.yml

pre-commit:
  scripts:
    "lint.js":
      runner: node
    "check.go":
      runner: go run
```

## Examples

We have a directory with few examples. You can check it [here](https://github.com/evilmartians/lefthook/tree/master/examples).

## More info

Have a question?

:monocle_face: Check the [wiki](https://github.com/evilmartians/lefthook/wiki)

:thinking: Or start a [discussion](https://github.com/evilmartians/lefthook/discussions)
