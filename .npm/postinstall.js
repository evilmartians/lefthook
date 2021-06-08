if (!process.env.CI) {
  const { spawnSync } = require('child_process');
  const { getExePath } = require('./get-exe');

  // run install
  spawnSync(getExePath(), ['install', '-f'], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: 'inherit',
  });
}
