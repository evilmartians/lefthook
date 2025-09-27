## `lefthook version`

`lefthook version` prints the current binary version. Print the commit hash with `lefthook version --full`

You can also use the global `--version` flag as a shortcut:

- `lefthook --version` - prints the version
- `lefthook --version=full` - prints the version with commit hash
- `lefthook -V` - short form of `--version`

**Examples**

```bash
$ lefthook version --full
1.1.3 bb099d13c24114d2859815d9d23671a32932ffe2

$ lefthook --version
1.1.3

$ lefthook --version=full
1.1.3 bb099d13c24114d2859815d9d23671a32932ffe2
```

