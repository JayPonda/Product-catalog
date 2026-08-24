const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('gofmt', (e, dir, files) => {
    const result = spawnSync('gofmt', ['-l', ...files], { cwd: dir });
    if (result.status !== 0) {
      throw new Error(`gofmt command failed: ${result.stderr ? result.stderr.toString() : ''}`);
    }
    const unformatted = result.stdout.toString().trim();
    if (unformatted !== '') {
      e.logln('❌ The following Go files are not formatted:');
      e.logln(unformatted);
      throw new Error("unformatted Go files found; please run 'make format-backend'");
    }
  });
};
