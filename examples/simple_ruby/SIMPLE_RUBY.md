### What we want to achieve:

Run lints/tests against changed files between feature_branch and master when we try to push our changes in repo.

### Steps:
1. Install lefthook

```bash
gem install lefthook
```

2. Initialize lefthook

```bash
lefthook install
```

3. Edit `lefthook.yml`

```yml
pre-push:
  commands:
    audit:
      run: bundle audit
    rubocop:
      files: git diff --name-only HEAD master
      glob: "*.{rb}"
      run: rubocop {files}
```

4. (optional) Test it
```bash
lefthook run pre-push
```
