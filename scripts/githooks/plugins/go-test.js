module.exports = (engine) => {
  engine.registerPlugin('go-test', (e, dir, files) => {
    const success = e.runCommand(dir, 'go', ['test', './...', '-count=1']);
    if (!success) {
      throw new Error('Go backend tests failed');
    }
  });
};
