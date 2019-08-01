## Installation

Add Lefthook to your system or build it from sources.

### go

```bash
go get github.com/Arkweid/lefthook
```

### npm

```bash
npm i @arkweid/lefthook --save-dev
# or yarn:
yarn add -D @arkweid/lefthook
```

NOTE: if you install it this way you should call it with `npx` or `yarn` for all listed examples below. (for example: `lefthook install` -> `npx lefthook install`)

### Rubygems

```bash
gem install lefthook
```

### Homebrew for macOS

```bash
brew install Arkweid/lefthook/lefthook
```

### AUR for Arch

You can install lefthook [package](https://aur.archlinux.org/packages/lefthook) from AUR

Or take it from [binaries](https://github.com/Arkweid/lefthook/releases) and install manually

## Scenarios

### Examples

We have a directory with few examples. You can check it [here](https://github.com/Arkweid/lefthook/tree/master/examples).

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

## More options

## Use glob patterns to choose what files you want to check

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.{js,ts}"
      run: yarn eslint
```

## Select specific file groups

In some cases you want to run checks only against some specific file group.
For example: you may want to run eslint for staged files only.

There are two shorthands for such situations:
`{staged_files}` - staged git files which you try to commit

`{all_files}` - all tracked files by git

```yml
# lefthook.yml

pre-commit:
  commands:
    frontend-linter:
      glob: "*.{js,ts}" # glob filter for list of files
      run: yarn eslint {staged_files} # {staged_files} - list of files
    backend-linter:
      glob: "*.{rb}" # glob filter for list of files
      exclude: "application.rb|routes.rb" # regexp filter for list of files
      run: bundle exec rubocop {all_files} # {all_files} - list of files
```

## Custom file list

Lefthook can be even more specific in selecting files.
If you want to choose diff of all changed files between the current branch and master branch you can do it this way:

```yml
# lefthook.yml

pre-push:
  commands:
    frontend-style:
      files: git diff --name-only master # custom list of files
      glob: "*.{js}"
      run: yarn stylelint {files}
```

`{files}` - shorthand for a custom list of files

## Managing scripts

If you run `lefthook add` command with `-d` flag, lefthook will create two directories where you can put scripts and reference them from `lefthook.yml` file.

Example:
Let's create `commit-msg` hook with `-d` flag

```bash
lefthook add -d commit-msg
```

This command will create `.lefthook/commit-msg` and `.lefthook-local/commit-msg` dirs.

The first one is for common project level scripts.
The second one is for personal scripts. It would be a good idea to add dir`.lefthook-local` to `.gitignore`.

Create scripts `.lefthook/commit-msg/hello.js` and `.lefthook/commit-msg/hi.rb`

```yml
# lefthook.yml

commit-msg:
  scripts:
    "hello.js":
      runner: node
    "hi.rb":
      runner: ruby
```

### Bash script example

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

When you try to commit `git commit -m "haha bad commit text"` script `template_checker` will be executed. Since commit text doesn't match the described pattern the commit process will be interrupted.

## Local config

We can use `lefthook-local.yml` as local config. Options in this file will overwrite options in `lefthook.yml`. (Don't forget to add this file to `.gitignore`)

## Skipping commands

You can skip commands by `skip` option:

```yml
# lefthook-local.yml

pre-push:
  commands:
    packages-audit:
      skip: true
```

## Skipping commands by tags

If we have a lot of commands and scripts we can tag them and run skip commands with a specific tag.

For example, if we have `lefthook.yml` like this:

```yml
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

## Piped option
If any command in the sequence fails, the other will not be executed.
```yml
database:
  piped: true
  commands:
    1_create:
      run: rake db:create
    2_migrate:
      run: rake db:migrate
    3_seed:
      run: rake db:seed
```

## Referencing commands from lefthook.yml

If you have the following config

```yml
# lefthook.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: node
```

You can wrap it in docker runner locally:

```yml
# lefthook-local.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: docker exec -it --rm <container_id_or_name> {cmd}
```

`{cmd}` - shorthand for the command from `lefthook.yml`

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

## Complete example

```yml
# lefthook.yml

pre-commit:
  commands:
    eslint:
      glob: "*.{js,ts}"
      run: yarn eslint {staged_files}
    rubocop:
      tags: backend style
      glob: "*.{rb}"
      exclude: "application.rb|routes.rb"
      run: bundle exec rubocop {all_files}
    govet:
      tags: backend style
      files: git ls-files -m
      glob: "*.{go}"
      run: go vet {files}

  scripts:
    "hello.js":
      runner: node
    "any.go":
      runner: go run

  parallel: true
```

```yml
# lefthook-local.yml

pre-commit:
  exclude_tags:
    - backend

  scripts:
    "hello.js":
      runner: docker exec -it --rm <container_id_or_name> {cmd}
  commands:
    govet:
      skip: true
```

## Skip lefthook execution

We can set env variable `LEFTHOOK` to zero for that

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

## Skip some tags on the fly

Use LEFTHOOK_EXCLUDE={list of tags to be excluded} for that

```bash
LEFTHOOK_EXCLUDE=ruby,security git commit -am "Skip some tag checks"
```

## Capture ARGS from git in the script

Example script for `prepare-commit-msg` hook:

```bash
COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

## Change directory for script files

You can do this through this config keys:

```yml
# lefthook.yml

source_dir: ".lefthook"
source_dir_local: ".lefthook-local"
```


## Disable colors

By agrs:
```bash
lefthook --no-colors run pre-commit
```
By config `lefthook.yml`, just add the option:
```yml
colors: false
```

## Version

```bash
lefthook version
```

## Uninstall

```bash
lefthook uninstall
```

## More info
Have a qustions? Check the [wiki](https://github.com/Arkweid/lefthook/wiki).
