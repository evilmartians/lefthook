## `skip_lfs`

**Default:** `false`

Skip running LFS hooks even if it exists on your system.

### Example

```yml
# lefthook.yml

skip_lfs: true

pre-push:
  commands:
    test:
      run: yarn test
```
