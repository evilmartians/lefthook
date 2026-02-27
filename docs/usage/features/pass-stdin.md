## Pass stdin to a command or script

When you need to read the data from stdin – specify [`use_stdin: true`](../../configuration/use_stdin.md). This option is good when you write a command or script that receives data from git using stdin (for the `pre-push` hook, for example).


