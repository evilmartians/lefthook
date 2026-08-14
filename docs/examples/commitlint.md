# Commitlint and commitzen

Use lefthook to generate commit messages using commitzen and validate them with commitlint.

## Install dependencies

```bash
yarn add -D @commitlint/cli @commitlint/config-conventional

# For commitzen
yarn add -D commitizen cz-conventional-changelog
```

## Configure

Setup `commitlint.config.js`. Conventional configuration:

```js
// commitlint.config.js

module.exports = {extends: ['@commitlint/config-conventional']};
```

If you are using commitzen, make sure to add this in `package.json`:

```json
"config": {
  "commitizen": {
    "path": "./node_modules/cz-conventional-changelog"
  }
}
```

Configure lefthook:

```yml
# lefthook.yml

# Build commit messages
prepare-commit-msg:
  commands:
    commitzen:
      interactive: true
      run: yarn run cz --hook # Or npx cz --hook
      env:
        LEFTHOOK: 0

# Validate commit messages
commit-msg:
  commands:
    "lint commit message":
      run: yarn run commitlint --edit {1}
```


## Test it

```bash
# You can type it without message, if you are using commitzen
git commit

# Or provide a commit message is using only commitlint
git commit -am 'fix: typo'
```
