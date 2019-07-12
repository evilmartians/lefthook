require("process");
const { spawn } = require("child_process");
const { join } = require("path");

const isCI = process.env.CI;
if (!isCI) {
  process.chdir(process.env.INIT_CWD);

  if (process.platform === 'darwin') {
    binary = 'lefthook-mac';
  } else if (process.platform === 'linux') {
      binary = 'lefthook-linux';
  } else if (process.platform === 'win32') {
      binary = 'lefthook-win.exe';
  } else {
    console.log("Unsupported OS");
    process.exit(0);
  }

  binpath = join(process.cwd(), 'node_modules', '@arkweid', 'lefthook', 'bin', binary);

  result = spawn(binpath, ["install", "-f"], { stdio: [process.stdin, process.stdout, process.stderr] });
}
