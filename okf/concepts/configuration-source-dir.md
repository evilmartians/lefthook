---
type: concept
title: source_dir
source: "https://lefthook.dev/configuration/source_dir/"
path: /configuration/source_dir/
updated: 2026-10-08
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-10-08T10:30:38.290Z"
---
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

