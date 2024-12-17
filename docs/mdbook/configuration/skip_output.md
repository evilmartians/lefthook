### `skip_output`

> [!IMPORTANT]
> **DEPRECATED** This feature is deprecated and might be removed in future versions. Please, use `[output]` instead for managing verbosity.

You can manage the verbosity using the `skip_output` config. You can set whether lefthook should print some parts of its output.

Possible values are `meta,summary,success,failure,execution,execution_out,execution_info,skips`.

You can also disable all output with setting `skip_output: true`. In this case only errors will be printed.

This config quiets all outputs except for errors.

**Example**

```yml
# lefthook.yml

skip_output:
  - meta           # Skips lefthook version printing
  - summary        # Skips summary block (successful and failed steps) printing
  - empty_summary  # Skips summary heading when there are no steps to run
  - success        # Skips successful steps printing
  - failure        # Skips failed steps printing
  - execution      # Skips printing any execution logs (but prints if the execution failed)
  - execution_out  # Skips printing execution output (but still prints failed commands output)
  - execution_info # Skips printing `EXECUTE > ...` logging
  - skips          # Skips "skip" printing (i.e. no files matched)
```

You can also *extend* this list with an environment variable `LEFTHOOK_QUIET`:

```bash
LEFTHOOK_QUIET="meta,success,summary" lefthook run pre-commit
```

