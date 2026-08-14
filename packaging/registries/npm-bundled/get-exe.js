const path = require("path")

function getExePath() {
  // Detect OS
  // https://nodejs.org/api/process.html#process_process_platform
  let goOS = process.platform;
  let extension = '';
  if (['win32', 'cygwin'].includes(process.platform)) {
    goOS = 'windows';
    extension = '.exe';
  }

  // Detect architecture
  // https://nodejs.org/api/process.html#process_process_arch
  let goArch = process.arch;
  let suffix = '';
  switch (process.arch) {
    case 'x32':
    case 'ia32': {
      goArch = '386';
      break;
    }
  }

  const dir = path.join(__dirname, 'bin');
  const executable = path.join(
    dir,
    `lefthook-${goOS}-${goArch}`,
    `lefthook${extension}`
  );
  return executable;
}
exports.getExePath = getExePath;
