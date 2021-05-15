#!/usr/bin/env node

var spawn = require('child_process').spawn;
const path = require("path")
const exePath = path.join(__dirname, `lefthook${["win32", "cygwin"].includes(process.platform) ? ".exe" : ""}`)

var command_args = process.argv.slice(2);
var child = spawn(
    exePath,
    command_args,
    { stdio: "inherit" });

child.on('close', function (code) {
    if (code !== 0) {
        process.exit(1);
    }
});
