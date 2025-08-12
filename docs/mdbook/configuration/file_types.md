## `file_types`

Filter files in a [`run`](./run.md) templates by their type. Special file types and MIME types are supported[^1]:

|File type| Exlanation|
|---------|-----------|
|`text`   | Any file that contains text. Symlinks are not followed. |
|`binary` | Any file that contains non-text bytes. Symlinks are not followed. |
|`executable` | Any file that has executable bits set. Symlinks are not followed. |
|`not executable` | Any file without executable bits in file mode. Symlinks included. |
|`symlink` | A symlink file. |
|`not symlink` | Any non-symlink file. |
|`text/html` | An HTML file. |
|`text/xml`  | An XML file. |
|`text/javascript` | A Javascript file. |
|`text/x-php` | A PHP file. |
|`text/x-lua` | A Lua file. |
|`text/x-perl` | A Perl file. |
|`text/x-python` | A Python file. |
|`text/x-shellscript` | Shell script file. |
|`text/x-sh` | Also shell script file. |
|`application/json` | JSON file. |

> **Important**
> The following types are applied using AND logic:
> - text
> - binary
> - executable
> - not executable
> - symlink
> - not symlink
>
> The mime types are applied using OR logic. So, you can have both `text/x-lua` and `text/x-sh`, but you can't specify both `symlink` and `not symlink`.

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

Check typos in scripts.

```yml
# lefthook.yml

pre-commit:
  jobs:
    - run: typos -w {staged_files}
      file_types:
        - text/x-perl
        - text/x-python
        - text/x-php
        - text/x-lua
        - text/x-sh
```

[^1]: All supported MIME types can be found here: [supported_mimes.md](https://github.com/gabriel-vasile/mimetype/blob/2f0854be3b9bbae4d449f8994d133f1c743f1885/supported_mimes.md)
