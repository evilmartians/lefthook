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
