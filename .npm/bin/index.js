#!/usr/bin/env node

var spawn = require('child_process').spawn;
const path = require("path")
const extension = ["win32", "cygwin"].includes(process.platform) ? ".exe" : ""
const exePath = path.join(__dirname, `lefthook${extension}`)

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
