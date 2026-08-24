module.exports = (engine) => {
  engine.registerPlugin('eslint-oxlint', (e, dir, files) => {
    let success = e.runCommand(dir, 'pnpm', ['exec', 'oxlint', ...files]);
    if (!success) {
      throw new Error('oxlint linting failed');
    }
    success = e.runCommand(dir, 'pnpm', ['exec', 'eslint', ...files]);
    if (!success) {
      throw new Error('eslint linting failed');
    }
  });
};
