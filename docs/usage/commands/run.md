---
title: "lefthook run"
---

## `lefthook run`

Executes the commands and scripts configured for a given hook. Installed Git hooks call `lefthook run` implicitly.

#### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint
      run: yarn lint --fix {staged_files}

test:
  jobs:
    - name: test
      run: yarn test
```

Install the hook.

```bash
$ lefthook install
```

```bash
$ lefthook run test # will run 'yarn test'
$ git commit # will run pre-commit hook ('yarn lint --fix')
$ lefthook run pre-commit # will run pre-commit hook (`yarn lint --fix`)
```

### Run specific jobs

You can specify which jobs to run (also `--tag` supported).

```bash
$ lefthook run pre-commit --job lints --job pretty --tag checks
```

### Specify files

You can force replacing files templates (like `{staged_files}`) with either all files (will acts as `{all_files}` template) or a list of files.

```bash
$ lefthook run pre-commit --all-files
$ lefthook run pre-commit --file file1.js --file file2.js
```

(if both are specified, `--all-files` is ignored)

### Machine-readable output

Pass `--output-format` when a wrapper, IDE plugin or CI job needs to know the outcome of each job without scraping the terminal output. The default, `text`, is unchanged.

```bash
$ lefthook run pre-commit --output-format json
$ lefthook run pre-commit --output-format jsonl
```

`json` prints one summary document when the run ends:

```json
{"hook":"pre-commit","status":"failed","duration_ms":82,"jobs":[{"name":"lint","status":"passed","duration_ms":7},{"name":"test","status":"failed","exit_code":3,"duration_ms":6},{"name":"docs","status":"skipped","duration_ms":0,"skip_reason":"by condition"}]}
```

`jsonl` prints one job per line as soon as that job finishes, which lets a consumer follow a long run. Each line is a job record from the `jobs` list above; no summary is printed.

Job records carry:

- `name` — the job name, or its `run`/`script` value when it has no name;
- `group_path` — names of the enclosing groups, outermost first (omitted for top-level jobs);
- `status` — `passed`, `failed`, `skipped` or `error`. A job that ran and exited non-zero is `failed`; a job that could not run, timed out, or failed to build its command is `error`;
- `exit_code` — the exit code of the job's command, when it ran;
- `duration_ms` — how long the job took;
- `skip_reason` — why the job was skipped (`by condition`, `no matching staged files`, `not selected by --job`, `broken pipe`, ...);
- `error` — the failure text for an `error` status.

The summary document adds `hook`, the combined `status` of the run, and `duration_ms` for the whole run; `error` is set when the run failed before producing results.

With either machine format the human-readable log and the job output move to standard error, so standard output stays parseable. Exit codes are unaffected: the run still exits non-zero when a job fails.

