const isEnabled = (value) => value && value !== "0" && value !== "false";
if (!isEnabled(process.env.CI) || isEnabled(process.env.LEFTHOOK)) {
  const { spawnSync } = require('child_process');
  const { getExePath } = require('./get-exe');

  // run install. No -f: --force also tells `install` to proceed when
  // core.hooksPath is set, and an unattended postinstall must not make that
  // call. See packaging/registries/npm/lefthook/postinstall.js.
  spawnSync(getExePath(), ['install'], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: 'inherit',
  });
}
