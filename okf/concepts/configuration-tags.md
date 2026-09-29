---
type: concept
title: tags
source: "https://lefthook.dev/configuration/tags/"
path: /configuration/tags/
updated: 2026-09-29
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-09-29T15:50:18.527Z"
---
---
title: "tags"
---

# `tags`

You can specify tags for commands and scripts. This is useful for [excluding](./exclude_tags.md). You can specify more than one tag using comma.

#### Example

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      tags:
        - frontend
        - js
      run: yarn lint
    test:
      tags:
        - backend
        - ruby
      run: bundle exec rspec
```
