---
type: concept
title: glob_matcher
source: "https://lefthook.dev/configuration/glob_matcher/"
path: /configuration/glob_matcher/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T07:57:38.007Z"
---
---
title: "glob_matcher"
---

# `glob_matcher`

Configure which glob matching engine lefthook uses to filter files.


**Values:**
- `gobwas` (default): see https://github.com/gobwas/glob
- `doublestar`: Usual glob behavior (like in Bash)

#### Example

```yml
# lefthook.yml

glob_matcher: doublestar

pre-commit:
  jobs:
    - name: lint
      run: yarn eslint {staged_files}
      glob: "**/*.{js,ts}"
```

#### Behaviour comparison

```yml
# gobwas (default): **/*.js matches src/app.js but NOT app.js
# doublestar:       **/*.js matches app.js, src/app.js, a/b/c/app.js
```

Use `doublestar` when migrating from other tools or when you need `**` to match files at any depth including the root. The setting applies globally to all `glob` and `exclude` patterns and is backwards compatible.
