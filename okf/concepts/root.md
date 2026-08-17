---
type: concept
title: "What is Lefthook?"
description: "Welcome to Lefthook documentation"
source: "https://lefthook.dev/"
path: /
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.271Z"
---
---
title: "What is Lefthook?"
description: "Welcome to Lefthook documentation"
---

**Lefthook** is a Git hooks manager. It is

- Fast
- Powerful
- Simple

## How does lefthook work?

You

- Create [`lefthook.yml`](./configuration.md) configuration file
- Run `lefthook install`

Lefthook installs the configured hooks into `.git/hooks/`. Hook is a simple script that calls `lefthook run {hook-name}` when executed.

## How to install lefthook?

The most common way is to use the package manager of your project, e.g. [gem](./installation/ruby.md) or [npm package](./installation/node.md).

You can also install lefthook via [Homebrew](./installation/homebrew.md), [`winget`](./installation/winget.md), [`yum`](./installation/rpm.md), [`apt`](./installation/deb.md), [`apk`](./installation/alpine.md), [`scoop`](./installation/scoop.md)

## Example configuration

Run linters on `pre-commit` hook.

```yml
# lefthook.yml

pre-commit:
  parallel: true
  jobs:
    - run: yarn run stylelint --fix '{staged_files}'
      glob: "*.css"
      stage_fixed: true

    - run: yarn run eslint --fix '{staged_files}'
      glob:
        - "*.ts"
        - "*.js"
        - "*.tsx"
        - "*.jsx"
      stage_fixed: true
```

---

**Lefthook** is built by **[Evil Martians](https://evilmartians.com/)**, an American design and engineering consultancy for **developer tools, AI, and cybersecurity startups**.
