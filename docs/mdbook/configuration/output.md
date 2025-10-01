## `output`

You can manage verbosity using the `output` config. You can specify what to print in your output by setting these values, which you need to have

Possible values are `meta,summary,success,failure,execution,execution_out,execution_info,skips`.
By default, all output values are enabled

You can also disable all output with setting `output: false`. In this case only errors will be printed.

This config quiets all outputs except for errors.

**Example**

```yml
# lefthook.yml

output:
  - meta           # Print lefthook version
  - summary        # Print summary block (successful and failed steps)
  - empty_summary  # Print summary heading when there are no steps to run
  - success        # Print successful steps
  - failure        # Print failed steps printing
  - execution      # Print any execution logs
  - execution_out  # Print execution output
  - execution_info # Print `EXECUTE > ...` logging
  - skips          # Print "skip" (i.e. no files matched)
```

You can also *extend* this list with an environment variable `LEFTHOOK_OUTPUT`:

```bash
LEFTHOOK_OUTPUT="meta,success,summary" lefthook run pre-commit
```
