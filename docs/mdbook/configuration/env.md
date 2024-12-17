### `env`

You can specify some ENV variables for the command or script.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      env:
        RAILS_ENV: test
      run: bundle exec rspec
```

#### Extending PATH

If your hook is run by GUI program, and you use some PATH tweaks in your ~/.<shell>rc, you might see an error saying *executable not found*. In that case You can extend the **$PATH** variable with `lefthook-local.yml` configuration the following way.

```yml
# lefthook.yml

pre-commit:
  commands:
    test:
      run: yarn test
```

```yml
# lefthook-local.yml

pre-commit:
  commands:
    test:
      env:
        PATH: $PATH:/home/me/path/to/yarn
```

**Notes**

This option is useful when using lefthook on different OSes or shells where ENV variables are set in different ways.
