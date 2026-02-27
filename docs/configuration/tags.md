---
title: "tags"
---

# `tags`

You can specify tags for commands and scripts. This is useful for [excluding](./exclude_tags.md). You can specify more than one tag using comma.

**Example**

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
