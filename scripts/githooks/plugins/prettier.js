const fs = require('fs');
const path = require('path');

module.exports = (engine) => {
  engine.registerPlugin('prettier', (e, dir, files) => {
    const existingFiles = files.filter(file => fs.existsSync(path.resolve(dir, file)));
    if (existingFiles.length === 0) {
      return;
    }
    const success = e.runCommand(dir, 'pnpm', ['exec', 'prettier', '--check', ...existingFiles]);
    if (!success) {
      throw new Error("frontend formatting check failed; please run 'pnpm --dir app run format'");
    }
  });
};

