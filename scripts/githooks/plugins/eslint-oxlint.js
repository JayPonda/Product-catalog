module.exports = (engine) => {
  engine.registerPlugin('eslint-oxlint', (e, dir, files) => {
    let success = e.runCommand(dir, 'node', ['./node_modules/.bin/oxlint', ...files]);
    if (!success) {
      throw new Error('oxlint linting failed');
    }
    success = e.runCommand(dir, 'node', ['./node_modules/.bin/eslint', ...files]);
    if (!success) {
      throw new Error('eslint linting failed');
    }
  });
};
