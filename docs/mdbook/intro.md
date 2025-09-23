# Introduction

<img align="right" width="147" height="125" title="Lefthook logo"
     src="./favicon.svg">

**Lefthook** is a Git hooks manager. Here is how to

- **[Install](./installation)** lefthook to your project or globally.

- **[Configure](./configuration)** `lefthook.yml` with detailed options explanation.

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
