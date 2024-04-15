if (!process.env.CI) {
  const { spawnSync } = require("child_process");
  const { getExePath, getRelativeExePath } = require("./get-exe");
  const { writeFileSync } = require("fs");
  const path = require("path");
  const pkg = require("./package.json");

  // run install
  spawnSync(getExePath(), ["install", "-f"], {
    cwd: process.env.INIT_CWD || process.cwd(),
    stdio: "inherit",
  });

  writeFileSync(
    path.resolve(__dirname, "package.json"),
    JSON.stringify(
      {
        ...pkg,
        bin: {
          dprint: "bin/" + getRelativeExePath().join("/"),
        },
      },
      null,
      2,
    ),
  );
}
