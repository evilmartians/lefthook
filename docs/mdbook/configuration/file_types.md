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
|`text/plain` | A plain text file. |
|`text/html` | An HTML file. |
|`text/xml`  | An XML file. |
|`text/x-php` | A PHP file. |
|`text/javascript` | A Javascript file. |
|`text/x-lua` | A Lua file. |
|`text/x-perl` | A Perl file. |
|`text/x-python` | A Python file. |
|`text/rtf` | A RTF text file. |
|`text/x-tcl` | A TCL file. |
|`text/csv` | A CSV file. |
|`text/tab-separated-values` | A TSV file. |
|`text/vcard` | A VCF file. |
|`text/calendar` | An iCal file. |
|`text/vtt` | Subtitles file. |
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
