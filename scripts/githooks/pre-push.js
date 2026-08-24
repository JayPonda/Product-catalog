const fs = require('fs');
const path = require('path');
const { Engine, getBranchChangedFiles } = require('./engine');

// Helper to read Git ref updates from stdin without blocking on interactive TTYs
function readStdinLines() {
  if (process.stdin.isTTY) {
    return [];
  }
  try {
    const buffer = fs.readFileSync(0);
    return buffer.toString().trim().split('\n').filter(Boolean);
  } catch (e) {
    return [];
  }
}

const e = new Engine(path.join(__dirname, 'plugins'));
e.FileResolver = getBranchChangedFiles;
e.stdinLines = readStdinLines();

// === Linear Check Chain (Tags, then Backend, then Frontend) ===

// 1. Tag Validation Check (always executes first)
e.register({
  name: 'Tag Validation Check',
  pattern: '',
  pluginName: 'tag-validator'
});

// 2. Go Formatting Check
e.register({
  name: 'Go Formatting Check',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Tag Validation Check'],
  pluginName: 'gofmt'
});

// 3. Go Linting
e.register({
  name: 'Go Linting',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Go Formatting Check'],
  pluginName: 'golangci-lint'
});

// 4. Go Backend Tests
e.register({
  name: 'Go Backend Tests',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Go Linting'],
  pluginName: 'go-coverage'
});

// 5. Frontend Formatting Check (runs after backend tests)
e.register({
  name: 'Frontend Formatting Check',
  dir: 'app',
  pattern: '.js,.ts,.vue,.css,.json',
  dependsOn: ['Go Backend Tests'],
  pluginName: 'prettier'
});

// 6. Frontend Linting
e.register({
  name: 'Frontend Linting',
  dir: 'app',
  pattern: '.js,.ts,.vue',
  dependsOn: ['Frontend Formatting Check'],
  pluginName: 'eslint-oxlint'
});

// 7. Frontend Vue/JS Tests
e.register({
  name: 'Frontend Vue/JS Tests',
  dir: 'app',
  pattern: '.js,.ts,.vue',
  dependsOn: ['Frontend Linting'],
  pluginName: 'vitest-coverage'
});

e.run(e.FileResolver, process.argv.slice(2));
