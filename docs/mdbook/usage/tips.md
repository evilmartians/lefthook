## Tips

> Small tips for better experience with lefthook

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

> If git-lfs binary is not installed and not required in your project, LFS hooks won't be executed, and you won't be warned about it.

Lefthook runs LFS hooks internally for the following hooks:

- post-checkout
- post-commit
- post-merge
- pre-push

Errors are suppressed if git LFS is not required for the project. You can use [`LEFTHOOK_VERBOSE`](./env.md#lefthook_verbose) ENV to make lefthook show git LFS output.

### Pass stdin to a command or script

When you need to read the data from stdin – specify [`use_stdin: true`](../configuration/use_stdin.md). This option is good when you write a command or script that receives data from git using stdin (for the `pre-push` hook, for example).

### Using an interactive command or script

When you need to interact with user – specify [`interactive: true`](../configuration/interactive.md). Lefthook will connect to the current TTY and forward it to your command's or script's stdin.
