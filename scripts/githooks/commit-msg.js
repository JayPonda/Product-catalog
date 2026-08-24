const path = require('path');
const { Engine } = require('./engine');

const e = new Engine(path.join(__dirname, 'plugins'));

const commitMsgFile = process.argv[2];

e.FileResolver = () => {
  if (!commitMsgFile) {
    throw new Error('Commit message file path not provided.');
  }
  return [commitMsgFile];
};

e.register({
  name: 'Commit Message Format Validation',
  pluginName: 'commit-msg-format'
});

e.run(e.FileResolver);
