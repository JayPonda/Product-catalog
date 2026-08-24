const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('branch-name', (e, dir, files) => {
    const result = spawnSync('git', ['rev-parse', '--abbrev-ref', 'HEAD'], { cwd: dir });
    if (result.status !== 0) {
      throw new Error(`Failed to get current branch name: ${result.stderr ? result.stderr.toString() : ''}`);
    }

    const branch = result.stdout.toString().trim();

    // Allow master/main commits directly, as well as detached HEAD (for CI runs)
    if (branch === 'main' || branch === 'master' || branch === 'HEAD') {
      return;
    }

    const allowedPrefixes = ['feature/', 'hotfix/', 'bugfix/', 'release/', 'chore/', 'refactor/'];
    const isValid = allowedPrefixes.some(prefix => branch.startsWith(prefix));

    if (!isValid) {
      throw new Error(`Branch name "${branch}" is invalid. It must start with one of: ${allowedPrefixes.join(', ')} (e.g. "feature/add-login" or "hotfix/fix-leak").`);
    }
  });
};
