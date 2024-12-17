### `rc`

Provide an [**rc**](https://www.baeldung.com/linux/rc-files) file, which is actually a simple `sh` script. Currently it can be used to set ENV variables that are not accessible from non-shell programs.

**Example**

Use cases:

- You have a GUI program that runs git hooks (e.g., VSCode)
- You reference executables that are accessible only from a tweaked $PATH environment variable (e.g., when using rbenv or nvm, fnm)
- Or even if your GUI program cannot locate the `lefthook` executable :scream:
- Or if you want to use ENV variables that control the executables behavior in `lefthook.yml`

```bash
# An npm executable which is managed by nvm
$ which npm
/home/user/.nvm/versions/node/v15.14.0/bin/npm
```

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: npm run eslint {staged_files}
```

Provide a tweak to access `npm` executable the same way you do it in your ~/<shell>rc.

```yml
# lefthook-local.yml

# You can choose whatever name you want.
# You can share it between projects where you use lefthook.
# Make sure the path is absolute.
rc: ~/.lefthookrc
```

Or

```yml
# lefthook-local.yml

# If the path contains spaces, you need to quote it.
rc: '"${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc"'
```

In the rc file, export any new environment variables or modify existing ones.

```bash
# ~/.lefthookrc

# An nvm way
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# An fnm way
export FNM_DIR="$HOME/.fnm"
[ -s "$FNM_DIR/fnm.sh" ] && \. "$FNM_DIR/fnm.sh"

# Or maybe just
PATH=$PATH:$HOME/.nvm/versions/node/v15.14.0/bin
```

```bash
# Make sure you updated git hooks. This is important.
$ lefthook install -f
```

Now any program that runs your hooks will have a tweaked PATH environment variable and will be able to get `nvm` :wink:
