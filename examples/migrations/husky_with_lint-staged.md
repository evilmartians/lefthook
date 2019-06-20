### Migration from husky

## Steps:
1. Uninstall husky
```bash
npm uninstall husky
```

2. Install lefthook
```bash
npm install @arkweid/lefthook --save-dev
```

3. Move hooks from `package.json` to `lefthook.yml`

```json
// package.json
{
  "husky": {
    "hooks": {
      "pre-commit": "lint-staged"
    }
  },
  "lint-staged": {
    "*.{js,jsx}": "eslint"
  }
}

```

```yml
# lefthook.yml
pre-commit:
  commands:
    eslint:
      glob: "*.{js,jsx}"
      run: npx eslint {staged_files}
```

4. (optional) Unhappy? Want to revert?
```bash
npx lefthook uninstall && npm uninstall @arkweid/lefthook && npm install husky --save-dev
```
