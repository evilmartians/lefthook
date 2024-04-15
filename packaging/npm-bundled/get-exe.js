const path = require("path");

function getRelativeExePath() {
  // Detect OS
  // https://nodejs.org/api/process.html#process_process_platform
  let goOS = process.platform;
  let extension = "";
  if (["win32", "cygwin"].includes(process.platform)) {
    goOS = "windows";
    extension = ".exe";
  }

  // Detect architecture
  // https://nodejs.org/api/process.html#process_process_arch
  let goArch = process.arch;
  let suffix = "";
  switch (process.arch) {
    case "x64": {
      goArch = "amd64";
      suffix = "_v1"; // GOAMD64
      break;
    }
    case "x32":
    case "ia32": {
      goArch = "386";
      break;
    }
  }

  return [`lefthook_${goOS}_${goArch}${suffix}`, `lefthook${extension}`];
}
function getExePath() {
  const dir = path.join(__dirname, "bin");
  const [folder, file] = getRelativeExePath();
  const executable = path.join(dir, folder, file);
  return executable;
}
exports.getRelativeExePath = getRelativeExePath;
exports.getExePath = getExePath;
