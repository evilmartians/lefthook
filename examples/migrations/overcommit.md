### Migration from overcommit

## Steps:
1. Uninstall overcommit
```bash
overcommit --uninstall
```

2. Install lefthook
```bash
gem install lefthook && lefthook install
```

3. Move hooks from `.overcommit.yml` to `lefthook.yml`
```yml
# .overcommit.yml
PreCommit:
  RuboCop:
    enabled: true
    command: ['rubocop']
  Audit:
    enabled: true
    command: ['bundle', 'audit']
```

```yml
# lefthook.yml
pre-commit:
  commands:
    rubocop:
      run: rubocop
    audit:
      run: bundle audit
```

4. Sync new hooks and test it!
```bash
lefthook install && lefthook run pre-commit
```

5. (optional) Unhappy? Want to revert?
```bash
lefthook uninstall && gem uninstall lefthook && overcommit --install
```
