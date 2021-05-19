const { spawnSync } = require("child_process")

async function install() {
  if (process.env.CI) {
    return
  }
  const exePath = await downloadBinary()
  // run install
  spawnSync(exePath, ["install", "-f"], {
    cwd: process.env.INIT_CWD || process.cwd,
    stdio: "inherit",
  })
}

function getDownloadURL() {
  // Detect OS
  // https://nodejs.org/api/process.html#process_process_platform
  let goOS = process.platform
  let extension = ""
  if (["win32", "cygwin"].includes(process.platform)) {
    goOS = "windows"
    extension = ".exe"
  }

  // Convert the goOS to the os name in the download URL
  let downloadOS = goOS === "darwin" ? "macOS" : goOS
  downloadOS = `${downloadOS.charAt(0).toUpperCase()}${downloadOS.slice(1)}`

  // Detect architecture
  // https://nodejs.org/api/process.html#process_process_arch
  let arch = process.arch
  switch (process.arch) {
    case "x64": {
      arch = "x86_64"
      break
    }
    case "x32":
    case "ia32": {
      arch = "i386"
      break
    }
  }
  const version = require("./package.json").version

  return `https://github.com/evilmartians/lefthook/releases/download/v${version}/lefthook_${version}_${downloadOS}_${arch}${extension}`
}

const { DownloaderHelper } = require("node-downloader-helper")
const path = require("path")

async function downloadBinary() {
  // TODO zip the binaries to reduce the download size
  const downloadURL = getDownloadURL()
  const extension = path.extname(downloadURL)
  const fileName = `lefthook${extension}`
  const binDir = path.join(__dirname, "bin")
  const dl = new DownloaderHelper(downloadURL, binDir, {
    fileName,
    retry: { maxRetries: 5, delay: 50 },
  })
  dl.on("end", () => console.log("lefthook binary was downloaded"))
  await dl.start()
  return path.join(binDir, fileName)
}

// start:
install().catch((e) => {
  throw e
})
