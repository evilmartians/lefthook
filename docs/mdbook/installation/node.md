## Node.js

```bash
npm install --save-dev lefthook
```

```bash
yarn add --dev lefthook
```

```bash
pnpm add -D lefthook
```

> **Note:** If you use `pnpm` package manager make sure to update `pnpm-workspace.yaml`s `onlyBuiltDependencies` with `lefthook` and add `lefthook` to `pnpm.onlyBuiltDependencies` in your root `package.json`, otherwise the `postinstall` script of the `lefthook` package won't be executed and hooks won't be installed.

**Note**: lefthook has three NPM packages with different ways to deliver the executables

 1. [lefthook](https://www.npmjs.com/package/lefthook) installs one executable for your system

    ```bash
    npm install --save-dev lefthook
    ```

 1. **legacy**[^1] [@evilmartians/lefthook](https://www.npmjs.com/package/@evilmartians/lefthook)  installs executables for all OS

    ```bash
    npm install --save-dev @evilmartians/lefthook
    ```

 1. **legacy**[^1] [@evilmartians/lefthook-installer](https://www.npmjs.com/package/@evilmartians/lefthook-installer) fetches the right executable on installation

    ```bash
    npm install --save-dev @evilmartians/lefthook-installer
    ```
[^1]: Legacy distributions are still maintained but they will be shut down in the future.
