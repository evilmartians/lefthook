## `extends`

You can extend your config with another one YAML file. Its content will be merged. Extends for `lefthook.yml`, `lefthook-local.yml`, and [`remotes`](./remotes.md) configs are handled separately, so you can have different extends in these files.

You can use asterisk to make a glob.

**Example**

```yml
# lefthook.yml

extends:
  - /home/user/work/lefthook-extend.yml
  - /home/user/work/lefthook-extend-2.yml
  - lefthook-extends/file.yml
  - ../extend.yml
  - projects/*/specific-lefthook-config.yml
```

> The extends will be merged to the main configuration in your file. Here is the order of settings applied:
>
> - `lefthook.yml` – main config file
> - `extends` – configs specified in [extends](./extends.md) option
> - `remotes` – configs specified in [remotes](./remotes.md) option
> - `lefthook-local.yml` – local config file
>
> So, `extends` override settings from `lefthook.yml`, `remotes` override `extends`, and `lefthook-local.yml` can override everything.


