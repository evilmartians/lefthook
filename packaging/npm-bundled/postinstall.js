const isEnabled = (value) => value && value !== "0" && value !== "false";
if (!isEnabled(process.env.CI) || isEnabled(process.env.LEFTHOOK)) {
  const { spawnSync } = require('child_process');
  const { getExePath } = require('./get-exe');

  // run install
  spawnSync(getExePath(), ['install', '-f'], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: 'inherit',
  });
}
