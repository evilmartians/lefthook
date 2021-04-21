![Build Status](https://api.travis-ci.org/evilmartians/lefthook.svg?branch=master)

# Lefthook

> The fastest polyglot Git hooks manager out there

<img align="right" width="147" height="100" title="Lefthook logo"
     src="./logo_sign.svg">

Fast and powerful Git hooks manager for Node.js, Ruby or any other type of projects.

* **Fast.** It is written in Go. Can run commands in parallel.
* **Powerful.** With a few lines in the config you can check only the changed files on `pre-push` hook.
* **Simple.** It is single dependency-free binary which can work in any environment.

📖 [Read the introduction post](https://evilmartians.com/chronicles/lefthook-knock-your-teams-code-back-into-shape?utm_source=lefthook)

```yml
# On `git push` lefthook will run spelling and links check for all of the changed files
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

Then you can find all Lefthook features in [the full guide](./docs/full_guide.md) and explore [wiki](https://github.com/evilmartians/lefthook/wiki).

***

## Why Lefthook

* ### **Parallel execution**
If you want more speed. [Example](./docs/full_guide.md#parallel-execution)

```yml
pre-push:
  parallel: true
```

* ### **Flexible list of files**
If you want your own list. [Custom](./docs/full_guide.md#custom-file-list) and [prebuilt](./docs/full_guide.md#select-specific-file-groups) examples.

```yml
pre-commit:
  commands:
    frontend-linter:
      run: yarn eslint {staged_files}
    backend-linter:
      run: bundle exec rubocop --force-exclusion {all_files}
    frontend-style:
      files: git diff --name-only HEAD @{push}
      run: yarn stylelint {files}
```

* ### **Glob and regexp filters**
If you want to filter list of files.

```yml
pre-commit:
  commands:
    backend-linter:
      glob: "*.rb" # glob filter
      exclude: "application.rb|routes.rb" # regexp filter
      run: bundle exec rubocop --force-exclusion {all_files}
```

* ### **Execute in sub-directory**
If you want to execute the commands in a relative path

```yml
pre-commit:
  commands:
    backend-linter:
      root: "api/" # Careful to have only trailing slash
      glob: "*.rb" # glob filter
      run: bundle exec rubocop {all_files}
```

* ### **Run scripts**

If oneline commands are not enough, you can execute files. [Example](./docs/full_guide.md#bash-script-example).

```yml
commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

* ### **Tags**
If you want to control a group of commands. [Example](./docs/full_guide.md#skipping-commands-by-tags).

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

* ### **Support Docker**

If you are in the Docker environment. [Example](./docs/full_guide.md#referencing-commands-from-lefthookyml).

```yml
pre-commit:
  scripts:
    "good_job.js":
      runner: docker run -it --rm <container_id_or_name> {cmd}
```

* ### **Local config**

If you a frontend/backend developer and want to skip unnecessary commands or override something into Docker. [Description](./docs/full_guide.md#local-config).

```yml
# lefthook-local.yml
pre-push:
  exclude_tags:
    - frontend
  commands:
    packages-audit:
      skip: true
```

* ### **Direct control**

If you want to run hooks group directly.

```bash
$ lefthook run pre-commit
```

* ### **Your own tasks**

If you want to run specific group of commands directly.

```yml
fixer:
  commands:
    ruby-fixer:
      run: bundle exec rubocop --force-exclusion --safe-auto-correct {staged_files}
    js-fixer:
      run: yarn eslint --fix {staged_files}
```
```bash
$ lefthook run fixer
```


---

## Table of contents:

### Guides
* [Node.js](./docs/node.md)
* [Ruby](./docs/ruby.md)
* [Other environments](./docs/other.md)
* [Full features guide](./docs/full_guide.md)
* [Troubleshooting](https://github.com/evilmartians/lefthook/wiki/Troubleshooting)

### Migrate from
* [Husky](https://github.com/evilmartians/lefthook/wiki/Migration-from-husky)
* [Husky and lint-staged](https://github.com/evilmartians/lefthook/wiki/Migration-from-husky-with-lint-staged)
* [Overcommit](https://github.com/evilmartians/lefthook/wiki/Migration-from-overcommit)

### Examples
* [Simple script](https://github.com/evilmartians/lefthook/tree/master/examples/scripts)
* [Full features](https://github.com/evilmartians/lefthook/tree/master/examples/complete)

### Benchmarks
* [vs Overcommit](https://github.com/evilmartians/lefthook/wiki/Benchmark-lefthook-vs-overcommit)
* [vs pre-commit](https://github.com/evilmartians/lefthook/wiki/Benchmark-lefthook-vs-pre-commit)

### Comparison list
* [vs Overcommit, Husky, pre-commit](https://github.com/evilmartians/lefthook/wiki/Comparison-with-other-solutions)

### Articles
* [Lefthook: Knock your team’s code back into shape](https://evilmartians.com/chronicles/lefthook-knock-your-teams-code-back-into-shape?utm_source=lefthook)
* [Lefthook + Crystalball](https://evilmartians.com/chronicles/lefthook-crystalball-and-git-magic?utm_source=lefthook)
* [Keeping OSS documentation in check with docsify, Lefthook, and friends](https://evilmartians.com/chronicles/keeping-oss-documentation-in-check-with-docsify-lefthook-and-friends?utm_source=lefthook)

