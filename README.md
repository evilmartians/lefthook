![Build Status](https://api.travis-ci.org/Arkweid/hookah.svg?branch=master)

# Hookah

Hookah it`s a simple manager of git hooks.

<a href="https://evilmartians.com/?utm_source=hookah">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

[![asciicast](https://asciinema.org/a/8KSu1ube3jFOYXeYDSBfIuY8m.svg)](https://asciinema.org/a/8KSu1ube3jFOYXeYDSBfIuY8m)

## Installation

Add Hookah to your system or build it from sources.

```go
go get github.com/Arkweid/hookah
```

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

Done! Pretty simple, huh?

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
    runner: "go"
    runner_args: "run"
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
