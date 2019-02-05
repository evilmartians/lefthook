# Hookah

Hookah it`s a simple manager of git hooks.

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

### I clone the existed repo whick uses hookah. How can I setup hooks?

We suppose repo already have the hookah structure. So all of you need it run install:

```bash
hookah install
```
Hookah wiil read existed hook groups and reproduce hooks in `.git/hooks` directory.
