#!/usr/bin/env node

var spawn = require('child_process').spawn;
var path = require('path');

var command_args = process.argv.slice(2);

var child = spawn(
    path.join(__dirname, 'lefthook'),
    command_args,
    { stdio: [process.stdin, process.stdout, process.stderr] });

child.on('close', function (code) {
    if (code !== 0) {
        process.exit(1);
    }
});
