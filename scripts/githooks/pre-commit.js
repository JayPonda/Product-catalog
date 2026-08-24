const path = require('path');
const { Engine, getStagedFiles } = require('./engine');

const e = new Engine(path.join(__dirname, 'plugins'));
e.FileResolver = getStagedFiles;

// 1. Go Formatting Check
e.register({
  name: 'Go Formatting Check',
  dir: 'server',
  pattern: '.go',
  pluginName: 'gofmt'
});

// 2. Go Linting
e.register({
  name: 'Go Linting',
  dir: 'server',
  pattern: '.go',
  dependsOn: ['Go Formatting Check'],
  pluginName: 'golangci-lint'
});

// 3. Frontend Formatting Check
e.register({
  name: 'Frontend Formatting Check',
  dir: 'app',
  pattern: '.js,.ts,.vue,.css,.json',
  pluginName: 'prettier'
});

// 4. Frontend Linting
e.register({
  name: 'Frontend Linting',
  dir: 'app',
  pattern: '.js,.ts,.vue',
  dependsOn: ['Frontend Formatting Check'],
  pluginName: 'eslint-oxlint'
});

e.run(e.FileResolver, process.argv.slice(2));
