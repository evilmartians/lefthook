const { spawnSync } = require("child_process")

function install() {
  if (process.env.CI) {
    return
  }

  const lefthook = path.join(__dirname, "bin", "lefthook")

  spawnSync(lefthook, ["install", "-f"], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: "inherit",
  })
}

try {
  install()
} catch(e) {
  console.warn("'lefthook install' command failed. Try running it manually")
}
