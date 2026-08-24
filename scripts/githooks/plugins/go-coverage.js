const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('go-coverage', (e, dir, files) => {
    const covPkgs = './src/repositories/...,./src/services/...,./src/controllers/...,./src/middleware/...,./src/routes/...,./utils/...';
    const success = e.runCommand(dir, 'go', [
      'test', './...', '-count=1', `-coverpkg=${covPkgs}`, '-coverprofile=cover.out'
    ]);
    if (!success) {
      throw new Error('Go backend tests failed to run');
    }

    // Parse coverage from `go tool cover -func=cover.out`
    const result = spawnSync('go', ['tool', 'cover', '-func=cover.out'], { cwd: dir });
    if (result.status !== 0) {
      throw new Error(`Failed to calculate coverage: ${result.stderr ? result.stderr.toString() : ''}`);
    }

    const output = result.stdout.toString();
    const lastLine = output.trim().split('\n').pop();
    const match = lastLine.match(/total:\s+\(statements\)\s+(\d+(?:\.\d+)?)%/);
    if (!match) {
      throw new Error(`Could not parse coverage percentage from output: ${lastLine}`);
    }

    const percent = parseFloat(match[1]);
    const rounded = Math.round(percent);
    e.logln(`📊 Backend Code Coverage: ${percent}% (rounded: ${rounded}%)`);
    if (rounded < 90) {
      throw new Error(`Go backend coverage is ${percent}% (rounded to ${rounded}%), which is below the required 90% threshold!`);
    }
  });
};
