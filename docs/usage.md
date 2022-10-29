# Usage TBD

- [Install](#install)
- [Uninstall](#uninstall)
- [Version](#version)
- [Disable in CI](#disable-lefthook-in-ci)

----

## Install

### First time user

Initialize lefthook with the following command

```bash
lefthook install
```

It creates `lefthook.yml` in the project root directory

Register your hook (You can choose any hook from [this list](https://git-scm.com/docs/githooks))
In our example it `pre-push` githook:

```bash
lefthook add pre-push
```

Describe pre-push commands in `lefthook.yml`:

```yml
pre-push: # githook name
  commands: # list of commands
    packages-audit: # command name
      run: yarn audit # command for execution
```

That's all! Now on `git push` the `yarn audit` command will be run.
If it fails the `git push` will be interrupted.

### If you already have a lefthook config file

Just initialize lefthook to make it work :)

```bash
lefthook install
```

## Uninstall

```bash
lefthook uninstall
```

## Version

```bash
lefthook version
```

## Disable lefthook in CI

Add `CI=true` env variable if it doesn't exists on your service by default. Otherwise, if you use lefthook NPM package it will try running `lefthook install` in postinstall scripts.


## Local config

We can use `lefthook-local.yml` as local config. Options in this file will overwrite options in `lefthook.yml`. (Don't forget to add this file to `.gitignore`)


## Bash script example with Commitlint

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


## Run githook group directly

```bash
lefthook run pre-commit
```


## Parallel execution

You can enable parallel execution if you want to speed up your checks.
Lets get example from [discourse](https://github.com/discourse/discourse/blob/master/.travis.yml#L77-L83) project.

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

## Skip lefthook execution

We can set env variable `LEFTHOOK` to zero for that

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

## Skip some tags on the fly

Use LEFTHOOK_EXCLUDE={list of tags or command names to be excluded} for that

```bash
LEFTHOOK_EXCLUDE=ruby,security,lint git commit -am "Skip some tag checks"
```

## Concurrent files overrides

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

## Capture ARGS from git in the script

Example script for `prepare-commit-msg` hook:

```bash
COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

## Git LFS support

Lefthook runs LFS hooks internally for the following hooks:

- post-checkout
- post-commit
- post-merge
- pre-push
