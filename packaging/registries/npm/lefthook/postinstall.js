const { spawnSync } = require("child_process");
const { getExePath } = require("./get-exe");

function install() {
  const isEnabled = (value) => value && value !== "0" && value !== "false";
  if (isEnabled(process.env.CI) && !isEnabled(process.env.LEFTHOOK)) {
    return
  }

  // No -f here. --force also tells `install` to proceed when core.hooksPath is
  // set, and an unattended postinstall must not make that call: npm runs this
  // from wherever it happens to be, including a temp clone of a git dependency,
  // where the user's global hooks directory is what would get overwritten.
  spawnSync(getExePath(), ["install"], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: "inherit",
  });
}

try {
  install();
} catch (e) {
  console.warn(
    "'lefthook install' command failed. Try running it manually.\n" + e,
  );
}
