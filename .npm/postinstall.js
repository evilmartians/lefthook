require("process");
const { spawn } = require("child_process");
const { createWriteStream, mkdirSync, unlink } = require("fs");
const { createGunzip } = require("zlib");
const { get } = require("https");
const { join, resolve } = require("path");

const { version } = require("./package.json");

const binaryName = "lefthook";
const packageDir = process.env.INIT_CWD || resolve(".");

function closeAndRemoveFile(file) {
  file.on("close", () => {
    unlink(file);
  });
  file.close();
}

function resolveDownloadPath(os, version, arch) {
  let archStr;
  let osStr;

  if (os === "darwin") {
    osStr = "MacOS";
  } else if (os === "linux") {
    osStr = "Linux";
  } else if (os === "win32") {
    if (arch === "ia32" || arch === "x86") {
      console.error("Lefthook is only supported on Windows with 64bit architecture.");
      process.exit(0);
    }
    osStr = "Windows";
  }

  if (arch === "x64") {
    archStr = "x86_64";
  } else if (arch === "ia32" || arch === "x86") {
    archStr = "i386";
  } else {
    console.error(`Unable to resolve architecture (${arch}), only supported architectures are: 'ia32', 'x86', and 'x64'`);
    process.exit(0);
  }

  return `https://github.com/Arkweid/lefthook/releases/download/v${version}/lefthook_${version}_${osStr}_${archStr}.gz`;
}

function downloadBinary(os, callback) {
  // construct single binary file path
  const binaryDir = join(packageDir, "node_modules", "@arkweid", "lefthook", "bin");
  const binaryPath = join(binaryDir, `${binaryName}${os === "win32" ? ".exe" : ""}`);

  // create directory if not existing, yet
  try {
    mkdirSync(binaryDir);
  } catch (err) {
    if (err.code !== 'EEXIST') {
      console.error(err.message);
      process.exit(0);
    }
  }

  // create write stream as file with the constructed binaryPath, set it
  // up to close as soon as no more data is expected
  const file = createWriteStream(binaryPath, { mode: 0o755 });
  file.on("finish", () => {
    file.close();
  });
  file.on("close", () => {
    callback(binaryPath);
  });

  // start download of binary file, set it up to remove the file if an
  // error occurs
  const downloadPath = resolveDownloadPath(os, version, process.arch);
  const download = get(downloadPath, downloadResponse => {
    // handle github binary redirect
    if (downloadResponse.statusCode === 302) {
      const redirectDownloadPath = downloadResponse.headers.location;

      const redirectDownload = get(redirectDownloadPath, redirectDownloadResponse => {
        if (redirectDownloadResponse.statusCode === 200) {
          const gzip = createGunzip();
          redirectDownloadResponse.pipe(gzip).pipe(file);
        } else {
          console.error("Unable to download lefthook binary file from redirect");
        }
      });

      redirectDownload.on("error", err => {
        console.error("Unable to download lefthook binary file");
        console.error(err.message);
        closeAndRemoveFile(file);
      });
    } else if (downloadResponse.statusCode === 200) {
      downloadResponse.pipe(file);
      downloadResponse.on("close", () => {
        callback(binaryPath);
      });
    } else {
      console.error("Unable to download lefthook binary file: Unhandled response", downloadResponse.statusCode);
    }
  });
  download.on("error", err => {
    console.error("Unable to download lefthook binary file");
    console.error(err.message);
    closeAndRemoveFile(file);
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
} else {
  console.log("Skipping downloading of `lefthook` binary due to CI environment.");
}
