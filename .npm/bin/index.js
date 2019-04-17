#!/usr/bin/env node

var spawn = require('child_process').spawn;
var path = require('path');

var command_args = process.argv.slice(2);

function spawnCommand(binaryExecutable) {
    var child = spawn(
        path.join(__dirname, binaryExecutable),
        command_args,
        { stdio: [process.stdin, process.stdout, process.stderr] });

    child.on('close', function (code) {
        if (code !== 0) {
            process.exit(1);
        }
    });
}

if (process.platform === 'darwin') {
    spawnCommand('lefthook-mac');
} else if (process.platform === 'linux') {
    spawnCommand('lefthook-linux');
} else if (process.platform === 'win32') {
    spawnCommand('lefthook-win.exe');
} else {
    console.log("Unsupported OS");
    process.exit(1);
}
