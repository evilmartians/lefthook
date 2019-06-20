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
pre-commit:
  commands:
    rubocop:
      glob: "*.{rb}"
      run: rubocop {staged_files}
```

4. (optional) Test it
```bash
lefthook run pre-commit
```
