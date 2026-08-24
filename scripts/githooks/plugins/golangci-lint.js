module.exports = (engine) => {
  engine.registerPlugin('golangci-lint', (e, dir, files) => {
    const success = e.runCommand(dir, 'go', ['run', 'github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5', 'run', './...']);
    if (!success) {
      throw new Error('Go linting failed');
    }
  });
};
