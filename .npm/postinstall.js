require("process");
const { spawn } = require("child_process");
const { createWriteStream, unlink } = require("fs");
const { get } = require("https");
const { join } = require("path");

const downloadPaths = {
  "darwin": "https://raw.githubusercontent.com/Arkweid/lefthook/master/.npm/bin/lefthook-mac",
  "linux": "https://raw.githubusercontent.com/Arkweid/lefthook/master/.npm/bin/lefthook-linux",
  "win32": "https://raw.githubusercontent.com/Arkweid/lefthook/master/.npm/bin/lefthook-win.exe"
};

function downloadBinary(os, callback) {
  // construct single binary file path
  const binaryPath = join(process.cwd(), 'node_modules', '@arkweid', 'lefthook', 'bin', binary);

  // create write stream as file with the constructed binaryPath, set it
  // up to close as soon as no more data is expected
  const file = createWriteStream(binaryPath, { mode: 0o755 });
  file.on("finish", () => {
    file.close();
  });

  // start download of binary file, set it up to remove the file if an
  // error occurs
  const download = get(downloadPaths[os], response => {
    response.pipe(file);
    response.on("end", () => {
      callback(binaryPath);
    });
  });
  download.on("error", err => {
    console.error("Unable to download lefthook binary file");
    console.error(err.message);
    unlink(file);
  });
}

function installGitHooks(binaryPath) {
  console.log("Installing git hooks");
  spawn(binaryPath, ["install", "-f"], { stdio: [process.stdin, process.stdout, process.stderr] });
}

const isCI = process.env.CI;
if (!isCI) {
  process.chdir(process.env.INIT_CWD);

  if (!["darwin", "linux", "win32"].includes(process.platform)) {
    console.log("Unsupported OS");
    process.exit(0);
  }

  console.log("Downloading `lefthook` binary");
  downloadBinary(process.platform, installGitHooks);
}
