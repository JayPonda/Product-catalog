const path = require('path');
const { Engine, getBranchChangedFiles } = require('./engine');

const e = new Engine(path.join(__dirname, 'plugins'));
e.FileResolver = getBranchChangedFiles;

// 1. Go Backend Tests
e.register({
  name: 'Go Backend Tests',
  dir: 'server',
  pattern: '.go',
  pluginName: 'go-test'
});

// 2. Frontend Vue/JS Tests
e.register({
  name: 'Frontend Vue/JS Tests',
  dir: 'app',
  pattern: '.js,.ts,.vue',
  pluginName: 'vitest'
});

e.run(e.FileResolver, process.argv.slice(2));
