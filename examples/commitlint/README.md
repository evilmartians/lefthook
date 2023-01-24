# Use commitlint and/or commitzen

## Install dependencies

```bash
yarn add -D @commitlint/cli @commitlint/config-conventional
# If using commitzen
yarn add -D commitizen cz-conventional-changelog
```

## Configure

Setup `commitlint.config.js`. Conventional configuration:

```bash
echo "module.exports = {extends: ['@commitlint/config-conventional']};" > commitlint.config.js
```

If you are using commitzen, make sure to add this in `package.json`:

```json
"config": {
  "commitizen": {
    "path": "./node_modules/cz-conventional-changelog"
  }
}
```

## Test it

```bash
# You can type it without message, if you are using commitzen
git commit

# Or provide a commit message is using only commitlint
git commit -am 'fix: typo'
```
