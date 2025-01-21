# Introduction

<img align="right" width="147" height="125" title="Lefthook logo"
     src="./favicon.svg">

**Lefthook** is a Git hooks manager. This documentation provides the reference for installing, configuring and using the lefthook.


- **[Installation instructions](./installation)** to install lefthook to your OS or project.

- **[Examples](./examples)** of lefthook common usage.

- **[Configuration](./configuration)** with detailed explanation of lefthook options.


**Example:** Run your linters on `pre-commit` hook and forget about the routine.

```yml
# lefthook.yml

pre-commit:
  parallel: true
  jobs:
    - run: yarn run stylelint --fix {staged_files}
      glob: "*.css"
      stage_fixed: true

    - run: yarn run eslint --fix "{staged_files}"
      glob:
        - "*.ts"
        - "*.js"
        - "*.tsx"
        - "*.jsx"
      stage_fixed: true
```

---

<a href="https://evilmartians.com/?utm_source=lefthook">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="100%" height="54"></a>


❓_If you have a question or found a mistake in the documentation, please create a new [discussion](https://github.com/evilmartians/lefthook/discussions/new/choose). Small contributions help maintaining the quality of the project._

