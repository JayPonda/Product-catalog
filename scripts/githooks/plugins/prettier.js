module.exports = (engine) => {
  engine.registerPlugin('prettier', (e, dir, files) => {
    const success = e.runCommand(dir, 'node', ['./node_modules/.bin/prettier', '--check', ...files]);
    if (!success) {
      throw new Error("frontend formatting check failed; please run 'pnpm --dir app run format'");
    }
  });
};
