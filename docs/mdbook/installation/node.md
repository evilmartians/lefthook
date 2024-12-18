## Node.js

Lefthook is available on NPM in the following flavors:

 1. [lefthook](https://www.npmjs.com/package/lefthook) that will install the proper binary:

    ```bash
    npm install lefthook --save-dev
    # or yarn:
    yarn add -D lefthook
    ```

 1. [@evilmartians/lefthook](https://www.npmjs.com/package/@evilmartians/lefthook) with pre-bundled binaries for all architectures:

    ```bash
    npm install @evilmartians/lefthook --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook
    ```

 1. [@evilmartians/lefthook-installer](https://www.npmjs.com/package/@evilmartians/lefthook-installer) that will fetch binary file on installation:

    ```bash
    npm install @evilmartians/lefthook-installer --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook-installer
    ```

> **Note:** If you use `pnpm` package manager make sure you set `side-effects-cache = false` in your .npmrc, otherwise the postinstall script of the lefthook package won't be executed and hooks won't be installed.
