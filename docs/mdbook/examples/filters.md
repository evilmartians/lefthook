## Filters

Files passed to your hooks can be filtered with the following options

- [`glob`](../configuration/glob.md)
- [`exclude`](../configuration/exclude.md)
- [`file_types`](../configuration/file_types.md)
- [`root`](../configuration/root.md)

In this example all **staged files** will pass through these filters.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint {staged_files} --fix
      glob: "*.{js,ts}"
      root: frontend
      exclude:
        - *.config.js
        - *.config.ts
      file_types:
        - not executable
```

Imagine you've staged the following files

```bash
backend/asset.js
frontend/src/index.ts
frontend/bin/cli.js # <- executable
frontend/eslint.config.js
frontend/README.md
```

After all filters applied the `lint` command will execute the following:

```bash
yarn lint frontend/src/index.ts --fix
```
