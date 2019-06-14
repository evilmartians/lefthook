![Build Status](https://api.travis-ci.org/Arkweid/lefthook.svg?branch=master)

# Lefthook

Fast and powerfull Git hook manager for Node.js, Ruby or any other type of projects. For example, it can run tests on changed files during `git commit`. [TODO ещё пару примеров]

* **Fast.** It is written on Go. Can run commands in parallel.
* **Powerful.** With a few lines in config you can check only changed files on `prepush` hook.
* **Simple.** It is single dependency-free binary, which can work in any environment.

```yml
# On `git push` lefthook will run spelling and links check for all changed files
pre-push:
  parallel: true
  commands:
    spelling:
      files: git diff --name-only @{push} || git diff --name-only master
      glob: "*.md"
      run: npx yaspeller {files}
    check-links:
      files: git diff --name-only @{push} || git diff --name-only master
      glob: "*.md"
      run: npx markdown-link-check {files}
```

<a href="https://evilmartians.com/?utm_source=lefthook">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

## Usage

Choose your environment:

* **[Node.js](./docs/node.md)**
* **[Ruby](./docs/node.md)**
* [Other environments](./docs/other.md)

Then you can find all Lefhook features in [config format docs](./docs/config.md).
