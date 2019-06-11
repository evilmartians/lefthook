require('process');
const { spawn } = require('child_process');

process.chdir(process.env.INIT_CWD);

result = spawn('npx', ['lefthook', 'install']);

result.stdout.on('data', (data) => {
    console.log(`${data}`);
});
