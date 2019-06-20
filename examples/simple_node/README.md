### What we want to achieve:

Run lints/tests against changed files between feature_branch and master when we try to push our changes in repo.

### Steps:
1. Install lefthook
```bash
npm install @arkweid/lefthook --save-dev
```

2. Edit `lefthook.yml`

```yml
pre-push:
  commands:
    js-linter:
      glob: "*.{js,jsx}"
      run: npx eslint {staged_files}
```

3. (optional) Test it
```bash
npx lefthook run pre-push
```
