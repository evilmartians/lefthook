## `glob_matcher`

You can configure which glob matching engine lefthook uses to filter files. By default, lefthook uses `gobwas/glob`, but you can opt-in to use `doublestar` for standard glob behavior.

**Values:**
- `gobwas` (default): The current glob implementation
- `doublestar`: Standard glob behavior where `**` matches 0 or more directories

**Example:**

```yml
# lefthook.yml

glob_matcher: doublestar

pre-commit:
  jobs:
    - name: lint
      run: yarn eslint {staged_files}
      glob: "**/*.{js,ts}"
```

### Key Differences

The main difference between the two matchers is how they handle `**`:

#### Default behavior (`gobwas`)

The `**` pattern matches **1 or more** directories:
- `**/*.js` matches `folder/file.js`, `a/b/c/file.js`
- `**/*.js` does **NOT** match `file.js` at the root level

#### Standard behavior (`doublestar`)

The `**` pattern matches **0 or more** directories:
- `**/*.js` matches `file.js`, `folder/file.js`, `a/b/c/file.js`
- This is consistent with most glob implementations

### When to Use

**Use `glob_matcher: doublestar` when:**
- You want standard glob behavior consistent with other tools
- You need `**` to match files at any level including the root
- You're migrating from other tools that use standard glob patterns

**Keep the default (`gobwas`) when:**
- You want to maintain existing behavior
- You specifically need `**` to require at least one directory level
- You have existing patterns that depend on the current behavior

### Example Comparison

```yml
# With default (gobwas)
glob_matcher: gobwas  # or omit this line

pre-commit:
  jobs:
    - run: eslint {staged_files}
      glob: "**/*.js"
      # Matches: src/app.js, lib/util.js
      # Does NOT match: app.js

    - run: eslint {staged_files}
      glob: "*.js"
      # Matches: app.js
      # Does NOT match: src/app.js
```

```yml
# With doublestar
glob_matcher: doublestar

pre-commit:
  jobs:
    - run: eslint {staged_files}
      glob: "**/*.js"
      # Matches: app.js, src/app.js, lib/util.js
```

### Notes

- The `glob_matcher` setting is global and applies to all `glob` and `exclude` patterns in your configuration
- This setting does not affect `root` or other path-related options
- The setting is fully backward compatible - existing configurations continue to work without modification
