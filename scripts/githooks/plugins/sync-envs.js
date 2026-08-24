const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

// Directories never worth descending into while hunting for .env files.
const SKIP_DIRS = new Set([
  'node_modules',
  'vendor',
  '.git',
  '.githooks',
  '.vscode',
  'dist',
  'build',
  'coverage',
  'tmp',
]);

// Recursively collect paths of files named exactly ".env", keyed by their
// containing directory. Variants like .env.production are ignored on purpose:
// this check guards the canonical local/env-example pair.
function findEnvDirs(root, dir, out) {
  let entries;
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch (err) {
    return out; // unreadable directory: skip rather than break the push
  }

  for (const entry of entries) {
    if (entry.isDirectory()) {
      if (entry.name.startsWith('.') || SKIP_DIRS.has(entry.name)) continue;
      findEnvDirs(root, path.join(dir, entry.name), out);
      continue;
    }
    if (entry.isFile() && entry.name === '.env') {
      out.push(dir);
    }
  }
  return out;
}

// Extract the set of KEYS from dotenv-style content.
// Values are irrelevant here by design ("no restrictions on values").
function parseEnvKeys(filePath) {
  const keys = new Set();
  const lines = fs.readFileSync(filePath, 'utf8').split(/\r?\n/);

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === '' || line.startsWith('#')) continue;

    // Allow an optional "export " prefix (shell-style dotenv files).
    const assignment = line.replace(/^export\s+/, '');
    const eq = assignment.indexOf('=');
    if (eq <= 0) continue; // no "=" or empty key

    const key = assignment.slice(0, eq).trim();
    if (key !== '') keys.add(key);
  }
  return keys;
}

module.exports = (engine) => {
  engine.registerPlugin('sync-envs', (e, dir) => {
    // Locate the repository root so every nested .env is covered,
    // regardless of which subdirectory the hook was invoked from.
    const rootResult = spawnSync('git', ['rev-parse', '--show-toplevel'], { cwd: dir });
    const root =
      rootResult.status === 0
        ? rootResult.stdout.toString().trim()
        : path.resolve(dir);

    const envDirs = findEnvDirs(root, root, []);
    if (envDirs.length === 0) {
      e.logln('ℹ️ No .env files found; nothing to sync.');
      return;
    }

    const problems = [];

    for (const envDir of envDirs) {
      const envPath = path.join(envDir, '.env');
      const examplePath = path.join(envDir, '.env.example');
      const relDir = path.relative(root, envDir) || '.';

      // A committed .env.example without a local .env is normal
      // (fresh clone before copying). Nothing to compare yet.
      if (!fs.existsSync(examplePath)) {
        problems.push(
          `${relDir}/.env exists but ${relDir}/.env.example is missing. ` +
            `Create it and list the same keys (values can be placeholders).`
        );
        continue;
      }

      const envKeys = parseEnvKeys(envPath);
      const exampleKeys = parseEnvKeys(examplePath);

      // Keys living only in .env: invisible to teammates AND to CI/deploys,
      // because .env is never committed. This is the classic silent gap.
      const missingInExample = [...envKeys].filter((k) => !exampleKeys.has(k));
      for (const key of missingInExample) {
        problems.push(
          `${relDir}/.env has key "${key}" but it is missing in ` +
            `${relDir}/.env.example. Add it there so others discover it.`
        );
      }

      // Keys only in the example: the template drifted ahead of your local
      // file (e.g. after pulling someone else's change).
      const missingInEnv = [...exampleKeys].filter((k) => !envKeys.has(k));
      for (const key of missingInEnv) {
        problems.push(
          `${relDir}/.env.example declares "${key}" but it is missing in ` +
            `${relDir}/.env. Add it locally (any value) to stay aligned.`
        );
      }
    }

    if (problems.length > 0) {
      e.logln(`🔍 Found ${problems.length} env sync problem(s):`);
      for (const problem of problems) {
        e.logln(`   • ${problem}`);
      }
      throw new Error('.env and .env.example files are out of sync. Fix the gaps listed above, then push again.');
    }

    e.logln(`✅ All ${envDirs.length} .env file(s) match their .env.example counterparts.`);
  });
};
