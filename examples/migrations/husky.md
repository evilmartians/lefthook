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
      "pre-commit": "npm test",
      "pre-push": "npm test"
    }
  }
}

```

```yml
# lefthook.yml
pre-commit:
  commands:
    sometest:
      run: npm test

pre-push:
  commands:
    anothertest:
      run: npm test
```

4. (optional) Unhappy? Want to revert?
```bash
npx lefthook uninstall && npm uninstall @arkweid/lefthook && npm install husky --save-dev
```
