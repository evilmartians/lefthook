# Usage

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
