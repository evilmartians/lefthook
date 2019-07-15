# Change log

## master (unreleased)

## 0.6.3 (2019-07-15)

- [Commit](https://github.com/Arkweid/lefthook/commit/0426936f48f248221126f15619932b0dc8c54d7a) Add `-a` means `aggresive` strategy for `install` command
```bash
lefthook install -a # clear .git/hooks dir and reinstall lefthook hooks
```

- [Commit](https://github.com/Arkweid/lefthook/commit/5efb0677a4a9ec1728d3cf1a083075e23315a796) Add Lefthook version indicator for commands and script execution

- [Commit](https://github.com/Arkweid/lefthook/commit/8b55d91eed46643a1674bd4ad96fa211a177e159) Remove `npx` as dependency from node wrapper

Now we will call directly binary from `./node_modules`

- [Commit](https://github.com/Arkweid/lefthook/commit/76ffed4c698bc074984e91f5610c0b98784bd10b) Add `-f` means `force` strategy for `install` command

```bash
lefthook install -f # reinstall lefthook hooks without sync info check
```

- PR [#27](https://github.com/Arkweid/lefthook/pull/27) Move LEFTHOOK env check in hooks files

Now if LEFTHOOK=0 we will not call the binary file

- PR [#26](https://github.com/Arkweid/lefthook/pull/26) + [commit](https://github.com/Arkweid/lefthook/commit/afd67f94631a10975209ed4c5fabc763f44280eb) Add `{push_files}` shortcut

Add shortcut `{push_files}`

```
pre-commit:
  commands:
    rubocop:
      run: rubocop {push_files}
```
It same as:
```
pre-commit:
  commands:
    rubocop:
      files: git diff --name-only HEAD @{push} || git diff --name-only HEAD master
      run: rubocop {push_files}
```

- [Commit](https://github.com/Arkweid/lefthook/commit/af087b032a14952aa1dd235a3d0b5a51bc760a10) Add `min_version` option

You can mark your config for minimum Lefthook version:
```
min_version: 0.6.1
```

## 0.6.0 (2019-07-10)

- PR [#24](https://github.com/palkan/logidze/pull/110) Wrap `run` command in shell context.

Now in `run` option available `sh` syntax.

```
pre-commit:
  commands:
    bashed:
      run: rubocop -a && git add
```
Will be executed in this way:
```
sh -c "rubocop -a && git add"
```

- PR [#23](https://github.com/Arkweid/lefthook/pull/24) Search Lefthook in Gemfile.

Now it's possible to use Lefthook from Gemfile.

```ruby
# Gemfile

gem 'lefthook'
```
