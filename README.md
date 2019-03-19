![Build Status](https://api.travis-ci.org/Arkweid/hookah.svg?branch=master)

# Hookah

A single dependency-free binary to manage all your git hooks that works with any language in any environment, and in all common team workflows.

<a href="https://evilmartians.com/?utm_source=hookah">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

## Installation

Add Hookah to your system or build it from sources.

### go

```bash
go get github.com/Arkweid/hookah
```

### npm

```bash
npm i @arkweid/hookah --save-dev
# or yarn:
yarn add -D @arkweid/hookah
```

NOTE: if you install it this way you should call it with `npx` or `yarn` for all listed examples below. (for example: `hookah install` -> `npx hookah install`)

### Homebrew for macOS

```bash
brew install Arkweid/hookah/hookah
```

### snap for Ubuntu

```bash
sudo snap install --devmode hookah
```

Or take it from [binaries](https://github.com/Arkweid/hookah/releases) and install manualy

## Scenarios

### First time user

Go to your project directory and run following commands:

Initialize hookah with the following command

```bash
hookah install
```

It creates `hookah.yml` in the project root directory

Register your hook (You can choose any hook from [this list](https://git-scm.com/docs/githooks))
In our example it `pre-push` githook:

```bash
hookah add pre-push
```

Describe pre-push commands in `hookah.yml`:

```yml
pre-push: # githook name
  commands: # list of commands
    packages-audit: # command name
      run: yarn audit # command for execution
```

That's all! Now on `git push` the `yarn audit` command will be run.
If it fails the `git push` will be interrupted.

### If you already have a hookah config file
Just initialize hookah to make it work :)
```bash
hookah install
```

## More options

## Use glob patterns to choose what files you want to check
```yml
# hookah.yml

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
# hookah.yml

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

Hookah can be even more specific in selecting files.
If you want to choose diff of all changed files between current branch and master branch you can do it this way:

```yml
# hookah.yml

pre-push:
  commands:
    frontend-style:
      files: git diff --name-only master # custom list of files
      glob: "*.{js}"
      run: yarn stylelint {files}
```

`{files}` - shorthand for custom list of files

## Managing bash scripts
If you run `hookah add` command with `-d` flag, hookah will create two directories where you can put bash scripts 
and reference them from `hookah.yml` file

Example:
Let's create `commit-msg` hook with `-d` flag
```bash
hookah add -d commit-msg
```

This command will create `.hookah/commit-msg` and `.hookah-local/commit-msg` dirs.

First one is for common project level scripts.
Second one is for personal scripts. It would be a good idea to add dir`.hookah-local` to `.gitignore`.

Let's create a bash script to check commit templates `.hookah/commit-msg/template_checker`:

```bash
INPUT_FILE=$1
START_LINE=`head -n1 $INPUT_FILE`
PATTERN="^(TICKET)-[[:digit:]]+: "
if ! [[ "$START_LINE" =~ $PATTERN ]]; then
  echo "Bad commit message, see example: TICKET-123: some text"
  exit 1
fi
```

Now we can ask hookah to run our bash script by adding this code to
 `hookah.yml` file:

```yml
# hookah.yml

commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

When you try to commit `git commit -m "haha bad commit text"` script `template_checker` will be executed. Since commit text doesn't match described pattern the commit process will be interrupted.

## Local config
We can use `hookah-local.yml` as local config. Options in this file will overwrite options in `hookah.yml`. (Don't forget to add this file to `.gitignore`)

## Skipping command

Also you can skip commands by `skip` option:

```yml
# hookah-local.yml

pre-push:
  commands:
    packages-audit:
      skip: true
```


## Skipping command by tags

If we have a lot of commands and scripts we can tag them and run skip commands with a specific tag.

For example if we have `hookah.yml` like this:

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
# hookah-local.yml

pre-push:
  exlude_tags:
    - frontend
```

## Referencing commands from hookah.yml

If you have the following config

```yml
# hookah.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: node
```

You can wrap it in docker runner locally:

```yml
# hookah-local.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: docker exec -it --rm <container_id_or_name> {cmd}
```

`{cmd}` - shorthand for command from `hookah.yml`

## Run githook group directly

```bash
hookah run pre-commit
```

## Parallel execution

You can eneable parallel execution if you want to speed up your checks.
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

Rewrite it in hookah custom group. We call it `lint`:

```yml
# hookah.yml

lint:
  commands:
    parallel: true

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
hookah run lint
```

## Complete example

```yml
# hookah.yml

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
# hookah-local.yml

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

## Skip hoookah execution

We can set env variable `HOOKAH` to zero for that

```bash
HOOKAH=0 git commit -am "Hookah skipped"
```

## Skip some tags on the fly

Use HOOKAH_EXCLUDE={list of tags to be excluded} for that

```bash
HOOKAH_EXCLUDE=ruby,security git commit -am "Skip some tag checks"
```

## Capture ARGS from git in script

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
# hookah.yml

source_dir: ".hookah"
source_dir_local: ".hookah-local"
```

## Version

```bash
hookah version
```

## Uninstall

```bash
hookah uninstall
```
