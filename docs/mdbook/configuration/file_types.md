## `file_types`

Filter files in a [`run`](./run.md) templates by their type. Supported types:

|File type| Exlanation|
|---------|-----------|
|`text`   | Any file that contains text. Symlinks are not followed. |
|`binary` | Any file that contains non-text bytes. Symlinks are not followed. |
|`executable` | Any file that has executable bits set. Symlinks are not followed. |
|`not executable` | Any file without executable bits in file mode. Symlinks included. |
|`symlink` | A symlink file. |
|`not symlink` | Any non-symlink file. |

> **Important:** When passed multiple file types all constraints will be applied to the resulting list of files

**Examples**

Apply some different linters on text and binary files.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint-code:
      run: yarn lint {staged_files}
      file_types: text
    check-hex-codes:
      run: yarn check-hex {staged_files}
      file_types: binary
```

Skip symlinks.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix {staged_files}
      file_types:
        - not symlink
```

Lint executable scripts.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix {staged_files}
      file_types:
        - executable
        - text
```
