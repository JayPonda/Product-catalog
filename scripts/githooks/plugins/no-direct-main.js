const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('no-direct-main', (e, dir, files) => {
    const result = spawnSync('git', ['rev-parse', '--abbrev-ref', 'HEAD'], { cwd: dir });
    if (result.status !== 0) {
      throw new Error(`Failed to get current branch name: ${result.stderr ? result.stderr.toString() : ''}`);
    }

    const branch = result.stdout.toString().trim();

    if (branch === 'main' || branch === 'master') {
      throw new Error(`Direct commits to "${branch}" branch are disabled. Please work inside a feature branch (e.g. "feature/add-login") and merge via Pull Request.`);
    }
  });
};
