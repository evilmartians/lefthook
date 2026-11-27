---
paths:
  - "tests/integration/*"
---

# Integration tests

Integration tests are useful for covering complex logic that unit tests may not catch. When behavior depends on ENV variables, git configuration, project layout and other environment-related things adding an integration test make sense. In other situations a unit test must be enough.

Lefthook uses github.com/rogpeppe/go-internal/testscript lib for integration tests. If a test runs `lefthook install` or uses any `git` command it must have `setup_git` as the first line.

Here are the commands available without leading `exec`:

- git
- lefthook
- setup_git
- rm
- mv
- mkdir
- cp
- cd

Other commands must be prefixed with `exec`, like `exec cat file.txt`

## Important!

Keep integration tests readable and small.

One integration test file must check only one case.

If unit test covers the case - don't add an integration test.
