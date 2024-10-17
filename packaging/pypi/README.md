![Build Status](https://github.com/evilmartians/lefthook/actions/workflows/test.yml/badge.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/evilmartians/lefthook/badge.svg?branch=master)](https://coveralls.io/github/evilmartians/lefthook?branch=master)

# Lefthook

> The fastest polyglot Git hooks manager out there

<img align="right" width="147" height="100" title="Lefthook logo"
     src="https://raw.githubusercontent.com/evilmartians/lefthook/refs/heads/master/logo_sign.svg">

A Git hooks manager for Node.js, Ruby and many other types of projects.

* **Fast.** It is written in Go. Can run commands in parallel.
* **Powerful.** It allows to control execution and files you pass to your commands.
* **Simple.** It is single dependency-free binary which can work in any environment.

📖 [Read the introduction post](https://evilmartians.com/chronicles/lefthook-knock-your-teams-code-back-into-shape?utm_source=lefthook)

<a href="https://evilmartians.com/?utm_source=lefthook">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

## Install

```bash
pip install lefthook
```

## Usage

Configure your hooks, install them once and forget about it: rely on the magic underneath.

#### TL;DR

```bash
# Configure your hooks
vim lefthook.yml

# Install them to the git project
lefthook install

# Enjoy your work with git
git add -A && git commit -m '...'
```

#### More details

- [**Configuration**](https://github.com/evilmartians/lefthook/blob/master/docs/configuration.md) for `lefthook.yml` config options.
- [**Usage**](https://github.com/evilmartians/lefthook/blob/master/docs/usage.md) for **lefthook** CLI options, supported ENVs, and usage tips.
- [**Discussions**](https://github.com/evilmartians/lefthook/discussions) for questions, ideas, suggestions.
<!-- - [**Wiki**](https://github.com/evilmartians/lefthook/wiki) for guides, examples, and benchmarks. -->
