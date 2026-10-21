---
type: concept
title: PHP
source: "https://lefthook.dev/installation/php/"
path: /installation/php/
updated: 2026-10-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-10-21T12:04:44.346Z"
---
# PHP

1. Add the package as a dependency

```sh
composer require --dev evilmartians/lefthook
```

2. Check that it works

```sh
vendor/bin/lefthook version
```

3. Install the Git hooks

```sh
vendor/bin/lefthook install
```

After that lefthook should automatically find its path via `composer exec` on Git hooks.
