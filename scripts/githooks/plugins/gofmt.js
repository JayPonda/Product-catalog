const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('gofmt', (e, dir, files) => {
    const existingFiles = files.filter(file => fs.existsSync(path.resolve(dir, file)));
    if (existingFiles.length === 0) {
      return;
    }
    const result = spawnSync('gofmt', ['-l', ...existingFiles], { cwd: dir });
    if (result.status !== 0) {
      throw new Error(`gofmt command failed: ${result.stderr ? result.stderr.toString() : ''}`);
    }
    const unformatted = result.stdout.toString().trim();
    if (unformatted !== '') {
      e.logln('❌ The following Go files are not formatted:');
      e.logln(unformatted);
      throw new Error("unformatted Go files found; please run 'make format-backend'");
    }
  });
};
