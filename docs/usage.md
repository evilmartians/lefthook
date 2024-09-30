# Usage

- [Commands](#commands)
  - [`lefthook install`](#lefthook-install)
  - [`lefthook uninstall`](#lefthook-uninstall)
  - [`lefthook add`](#lefthook-add)
  - [`lefthook run`](#lefthook-run)
  - [`lefthook version`](#lefthook-version)
  - [`lefthook self-update`](#lefthook-self-update)
- [ENV variables](#env-variables)
  - [`LEFTHOOK`](#lefthook)
  - [`LEFTHOOK_EXCLUDE`](#lefthook_exclude)
  - [`LEFTHOOK_OUTPUT`](#lefthook_output)
  - [`LEFTHOOK_QUIET`](#lefthook_quiet)
  - [`LEFTHOOK_VERBOSE`](#lefthook_verbose)
  - [`LEFTHOOK_BIN`](#lefthook_bin)
  - [`NO_COLOR`](#no_color)
  - [`CLICOLOR_FORCE`](#clicolor_force)
- [Tips](#tips)
  - [Local config](#local-config)
  - [Disable lefthook in CI](#disable-lefthook-in-ci)
  - [Commitlint example](#commitlint-example)
  - [Parallel execution](#parallel-execution)
  - [Concurrent files overrides](#concurrent-files-overrides)
  - [Capture ARGS from git in the script](#capture-args-from-git-in-the-script)
  - [Git LFS support](#git-lfs-support)
  - [Pass stdin to a command or script](#pass-stdin-to-a-command-or-script)
  - [Using an interactive command or script](#using-an-interactive-command-or-script)

----

## Commands

Use `lefthook help` or `lefthook <command> -h/--help` to discover available commands and their options.

### `lefthook install`

`lefthook install` creates an empty `lefthook.yml` if a configuration file does not exist and updates git hooks to use lefthook. Run `lefthook install` after cloning the git repo.

> [!NOTE]
>
> NPM package `lefthook` installs the hooks in a postinstall script automatically.

### `lefthook uninstall`

`lefthook uninstall` clears git hooks affected by lefthook.

### `lefthook add`

`lefthook add pre-commit` will create a file `.git/hooks/pre-commit`. This is the same lefthook does for [`install`](#lefthook-install) command but you don't need to create a configuration first.

To use custom scripts as hooks create the required directories with `lefthook add pre-commit --dirs`.

**Example**

```bash
$ lefthook add pre-push --dirs
```

Describe pre-push commands in `lefthook.yml`:

```yml
pre-push:
  scripts:
    "audit.sh":
      runner: bash
```

Edit the script:

```bash
$ vim .lefthook/pre-push/audit.sh
...
```

Run `git push` and lefthook will run `bash audit.sh` as a pre-push hook.

### `lefthook run`

`lefthook run` executes the commands and scripts configured for a given hook. Generated hooks call `lefthook run` implicitly.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix

test:
  commands:
    js-test:
      run: yarn test
```

Install the hook.

```bash
$ lefthook install
```

Run `test`.

```bash
$ lefthook run test # will run 'yarn test'
```

Commit changes.

```bash
$ git commit # will run pre-commit hook ('yarn lint --fix')
```

Or run manually also

```bash
$ lefthook run pre-commit
```

You can also specify a flag to run only some commands:

```bash
$ lefthook run pre-commit --commands lint
```

and optionally run either on all files (any `{staged_files}` placeholder acts as `{all_files}`) or a list of files:

```bash
$ lefthook run pre-commit --all-files
$ lefthook run pre-commit --file file1.js --file file2.js
```

(if both are specified, `--all-files` is ignored)

### `lefthook version`

`lefthook version` prints the current binary version. Print the commit hash with `lefthook version --full`

**Example**

```bash
$ lefthook version --full

1.1.3 bb099d13c24114d2859815d9d23671a32932ffe2
```

### `lefthook self-update`

`lefthook self-update` updates the binary with the latest lefthook release on Github. This command is available only if you install lefthook from sources or download the binary from the Github Releases. For other ways use package-specific commands to update lefthook.

## ENV variables

### `LEFTHOOK`

Use `LEFTHOOK=0 git ...` or `LEFTHOOK=false git ...` to disable lefthook when running git commands.

**Example**

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

### `LEFTHOOK_EXCLUDE`

Use `LEFTHOOK_EXCLUDE={list of tags or command names to be excluded}` to skip some commands or scripts by tag or name (for commands only). See the [`exclude_tags`](./configuration.md#exclude_tags) configuration option for more details.

**Example**

```bash
LEFTHOOK_EXCLUDE=ruby,security,lint git commit -am "Skip some tag checks"
```

### `LEFTHOOK_OUTPUT`

Use `LEFTHOOK_OUTPUT={list of output values}` to specify what to print in your output. You can also set `LEFTHOOK_OUTPUT=false` to disable all output except for errors. Refer to the [`output`](./configuration.md#output) configuration option for more details.

**Example**

```bash
$ LEFTHOOK_OUTPUT=summary lefthook run pre-commit
summary: (done in 0.52 seconds)
✔️  lint
```

### `LEFTHOOK_QUIET`

You can skip some outputs printed by lefthook by setting the `LEFTHOOK_QUIET` environment variable. Provide a list of output types to be skipped. See the [`skip_output`](./configuration.md#skip_output) configuration option for more details.

**Example**

```bash
$ LEFTHOOK_QUIET=meta,execution lefthook run pre-commit

  EXECUTE > lint

SUMMARY: (done in 0.01 seconds)
🥊  lint
```

### `LEFTHOOK_VERBOSE`

Set `LEFTHOOK_VERBOSE=1` or `LEFTHOOK_VERBOSE=true` to enable verbose printing.

### `LEFTHOOK_BIN`

Set `LEFTHOOK_BIN` to a location where lefthook is installed to use that instead of trying to detect from the it the PATH or from a package manager.

Useful for cases when:

- lefthook is installed multiple ways, and you want to be explicit about which one is used (example: installed through homebrew, but also is in Gemfile but you are using a ruby version manager like rbenv that prepends it to the path)
- debugging and/or developing lefthook

### `NO_COLOR`

Set `NO_COLOR=true` to disable colored output in lefthook and all subcommands that lefthook calls.

### `CLICOLOR_FORCE`

Set `CLICOLOR_FORCE=true` to force colored output in lefthook and all subcommands.

## Tips

### Local config

Use `lefthook-local.yml` to overwrite or extend options from the main config. (Don't forget to add this file to `.gitignore`)

### Disable lefthook in CI

When using NPM package `lefthook` set `CI=true` in your CI (if it does not set automatically). When `CI=true` is set lefthook NPM package won't install the hooks in the postinstall script.

### Commitlint example

Let's create a bash script to check conventional commit status `.lefthook/commit-msg/commitlint.sh`:

```bash
echo $(head -n1 $1) | npx commitlint --color
```

Now we can ask lefthook to run our bash script by adding this code to
`lefthook.yml` file:

```yml
# lefthook.yml

commit-msg:
  scripts:
    "commitlint.sh":
      runner: bash
```

When you try to commit `git commit -m "haha bad commit text"` script `commitlint.sh` will be executed. Since commit text doesn't match the default config or custom config that you setup for `commitlint`, the process will be interrupted.

### Parallel execution

You can enable parallel execution if you want to speed up your checks.
Lets imagine we have the following rules to lint the whole project:

```
bundle exec rubocop --parallel && \
bundle exec danger && \
yarn eslint --ext .es6 app/assets/javascripts && \
yarn eslint --ext .es6 test/javascripts && \
yarn eslint --ext .es6 plugins/**/assets/javascripts && \
yarn eslint --ext .es6 plugins/**/test/javascripts && \
yarn eslint app/assets/javascripts test/javascripts
```

Rewrite it in lefthook custom group. We call it `lint`:

```yml
# lefthook.yml

lint:
  parallel: true
  commands:
    rubocop:
      run: bundle exec rubocop --parallel
    danger:
      run: bundle exec danger
    eslint-assets:
      run: yarn eslint --ext .es6 app/assets/javascripts
    eslint-test:
      run: yarn eslint --ext .es6 test/javascripts
    eslint-plugins-assets:
      run: yarn eslint --ext .es6 plugins/**/assets/javascripts
    eslint-plugins-test:
      run: yarn eslint --ext .es6 plugins/**/test/javascripts
    eslint-assets-tests:
      run: yarn eslint app/assets/javascripts test/javascripts
```

Then call this group directly:

```
lefthook run lint
```

### Concurrent files overrides

To prevent concurrent problems with read/write files try `flock`
utility.

```yml
# lefthook.yml

graphql-schema:
  glob: "{Gemfile.lock,app/graphql/**/*}"
  run: flock webpack/application/typings/graphql-schema.json yarn typings:update && git diff --exit-code --stat HEAD webpack/application/typings
frontend-tests:
  glob: "**/*.js"
  run: flock -s webpack/application/typings/graphql-schema.json yarn test --findRelatedTests {files}
frontend-typings:
  glob: "**/*.js"
  run: flock -s webpack/application/typings/graphql-schema.json yarn run flow focus-check {files}
```

### Capture ARGS from git in the script

Example script for `prepare-commit-msg` hook:

```bash
COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

### Git LFS support

> :warning: If git-lfs binary is not installed and not required in your project, LFS hooks won't be executed, and you won't be warned about it.

Lefthook runs LFS hooks internally for the following hooks:

- post-checkout
- post-commit
- post-merge
- pre-push

Errors are suppressed if git LFS is not required for the project. You can use [`LEFTHOOK_VERBOSE`](#lefthook_verbose) ENV to make lefthook show git LFS output.


### Pass stdin to a command or script

When you need to read the data from stdin – specify [`use_stdin: true`](./configuration.md#use_stdin). This option is good when you write a command or script that receives data from git using stdin (for the `pre-push` hook, for example).

### Using an interactive command or script

When you need to interact with user – specify [`interactive: true`](./configuration.md#interactive). Lefthook will connect to the current TTY and forward it to your command's or script's stdin.
