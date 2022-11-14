# Configure lefthook.yml

- [Top level options](#top-level-options)
  - [`colors`](#colors)
  - [`extends`](#extends)
  - [`min_version`](#min_version)
  - [`skip_output`](#skip_output)
  - [`source_dir`](#source_dir)
  - [`source_dir_local`](#source_dir_local)
- [`remote` (Beta :test_tube:)](#remote)
  - [`git_url`](#git_url)
  - [`ref`](#ref)
  - [`config`](#config)
- [Hook](#git-hook)
  - [`files`](#files-global)
  - [`parallel`](#parallel)
  - [`piped`](#piped)
  - [`exclude_tags`](#exclude_tags)
  - [`commands`](#commands)
  - [`scripts`](#scripts)
- [Command](#command)
  - [`run`](#run)
  - [`skip`](#skip)
  - [`tags`](#tags)
  - [`glob`](#glob)
  - [`files`](#files)
  - [`env`](#env)
  - [`root`](#root)
  - [`exclude`](#exclude)
  - [`fail_text`](#fail_text)
  - [`interactive`](#interactive)
- [Script](#script)
  - [`runner`](#runner)
  - [`skip`](#skip)
  - [`tags`](#tags)
  - [`env`](#env)
  - [`fail_text`](#fail_text)
  - [`interactive`](#interactive)
- [Examples](#examples)
- [More info](#more-info)

----

## Top level options

These options are not related to git hooks, and they only control lefthook behavior.

### `colors`

**Default: `true`**

Whether enable or disable colorful output of Lefthook. This option can be overwritten with `--no-colors` option.

**Example**

```yml
# lefthook.yml

colors: false
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

### `skip_output`

You can manage the verbosity using the `skip_output` config. You can set whether lefthook should print some parts of its output.

Possible values are `meta,success,failure,summary,execution`.

This config quiets all outputs except for errors.

**Example**

```yml
# lefthook.yml

skip_output:
  - meta       # Skips lefthook version printing
  - summary    # Skips summary block (successful and failed steps) printing
  - success    # Skips successful steps printing
  - failure    # Skips failed steps printing
  - execution  # Skips printing successfully executed commands and their output (but still prints failed executions)
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

## `remote`

> :test_tube: This feature is in **Beta** version

You can provide a remote config if you want to share your lefthook configuration across many projects. Lefthook will automatically download and merge the configuration into your local `lefthook.yml`.

You can use [`extends`](#extends) related to the config file (not absolute paths).

If you provide [`scripts`](#scripts) in a remote file, the [scripts](#source_dir) folder must be in the **root of the repository**.

**Note**

Configuration in `remote` will be merged to confiuration in `lefthook.yml`, so the priority will be the following:

- `lefthook.yml`
- `remote`
- `lefthook-local.yml`

This can be changed in the future. For convenience, please use `remote` configuration without any hooks configuration in `lefthook.yml`.

### `git_url`

A URL to Git repository. It will be accessed with priveleges of the machine lefthook runs on.

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

An optional *branch* or *tag* name.

**Example**

```yml
# lefthook.yml

remote:
  git_url: git@github.com:evilmartians/lefthook
  ref: v1.0.0
```


### `config`

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

Whether run commands and scripts in concurrently.

### `piped`

**Default: `false`**

Whether run commands and scripts sequentially. Will stop execution if one of the commands/scripts fail.

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

**:warning: Important**: script must exist under `<source_dir>/<git-hook-name>/` folder. See [`source_dir`](#source_dir).

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

#### `{all_files}` template

Simply run `bundle exec rubocop` on all files with `.rb` extension excluding `application.rb` and `routes.rb` files.

**Note:** `--force-exclusion` will apply `Exclude` configuration setting of Rubocop.

```yml
# lefthook.yml

pre-commit:
  commands:
    rubocop:
      tags: backend style
      glob: "*.rb"
      exclude: "application.rb|routes.rb"
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

If you want to have all you files quoted with double quotes `"` or single quotes `'`, quote the appropriate shorthand:

```yml
pre-commit
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

You can skip commands or scripts using `skip` option. You can only skip when merging or rebasing if you want.

**Example**

Always skipping:

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

**Notes**

Always skipping is useful when you have a `lefthook-local.yml` config and you don't want to run some commands locally. So you just overwrite the `skip` option for them to be `true`.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
```

```yml
# lefthook-local.yml

pre-commit:
  commands:
    lint:
      skip: true
```

### `tags`

You can specify tags for commands and scripts. This is useful for [excluding](#exclude_tags). You can specify more than one tag using comma or space.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      tags: frontend,js
      run: yarn lint
    test:
      tags: backend ruby
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

For patterns that you can use see [this](https://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm) reference. We use [glob](https://github.com/gobwas/glob) library and

### `files`

A custom git command for files to be referenced in `{files}` template for [`run`](#run) setting.

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

**Notes**

This option is useful when using lefthook on different OSes or shells where ENV variables are set in different ways.

### `root`

You can change the CWD for the command you execute using `root` option.

This is useful when you execute some `npm` or `yarn` command but the `package.json` is in another directory.

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

You can provide a regular expression to exclude some files from being passed to [`run`](#run) command.

**Example**

Run Rubocop on staged files with `.rb` extension except for `application.rb`, `routes.rb`, and `rails_helper.rb` (wherever they are).

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: ".rb"
      exclude: "application.rb|routes.rb|rails_helper.rb"
      run: bundle exec rubocop --force-exclusion {staged_files}
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

### `interactive`

**Default: `false`**

Whether to use interactive mode and provide a STDIN for a command or script.

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
