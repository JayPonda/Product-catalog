const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

class Engine {
  constructor(pluginsDir) {
    this.tasks = [];
    this.plugins = {};
    this.pluginsDir = pluginsDir;
    this.logFile = null;
    this.logPath = null;
  }

  register(task) {
    this.tasks.push(task);
  }

  registerPlugin(name, check) {
    this.plugins[name] = check;
  }

  log(msg) {
    process.stdout.write(msg);
    if (this.logFile) {
      fs.writeSync(this.logFile, msg);
    }
  }

  logln(msg = '') {
    this.log(msg + '\n');
  }

  runCommand(dir, name, args = []) {
    // Clean Git environment variables to avoid VCS resolution conflicts
    const env = { ...process.env };
    const gitVars = [
      'GIT_ALTERNATE_OBJECT_DIRECTORIES', 'GIT_CONFIG', 'GIT_CONFIG_PARAMETERS',
      'GIT_CONFIG_COUNT', 'GIT_OBJECT_DIRECTORY', 'GIT_DIR', 'GIT_WORK_TREE',
      'GIT_IMPLICIT_WORK_TREE', 'GIT_GRAFT_FILE', 'GIT_INDEX_FILE',
      'GIT_NO_REPLACE_OBJECTS', 'GIT_REPLACE_REF_BASE', 'GIT_PREFIX',
      'GIT_SHALLOW_FILE', 'GIT_COMMON_DIR'
    ];
    gitVars.forEach(v => delete env[v]);

    const result = spawnSync(name, args, {
      cwd: dir,
      env,
      stdio: this.logFile ? ['inherit', 'pipe', 'pipe'] : 'inherit'
    });

    if (result.stdout) {
      process.stdout.write(result.stdout);
      if (this.logFile) {
        fs.writeSync(this.logFile, result.stdout);
      }
    }
    if (result.stderr) {
      process.stderr.write(result.stderr);
      if (this.logFile) {
        fs.writeSync(this.logFile, result.stderr);
      }
    }

    if (result.error) {
      this.logln(`❌ Command execution error: ${result.error.message}`);
    }

    return result.status === 0;
  }

  loadPlugins() {
    if (!this.pluginsDir) return;
    const absPluginsDir = path.resolve(this.pluginsDir);
    if (!fs.existsSync(absPluginsDir)) return;

    const files = fs.readdirSync(absPluginsDir);
    files.forEach(file => {
      if (file.endsWith('.js')) {
        const pluginPath = path.join(absPluginsDir, file);
        const registerFn = require(pluginPath);
        if (typeof registerFn === 'function') {
          registerFn(this);
        }
      }
    });
  }

  run(filesResolver, args = []) {
    // Resolve log file path (.githooks/tmp/<hook-name>.log)
    const hookName = path.basename(process.argv[1], '.js');
    const logDir = path.resolve(__dirname, '../../.githooks/tmp');
    if (!fs.existsSync(logDir)) {
      fs.mkdirSync(logDir, { recursive: true });
    }
    this.logPath = path.join(logDir, `${hookName}.log`);

    try {
      this.logFile = fs.openSync(this.logPath, 'w');
    } catch (e) {
      console.warn(`⚠️ Warning: Could not create log file ${this.logPath}: ${e.message}. Running without file logging.`);
    }

    const exit = (code) => {
      if (this.logFile) {
        fs.closeSync(this.logFile);
        console.log(`📝 Full logs written to: .githooks/tmp/${hookName}.log. You can review it anytime!`);
      }
      process.exit(code);
    };

    if (!filesResolver) {
      this.logln('❌ Error: filesResolver is not configured.');
      exit(1);
    }

    let files;
    try {
      files = filesResolver(args);
    } catch (e) {
      this.logln(`❌ Error resolving files: ${e.message}`);
      exit(1);
    }

    if (!files || files.length === 0) {
      this.logln('ℹ️ No changed files found. Skipping checks.');
      exit(0);
    }

    // Resolve plugins
    this.loadPlugins();

    // Initialize status map for registered tasks
    const status = {};
    this.tasks.forEach(t => {
      status[t.name] = 'Pending';
    });

    let hasFailures = false;

    while (true) {
      let progress = false;
      let pendingCount = 0;

      for (const task of this.tasks) {
        if (status[task.name] !== 'Pending') continue;
        pendingCount++;

        // Check if dependencies are resolved
        let depFailed = false;
        let depNotResolved = false;

        for (const depName of (task.dependsOn || [])) {
          const depStatus = status[depName];
          if (!depStatus) continue; // Skip if depending on non-existent task
          if (depStatus === 'Pending') {
            depNotResolved = true;
          } else if (depStatus === 'Failed' || depStatus === 'Blocked') {
            depFailed = true;
          }
        }

        if (depNotResolved) continue;

        if (depFailed) {
          this.logln(`⏭️ Skipping task: ${task.name} (a dependency failed)`);
          this.logln('-'.repeat(50));
          status[task.name] = 'Blocked';
          progress = true;
          continue;
        }

        if (task.skip) {
          this.logln(`⏭️ Skipping task: ${task.name} (explicitly skipped in configuration)`);
          this.logln('-'.repeat(50));
          status[task.name] = 'Skipped';
          progress = true;
          continue;
        }

        // Gather files matching task filters
        const dirPath = task.dir ? path.resolve(task.dir) : process.cwd();
        const matched = [];
        for (const file of files) {
          if (task.dir) {
            const prefix = task.dir + '/';
            if (!file.startsWith(prefix)) continue;
          }

          if (this.matchPattern(file, task.pattern)) {
            const relPath = task.dir ? file.substring(task.dir.length + 1) : file;
            if (fs.existsSync(path.resolve(dirPath, relPath))) {
              matched.push(relPath);
            }
          }
        }

        if (matched.length === 0) {
          status[task.name] = 'Skipped';
          progress = true;
          continue;
        }

        const dirName = task.dir || 'root';
        this.logln(`🏃 Running task: ${task.name} (${matched.length} files matched in '${dirName}')...`);

        const checkFn = task.check || this.plugins[task.pluginName];

        if (!checkFn) {
          this.logln(`❌ Error: Task check implementation or plugin '${task.pluginName}' not found.`);
          status[task.name] = 'Failed';
          hasFailures = true;
          this.logln('-'.repeat(50));
          progress = true;
          continue;
        }

        try {
          checkFn(this, dirPath, matched);
          this.logln(`✅ Task '${task.name}' passed!`);
          status[task.name] = 'Success';
        } catch (e) {
          this.logln(`❌ Task '${task.name}' failed: ${e.message}`);
          status[task.name] = 'Failed';
          hasFailures = true;
        }
        this.logln('-'.repeat(50));
        progress = true;
      }

      if (pendingCount === 0) break;

      if (!progress) {
        this.logln('❌ Error: Circular or unresolved dependencies detected in tasks configuration!');
        exit(1);
      }
    }

    if (hasFailures) {
      this.logln('❌ Hook checks failed! Please fix the errors above.');
      exit(1);
    }

    this.logln('🎉 All checks passed successfully!');
    exit(0);
  }

  matchPattern(file, pattern) {
    if (!pattern) return true;
    const suffixes = pattern.split(',');
    return suffixes.some(suffix => file.endsWith(suffix.trim()));
  }
}

// Helper: Query staged files using git (useful for pre-commit)
function getStagedFiles(args = []) {
  if (args.length > 0) return args;

  const result = spawnSync('git', ['diff', '--cached', '--name-only', '--diff-filter=d']);
  if (result.status !== 0) throw new Error(result.stderr ? result.stderr.toString() : 'failed to run git diff');
  return parseLines(result.stdout.toString());
}

// Helper: Query files changed in current branch compared to remote tracking branch (useful for pre-push)
function getBranchChangedFiles() {
  // 1. Try to diff against upstream branch
  let result = spawnSync('git', ['diff', '--name-only', '--diff-filter=d', '@{u}...HEAD']);
  if (result.status === 0) return parseLines(result.stdout.toString());

  // 2. Fallback: diff against local main branch
  result = spawnSync('git', ['diff', '--name-only', '--diff-filter=d', 'main...HEAD']);
  if (result.status === 0) return parseLines(result.stdout.toString());

  // 3. Fallback: diff against origin/main
  result = spawnSync('git', ['diff', '--name-only', '--diff-filter=d', 'origin/main...HEAD']);
  if (result.status === 0) return parseLines(result.stdout.toString());

  // 4. Fallback: list all tracked files in repository to force checking everything
  result = spawnSync('git', ['ls-files']);
  if (result.status === 0) return parseLines(result.stdout.toString());

  throw new Error('failed to get changed files list from git');
}

function parseLines(output) {
  return output.trim().split('\n').map(l => l.trim()).filter(l => l !== '');
}

module.exports = {
  Engine,
  getStagedFiles,
  getBranchChangedFiles
};
