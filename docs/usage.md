# Usage

You want to use lefthook in your git project. Here is what you need:

1. Create a `lefthook.yml` (or use any other [supported name](./configuration.md#config-file))
1. [Install](#lefthook-install) lefthook git hooks

Then use git as usually, you don't need to reinstall lefthook when you change the config.

- [Commands](#commands)
  - [`lefthook install`](#lefthook-install)
  - [`lefthook uninstall`](#lefthook-uninstall)
  - [`lefthook add`](#lefthook-add)
  - [`lefthook run`](#lefthook-run)
  - [`lefthook version`](#lefthook-version)
- [Control behavior with ENV variables](#control-behavior-with-env-variables)
  - [`LEFTHOOK`](#lefthook)
  - [`LEFTHOOK_EXCLUDE`](#lefthook_exclude)
  - [`LEFTHOOK_QUIET`](#lefthook_quiet)
  - [`LEFTHOOK_VERBOSE`](#lefthook_verbose)
- [Features and tips](#features-and-tips)
  - [Disable lefthook in CI](#disable-lefthook-in-ci)
  - [Local config](#local-config)
  - [Commitlint example](#commitlint-example)
  - [Parallel execution](#parallel-execution)
  - [Concurrent files overrides](#concurrent-files-overrides)
  - [Capture ARGS from git in the script](#capture-args-from-git-in-the-script)
  - [Git LFS support](#git-lfs-support)
  - [Pass stdin to a command or script](#pass-stdin-to-a-command-or-script)
  - [Using an interactive command or script](#using-an-interactive-command-or-script)

----

## Commands

Lefthook is a CLI utility and it has several commands for convenience. You can check the usage via `lefthook help` or `lefthook <command> -h/--help`.

Here are the description of common usage of these commands.

### `lefthook install`

Run `lefthook install` to initialize a `lefthook.yml` config and/or synchronize `.git/hooks/` with your configuration. This must be the first thing you do after cloning the repo with `lefthook.yml` config. For config options see our [configuration documentation](./configuration.md).

> If you use lefthook with NPM package manager it should have already run `lefthook install` in postinstall scripts.

### `lefthook uninstall`

Run `lefthook uninstall` when you want to clear hooks `.git/hooks/` related to `lefthook.yml` configuration. Use `-f/--force` flag to remove all git hooks.

> Sometimes you feel like your git hooks are a mess and you want to start from the beginning. Use `lefthook uninstall` in this case.

### `lefthook add`

Run `lefthook add pre-commit`, and lefthook will create a hook `.git/hooks/pre-commit`. This is the same lefthook does for [`install`](#lefthook-install) command but you don't need to create a configuration first.

If you want to use scripts you can simplify adding new scripts with `lefthook add --dirs pre-commit`.

**Example**

```bash
$ lefthook add --dirs pre-push
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

That's all! Now on `git push` the `audit.sh` command will be run in `bash` interpreter.
If it fails the `git push` will be interrupted.

### `lefthook run`

This command is implicitly called in every git hook managed by lefthook. You can also use it for manual invoking some hooks handlers. You can also describe your own hooks that can be called manually only.

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
$ lefthook run pre-commit --files file1.js,file2.js
```

(if both are specified, `--all-files` is ignored)

### `lefthook version`

You can check version with `lefthook version` and you can also check the commit hash with `lefthook version --full`

**Example**

```bash
$ lefthook version --full

1.1.3 bb099d13c24114d2859815d9d23671a32932ffe2
```

## Control behavior with ENV variables

### `LEFTHOOK`

You can set ENV variable `LEFTHOOK` to `0` or `false` to disable lefthook.

**Example**

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

### `LEFTHOOK_EXCLUDE`

Use `LEFTHOOK_EXCLUDE=`{list of tags or command names to be excluded} to skip some commands or scripts by tag or name (for commands only). See [`exclude_tags`](./configuration.md#exclude_tags) config option for more details.

**Example**

```bash
LEFTHOOK_EXCLUDE=ruby,security,lint git commit -am "Skip some tag checks"
```

### `LEFTHOOK_QUIET`

You can skip some output printed by lefthook with `LEFTHOOK_QUIET` ENV variable. Just provide a list of output types. See [`skip_output`](./configuration.md#skip_output) config option for more details.

**Example**

```bash
$ LEFTHOOK_QUIET=meta,execution lefthook run pre-commit

  EXECUTE > lint

SUMMARY: (done in 0.01 seconds)
🥊  lint
```

### `LEFTHOOK_VERBOSE`

Set `LEFTHOOK_VERBOSE=1` or `LEFTHOOK_VERBOSE=true` to enable verbose printing.

## Features and tips

### Disable lefthook in CI

Add `CI=true` env variable if it doesn't exists on your service by default. Otherwise, if you use lefthook NPM package it will try running `lefthook install` in postinstall scripts.


### Local config

We can use `lefthook-local.yml` as local config. Options in this file will overwrite options in `lefthook.yml`. (Don't forget to add this file to `.gitignore`)

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
