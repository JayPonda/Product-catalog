const path = require('path');
const { Engine, getStagedFiles } = require('./engine');

const e = new Engine(path.join(__dirname, 'plugins'));
e.FileResolver = getStagedFiles;

// 1. No Direct Commits to Main
e.register({
  name: 'No Direct Commits to Main',
  pattern: '',
  pluginName: 'no-direct-main'
});

// 2. Branch Name Check
e.register({
  name: 'Branch Name Check',
  pattern: '',
  dependsOn: ['No Direct Commits to Main'],
  pluginName: 'branch-name'
});

// 3. Go Formatting Check
e.register({
  name: 'Go Formatting Check',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Branch Name Check'],
  pluginName: 'gofmt'
});

// 4. Go Linting
e.register({
  name: 'Go Linting',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Go Formatting Check'],
  pluginName: 'golangci-lint'
});

// 5. Frontend Formatting Check
e.register({
  name: 'Frontend Formatting Check',
  dir: 'app',
  pattern: '.js,.ts,.vue,.css,.json',
  dependsOn: ['Branch Name Check'],
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

e.run(e.FileResolver, process.argv.slice(2));
