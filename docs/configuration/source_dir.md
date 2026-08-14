---
title: "source_dir"
---

# `source_dir`

**Default: `.lefthook/`**

Change a directory for script files. The directory contains subfolders named after git hooks, each containing script files.

#### Example

```
.lefthook/
├── pre-commit/
│   ├── lint.sh
│   └── test.py
└── pre-push/
    └── check-files.rb
```

