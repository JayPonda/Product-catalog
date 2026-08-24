module.exports = (engine) => {
  engine.registerPlugin('vitest-coverage', (e, dir, files) => {
    const success = e.runCommand(dir, 'pnpm', [
      'exec', 'vitest', 'run', '--coverage.enabled=true', '--coverage.thresholds.statements=90'
    ]);
    if (!success) {
      throw new Error('Frontend tests failed or statement coverage was below 90%');
    }
  });
};
