## Git LFS support

::: callout info Note
If git-lfs binary is not installed and not required in your project, LFS hooks won't be executed, and you won't be warned about it.

Git LFS hooks may be slow. Disable them with the global `skip_lfs: true` setting.
:::

Lefthook runs LFS hooks internally for the following hooks:

- post-checkout
- post-commit
- post-merge
- pre-push

Errors are suppressed if git LFS is not required for the project. You can use [`LEFTHOOK_VERBOSE`](../envs/LEFTHOOK_VERBOSE.md) ENV to make lefthook show git LFS output.

To avoid calling LFS hooks set [`skip_lfs: true`](../../configuration/skip_lfs.md) in lefthook.yml or lefthook-local.yml
