![Build Status](https://api.travis-ci.org/Arkweid/hookah.svg?branch=master)

# Hookah

Hookah it`s a simple manager of git hooks.

<a href="https://evilmartians.com/?utm_source=hookah">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

[![asciicast](https://asciinema.org/a/rupBzaCqin2n3qGlNFM9Agm7f.svg)](https://asciinema.org/a/rupBzaCqin2n3qGlNFM9Agm7f)

## Installation

Add Hookah to your system or build it from sources.

### go
```bash
go get github.com/Arkweid/hookah
```

### npm and yarn
```bash
npm i @arkweid/hookah-js --save-dev
# or yarn:
yarn add -D @arkweid/hookah-js

# Now you can call it:
npx hookah -h
```
NOTE: if you install it this way you should call it with `npx` for all listed examples below.

### snap
```bash
sudo snap install --devmode hookah
```

### brew
```bash
brew install Arkweid/hookah/hookah
```

Or take it from [binaries](https://github.com/Arkweid/hookah/releases) and install manualy

## Scenarios

### First time user

Go to your project directory and run:


```bash
hookah install
```

It add for you configuration file `hookah.yml` with default directories for hooks sources.
Now we ready to add hooks! For example we want to add pre commit hooks. Lets do that:


```bash
hookah add pre-commit
```

It will add a hook `.git/hooks/pre-commit`. So every time when you run `git commit` this file will be executed.
That directories also will be created `.hookah` and `.hookah-local`.
Use first one for project/team hooks. Second one for you personal hooks. Add it to `.gitignore`

Next fill the directory `.hookah/pre-commit` with executables you like

```
├───.hookah
│   └───pre-commit
│       ├─── fail_script
│       └─── ok_script
```

Example:
```bash
cat > .hookah/pre-commit/fail_script

#!/bin/sh
exit 1

cat > .hookah/pre-commit/ok_script

#!/bin/sh
exit 0

# Now we can commit:
git commit -am "It fail"
```

Done! Pretty simple, huh?

### Complete example
`hookah.yml`
```yml
source_dir: ".hookah"
source_dir_local: ".hookah-local"

pre-commit:
  # Specify additional parameters for script files
  # If your scripts have shebang notation you can skip
  # this section
  scripts:
    "hello.js":
      runner: node
    "any.go":
      runner: go run

  # Not enough speed? Run all of them in parallel!
  # Default: false
  parallel: true

  commands:
    eslint:
      glob: "*.{js,ts}"
      runner: yarn eslint {staged_files}
    rubocop:
      tags: backend style
      glob: "*.{rb}"
      exclude: "application.rb|routes.rb" # simple regexp for more flexibility
      runner: bundle exec rubocop {all_files}
    govet:
      tags: backend style
      files: git ls-files -m # we can explicity define scope of files
      glob: "*.{go}"
      runner: go vet {files} # {files} will be replaced by matched files as arguments
```
If your team have backend and frontend developers, you can skip unnsecesary hooks this way:
`hookah-local.yml`
```yml
pre-commit:
  # I am fronted developer. Skip all this backend stuff!
  exclude_tags:
    - backend

  scripts:
    "any.go":
      runner: docker exec -it --rm <container_id_or_name> {cmd} # Wrap command from hookah.yml in docker
  commands:
    govet:
      skip: true # You can also skip command with this option
```

### I want to run hook groups directly!

No problem, hookah have command for that:

```bash
hookah run pre-commit

# You will see the summary:
[ FAIL ] fail_script
[ OK ] ok_script
```

### I want to use my own runner! And I dont want to change team/repository scripts.

Ok! For example you have `any.go` script. We can run it in this way:

Add `hookah-local.yml`

Add it to `.gitignore`. It your personal settings.

Next customize the `any.go` script:

```yaml
pre-commit:
  "any.go":
    runner: "go run"
```

Done! Now our script will be executed like this:
```bash
go run any.go
```

### I clone the existed repo which use hookah. How can I setup hooks?

We suppose repo already have the hookah structure. So all of you need it run install:

```bash
hookah install
```
Hookah wiil read existed hook groups and reproduce hooks in `.git/hooks` directory.

### How can I can skip hoookah execution?

We have env HOOKAH=0 for that

```bash
HOOKAH=0 git commit -am "Hookah skipped"
```

### How can I run my linter against only modified files?

No problem. Lets take `rubocop` linter for ruby as example:

```bash
#!/bin/sh

git ls-files -m | xargs rubocop
```

### I dont like bash. Give me working example for golang

Ok-ok! This is how `any.go` may looks like:

```go
package main

import (
  "fmt"
  "os"
  "os/exec"
  "strings"
  
  "github.com/Arkweid/hookah/context"
)

func main() {
  files, _ := context.StagedFiles()
  files = context.FilterByExt(files, ".rb")

  cmd := exec.Command("rubocop", strings.Join(files, " "))

  outputBytes, err := cmd.CombinedOutput()

  fmt.Println(string(outputBytes))

  if err != nil {
    os.Exit(1)
  }
}
```
We include context package only for convenience. It`s just few useful functions.

### Some hooks proved ARGS from git. How can I capture it in my script?
For pure script you can do it like that:

Example for `prepare-commit-msg` hook:
```bash
COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

### Uninstall

```bash
hookah uninstall
```