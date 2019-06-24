# Lefthook in Node.js

This is guide of using Lefthook git hook manager in Node.js projects. You can find guides for other environments in [README.md](../README.md).

## Install

```bash
npm install @arkweid/lefthook --save-dev
```

## Edit

Edit `lefthook.yml`:

```yml
pre-commit:
  parallel: true
  commands:
    linter:
      files: git diff --name-only @{push}
      glob: "*.{js,ts}"
      run: npx eslint {files}
    tests:
      files: git diff --name-only @{push}
      glob: "*.{js,ts}"
      run: jest --findRelatedTests {files}
```

## Test it
```bash
npx lefthook install && npx lefthook run pre-commit
```
