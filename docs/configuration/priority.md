---
title: "priority"
---

# `priority`

**Default: `0`**

> **Note:** This option makes sense only when `parallel: false` or `piped: true` is set.
>
> Value `0` is considered an `+Infinity`, so commands or scripts with `priority: 0` or without this setting will be run at the very end.

Set priority from 1 to +Infinity. This option can be used to configure the order of the sequential steps.

**Example**

```yml
# lefthook.yml

post-checkout:
  piped: true
  commands:
    db-create:
      priority: 1
      run: rails db:create
    db-migrate:
      priority: 2
      run: rails db:migrate
    db-seed:
      priority: 3
      run: rails db:seed

  scripts:
    "check-spelling.sh":
      runner: bash
      priority: 1
    "check-grammar.rb":
      runner: ruby
      priority: 2
```
