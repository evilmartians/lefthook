![Build Status](https://api.travis-ci.org/Arkweid/lefthook.svg?branch=master)

# Lefthook

<img align="right" width="100" height="100" title="Lefthook logo"
     src="./logo_sign.svg">

Fast and powerfull Git hook manager for Node.js, Ruby or any other type of projects.

* **Fast.** It is written on Go. Can run commands in parallel.
* **Powerful.** With a few lines in config you can check only changed files on `pre-push` hook.
* **Simple.** It is single dependency-free binary, which can work in any environment.

```yml
# On `git push` lefthook will run spelling and links check for all changed files
pre-push:
  parallel: true
  commands:
    spelling:
      files: git diff --name-only HEAD @{push}
      glob: "*.md"
      run: npx yaspeller {files}
    check-links:
      files: git diff --name-only HEAD @{push}
      glob: "*.md"
      run: npx markdown-link-check {files}
```

<a href="https://evilmartians.com/?utm_source=lefthook">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

## Usage

Choose your environment:

* **[Node.js](./docs/node.md)**
* **[Ruby](./docs/ruby.md)**
* [Other environments](./docs/other.md)

Then you can find all Lefhook features in [full guide](./docs/full_guide.md) and explore [wiki](https://github.com/Arkweid/lefthook/wiki).

---

## Table of content:

### Guides:

* [Node.js](./docs/node.md)
* [Ruby](./docs/ruby.md)
* [Other environments](./docs/other.md)
* [Full features guide](./docs/full_guide.md)

### Migrations from:
* [Husky](https://github.com/Arkweid/lefthook/wiki/Migration-from-husky)
* [Husky and lint-staged](https://github.com/Arkweid/lefthook/wiki/Migration-from-husky-with-lint-staged)
* [Overcommit](https://github.com/Arkweid/lefthook/wiki/Migration-from-overcommit)

### Examples
* [Simple script](https://github.com/Arkweid/lefthook/tree/master/examples/scripts)
* [Full features](https://github.com/Arkweid/lefthook/tree/master/examples/complete)

### Benchmarks
* [vs Overcommit](https://github.com/Arkweid/lefthook/wiki/Benchmark-lefthook-vs-overcommit)

### Comprasion
* [vs Overcommit and Husky](https://github.com/Arkweid/lefthook/wiki/Comprasion-with-other-solutions)
