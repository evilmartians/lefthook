---
type: concept
title: env
source: "https://lefthook.dev/configuration/env/"
path: /configuration/env/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:12:03.059Z"
---
---
title: "env"
---

# `env`

You can specify some ENV variables for the command or script.

#### Example

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      env:
        RAILS_ENV: test
      run: bundle exec rspec
```

#### Extending `PATH`

If your hook is run by a GUI program and you use PATH tweaks in your `~/.<shell>rc`, you might see an *executable not found* error. You can extend `$PATH` via `lefthook-local.yml`:

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      run: yarn test
```

```yml
# lefthook-local.yml

pre-commit:
  commands:
    test:
      env:
        PATH: $PATH:/home/me/path/to/yarn
```

::: callout tip
Useful when running lefthook across different OSes or shells where environment variables are set differently.
:::
