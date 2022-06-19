# Lefthook in Node.js

This is guide of using Lefthook git hook manager in Node.js projects. You can find guides for other environments in [README.md](../README.md).

## Install

Lefthook is available on NPM in two flavors:

 1. [@evilmartians/lefthook](https://www.npmjs.com/package/@evilmartians/lefthook) with pre-bundled binaries for all architectures:

    ```bash
    npm install @evilmartians/lefthook --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook
    ```

 2. [@evilmartians/lefthook-installer](https://www.npmjs.com/package/@evilmartians/lefthook-installer) that wil fetch binary file on installation:

    ```bash
    npm install @evilmartians/lefthook-installer --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook-installer
    ```

## Edit

Edit `lefthook.yml`:

```yml
pre-commit:
  parallel: true
  commands:
    linter:
      files: git diff --name-only @{push}
      glob: "*.{js,ts,jsx,tsx}"
      run: npx eslint {files}
    tests:
      files: git diff --name-only @{push}
      glob: "*.{js,ts, jsx, tsx}"
      run: jest --findRelatedTests {files}
```

## Test it
```bash
npx @evilmartians/lefthook install && npx @evilmartians/lefthook run pre-commit
```

### More info
Have a question? Check the [wiki](https://github.com/evilmartians/lefthook/wiki).
