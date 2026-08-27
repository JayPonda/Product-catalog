module.exports = (engine) => {
  engine.registerPlugin('golangci-lint', (e, dir, files) => {
    const success = e.runCommand(dir, 'go', ['run', 'github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1', 'run', './...']);
    if (!success) {
      throw new Error('Go linting failed');
    }
  });
};
