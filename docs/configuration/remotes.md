## `remotes`

You can provide multiple remote configs if you want to share yours lefthook configurations across many projects. Lefthook will automatically download and merge configurations into your local `lefthook.yml`.

You can use [`extends`](./extends.md) but the paths must be relative to the remote repository root.

If you provide [`scripts`](./scripts.md) in a remote config file, the [scripts](./source_dir.md) folder must also be in the **root of the repository**.

**Note**

The configuration from `remotes` will be merged to the local config using the following priority:

1. Local main config (`lefthook.yml`)
1. Remote configs (`remotes`)
1. Local overrides (`lefthook-local.yml`)

This priority may be changed in the future. For simplicity, try to keep jobs in remote settings independent from any other steps.
