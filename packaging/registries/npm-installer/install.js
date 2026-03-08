const http = require('https')
const fs = require('fs')
const path = require("path")
const chp = require("child_process")

const iswin = ["win32", "cygwin"].includes(process.platform)

async function install() {
  const isEnabled = (value) => value && value !== "0" && value !== "false";
  if (isEnabled(process.env.CI) && !isEnabled(process.env.LEFTHOOK)) {
    return
  }
  const downloadURL = getDownloadURL()
  const extension = iswin ? ".exe" : ""
  const fileName = `lefthook${extension}`
  const exePath = path.join(__dirname, "bin", fileName)
  await downloadBinary(downloadURL, exePath)
  console.log('downloaded to', exePath)
  if (!iswin) {
    fs.chmodSync(exePath, "755")
  }
  // run install
  chp.spawnSync(exePath, ['install',  '-f'], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: 'inherit',
  })
}

function getDownloadURL() {
  // Detect OS
  // https://nodejs.org/api/process.html#process_process_platform
  let goOS = process.platform
  let extension = ""
  if (iswin) {
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
  }
  const version = require("./package.json").version

  return `https://github.com/evilmartians/lefthook/releases/download/v${version}/lefthook_${version}_${downloadOS}_${arch}${extension}`
}

async function downloadBinary(url, dest) {
  console.log('downloading', url)
  const file = fs.createWriteStream(dest)
  return new Promise((resolve, reject) => {
    http.get(url, function(response) {
      if (response.statusCode === 302 && response.headers.location) {
        // If the response is a 302 redirect, follow the new location
        downloadBinary(response.headers.location, dest)
          .then(resolve)
          .catch(reject)
      } else {
        response.pipe(file)

        file.on('finish', function() {
          file.close(() => {
            resolve(dest)
          })
        })
      }
    }).on('error', function(err) {
      fs.unlink(file, () => {
        reject(err)
      })
    })
  })
}

// start:
install().catch((e) => {
  throw e
})
