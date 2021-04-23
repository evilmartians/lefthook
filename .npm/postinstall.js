const { spawn } = require("child_process");
const { join } = require("path");

const isCI = process.env.CI;

if (!isCI) {
  binpath = join(__dirname, 'bin', 'lefthook');

  result = spawn(binpath, ["install", "-f"], {
      cwd: process.env.INIT_CWD,
      stdio: [process.stdin, process.stdout, process.stderr]
  });
}
