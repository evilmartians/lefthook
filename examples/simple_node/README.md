### Steps:
1. Install lefthook
```bash
npm install @arkweid/lefthook --save-dev
```

2. Edit `lefthook.yml`

```yml
pre-commit:
  commands:
    js-linter:
      glob: "*.{js,jsx}"
      run: npx eslint {staged_files}
```

3. (optional) Test it
```bash
npx lefthook run pre-commit
```
