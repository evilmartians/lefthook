![Build Status](https://api.travis-ci.org/Arkweid/lefthook.svg?branch=master)

# Lefthook

Lefthook a single dependency-free binary to manage all your git hooks that works with any language in any environment, and in all common team workflows.

<a href="https://evilmartians.com/?utm_source=lefthook">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

## Installation

Add Lefthook to your system or build it from sources.

### go
```bash
go get github.com/Arkweid/lefthook
```

### npm and yarn
```bash
npm i @arkweid/lefthook --save-dev
# or yarn:
yarn add -D @arkweid/lefthook
```
NOTE: if you install it this way you should call it with `npx` or `yarn` for all listed examples below.

### Homebrew for macOS
```bash
brew install Arkweid/lefthook/lefthook
```

### snap for Ubuntu
```bash
sudo snap install --devmode lefthook
```

Or take it from [binaries](https://github.com/Arkweid/lefthook/releases) and install manualy

## Scenarios



### First time user

Go to your project directory and run following commands:

Add configuration file `lefthook.yml`:
```bash
lefthook install
```

Register desired githook. In our example it `pre-push` githook:
```bash
lefthook add pre-push
```
[Here](https://git-scm.com/docs/githooks) you can refresh your knowledge about githooks.

Describe pre-push commands in `lefthook.yml`:
```yml
pre-push:             # githook name
  commands:           # list of commands
    packages-audit:   # command name
      run: yarn audit # command for execution
```
That all! Now on `git push` the `yarn audit` command will run.
If it fail the `git push` will be interrupt.

### Project with existed lefthook
Run:
```bash
lefthook install
```
lefthook wiil read existed hook groups and reproduce hooks in `.git/hooks` directory.

## More options

## Filters for list of files
```yml
# lefthook.yml

pre-commit:
  commands:
    frontend-linter:
      glob: "*.{js,ts}"                    # glob filter for list of files
      run: yarn eslint {staged_files}      # {staged_files} - list of files
    backend-linter:
      glob: "*.{rb}"                       # glob filter for list of files
      exclude: "application.rb|routes.rb"  # regexp filter for list of files
      run: bundle exec rubocop {all_files} # {all_files} - list of files
```

`{staged_files}` - shorthand for staged git files which you try to commit

`{all_files}` - shorthand for all tracked files by git

## Custom list of files
You can describe a custom list of files. Common scenario for pre-push "list of all changed files between current branch and master branch" you can do it this way:
```yml
# lefthook.yml

pre-push:
  commands:
    frontend-style:
      files: git diff --name-only master # custom list of files
      glob: "*.{js}"
      run: yarn stylelint {files}     
```
`{files}` - shorthand for custom list of files

## Scripts
Lefthook also can run script files. Common scenario for `commit-msg` "check commit text template".
Register `commit-msg` githook:
```bash
lefthook add -d commit-msg
```
This command also create two dirs `.lefthook/commit-msg` and `.lefthook-local/commit-msg`.
First dir for common project level scripts. Second  one for your personal scripts. Add dir`.lefthook-local` to `.gitignore`.

Create template_checker script file `.lefthook/commit-msg/template_checker`:
```bash
INPUT_FILE=$1
START_LINE=`head -n1 $INPUT_FILE`
PATTERN="^(TICKET)-[[:digit:]]+: "
if ! [[ "$START_LINE" =~ $PATTERN ]]; then
  echo "Bad commit message, see example: TICKET-123: some text"
  exit 1
fi
```
We need to know which program can execute the code in `template_checker`. Describe it in `lefthook.yml`:
```yml
# lefthook.yml

commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

Now when you try to commit `git commit -m "haha bad commit text"` script `template_checker` will be executed. And because commit text not match with described pattern process will interrupt.

## Tags
If we have a lot of commands and scripts we can divide them by tags and run only relevent for our work commands.
For example we have `lefthook.yml` like this:
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

We provide `lefthook-local.yml` as local config. Options in this file overwrite options in `lefthook.yml`. s Add it to `.gitignore`

You can skip commands by tags:
```yml
# lefthook-local.yml

pre-push:
  exlude_tags:
    - frontend
```
Also you can skip commands by `skip` option:
```yml
# lefthook-local.yml

pre-push:
  commands:
    packages-audit:
      skip: true
```

## Wrapper {cmd}
If some runner installed in docker you can wrap it in docker runner:

```yml
# lefthook.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: node
```

```yml
# lefthook-local.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: docker exec -it --rm <container_id_or_name> {cmd}
```
`{cmd}` - shorthand for command from `lefthook.yml`

## Run githook group directly

```bash
lefthook run pre-commit
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

Rewrite it in lefthook custom group. We call it `lint`:
```yml
# lefthook.yml

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

## Skip hoookah execution

We have env `LEFTHOOK=0` for that

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

## Skip some tags on the fly

We have env HOOKAH_EXCLUDE=tag,tag for that

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
# lefthook.yml

source_dir: ".lefthook"
source_dir_local: ".lefthook-local"
```

## Version

```bash
lefthook version
```

## Uninstall

```bash
lefthook uninstall
```
