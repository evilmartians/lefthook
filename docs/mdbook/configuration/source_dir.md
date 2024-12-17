### `source_dir`

**Default: `.lefthook/`**

Change a directory for script files. Directory for script files contains folders with git hook names which contain script files.

Example of directory tree:

```
.lefthook/
├── pre-commit/
│   ├── lint.sh
│   └── test.py
└── pre-push/
    └── check-files.rb
```

