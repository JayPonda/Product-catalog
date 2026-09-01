const fs = require('fs');
const path = require('path');

module.exports = (engine) => {
  engine.registerPlugin('eslint-oxlint', (e, dir, files) => {
    const existingFiles = files.filter(file => fs.existsSync(path.resolve(dir, file)));
    if (existingFiles.length === 0) {
      return;
    }
    let success = e.runCommand(dir, 'pnpm', ['exec', 'oxlint', ...existingFiles]);
    if (!success) {
      throw new Error('oxlint linting failed');
    }
    success = e.runCommand(dir, 'pnpm', ['exec', 'eslint', ...existingFiles]);
    if (!success) {
      throw new Error('eslint linting failed');
    }
  });
};

