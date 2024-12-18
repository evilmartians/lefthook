## `assert_lefthook_installed`

**Default: `false`**

When set to `true`, fail (with exit status 1) if `lefthook` executable can't be found in $PATH, under node_modules/, as a Ruby gem, or other supported method. This makes sure git hook won't omit `lefthook` rules if `lefthook` ever was installed.
