module.exports = (engine) => {
  engine.registerPlugin('vitest', (e, dir, files) => {
    const success = e.runCommand(dir, 'pnpm', ['exec', 'vitest', 'run']);
    if (!success) {
      throw new Error('frontend tests failed');
    }
  });
};
