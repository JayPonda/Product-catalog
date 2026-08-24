const fs = require('fs');

module.exports = (engine) => {
  engine.registerPlugin('commit-msg-format', (e, dir, files) => {
    const msgPath = files[0];
    if (!msgPath || !fs.existsSync(msgPath)) {
      throw new Error(`Commit message file not found at path: ${msgPath}`);
    }

    const msg = fs.readFileSync(msgPath, 'utf8').trim();

    // Ignore special merge/rebase commits
    if (msg.startsWith('Merge ') || msg.startsWith('Revert ') || msg.startsWith('squash!') || msg.startsWith('fixup!')) {
      return;
    }

    const conventionalCommitRegex = /^(feat|fix|chore|docs|style|refactor|perf|test)(\(.+\))?!?: .+/;
    if (!conventionalCommitRegex.test(msg)) {
      throw new Error(`Commit message format is invalid: "${msg}"\nMust match Conventional Commits format, e.g. "feat: add user login" or "fix(auth): correct token refresh".`);
    }
  });
};
