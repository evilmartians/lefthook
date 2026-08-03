---
type: concept
title: Usage
source: "https://lefthook.dev/usage/"
path: /usage/
updated: 2026-08-03
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-03T09:42:48.620Z"
---
# Usage

Here are the most common usage cases. You can find more info in the docs.

## Basic CLI commands

```bash
# Create/update Git hooks based on lefthook.yml, or create an empty lefthook.yml
lefthook install

# Run pre-commit hook commands and scripts (requires lefthook.yml)
lefthook run pre-commit

# Validate the configuration
lefthook validate

# Dump the configuration (useful when you have remotes, extends that overwrite the configuration)
lefthook dump
```

## Skip running lefthook when committing changes

```bash
LEFTHOOK=0 git commit
```
