## Scripts

Scripts are stored under `<source_dir>/<hook-name>/` folder. These scripts are your own executables which are being run in the project root.

To add a script for a `pre-commit` hook:

1. Run `lefthook add -d pre-commit`
1. Edit `.lefthook/pre-commit/my-script.sh`
1. Add an entry to `lefthook.yml`
   ```yml
   # lefthook.yml

   pre-commit:
     scripts:
       "my-script.sh":
         runner: bash
   ```

### Script options

- [`runner`](./runner.md)
- [`skip`](./skip.md)
- [`only`](./only.md)
- [`tags`](./tags.md)
- [`env`](./env.md)
- [`fail_text`](./fail_text.md)
- [`stage_fixed`](./stage_fixed.md)
- [`interactive`](./interactive.md)
- [`use_stdin`](./use_stdin.md)
- [`priority`](./priority.md)

### Example

Let's create a bash script to check commit templates `.lefthook/commit-msg/template_checker`:

```bash
INPUT_FILE=$1
START_LINE=`head -n1 $INPUT_FILE`
PATTERN="^(TICKET)-[[:digit:]]+: "
if ! [[ "$START_LINE" =~ $PATTERN ]]; then
  echo "Bad commit message, see example: TICKET-123: some text"
  exit 1
fi
```

Now we can ask lefthook to run our bash script by adding this code to
`lefthook.yml` file:

```yml
# lefthook.yml

commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

When you try to commit `git commit -m "bad commit text"` script `template_checker` will be executed. Since commit text doesn't match the described pattern the commit process will be interrupted.
