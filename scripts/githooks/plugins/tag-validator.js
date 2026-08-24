const { spawnSync } = require('child_process');

module.exports = (engine) => {
  engine.registerPlugin('tag-validator', (e, dir, files) => {
    // Intercept stdinLines from the engine
    const lines = e.stdinLines || [];
    const tagsPushed = [];

    // Parse stdin lines: <local-ref> <local-sha> <remote-ref> <remote-sha>
    for (const line of lines) {
      const parts = line.trim().split(/\s+/);
      if (parts.length < 4) continue;
      const [localRef, localSha, remoteRef, remoteSha] = parts;

      // Check if it is a tag being pushed
      if (remoteRef.startsWith('refs/tags/')) {
        const tagName = remoteRef.replace('refs/tags/', '');
        tagsPushed.push({ name: tagName, sha: localSha });
      }
    }

    if (tagsPushed.length === 0) {
      return; // No tags being pushed
    }

    // Retrieve all existing tags from Git
    const tagsResult = spawnSync('git', ['tag'], { cwd: dir });
    let existingTags = [];
    if (tagsResult.status === 0) {
      existingTags = tagsResult.stdout.toString().trim().split('\n').map(t => t.trim()).filter(Boolean);
    }

    for (const tag of tagsPushed) {
      e.logln(`🏷️ Validating pushed tag: "${tag.name}" (commit ${tag.sha.substring(0, 7)})...`);

      // 1. Tag must start with "v" prefix
      if (!tag.name.startsWith('v')) {
        throw new Error(`Tag name "${tag.name}" is invalid. It must be prefixed with "v" (e.g. "v1.0.0").`);
      }

      // 2. Tag must be on the main branch
      // Verify localSha is an ancestor of main
      const ancestorResult = spawnSync('git', ['merge-base', '--is-ancestor', tag.sha, 'main'], { cwd: dir });
      if (ancestorResult.status !== 0) {
        throw new Error(`Tag "${tag.name}" (commit ${tag.sha.substring(0, 7)}) must be on the "main" branch.`);
      }

      // 3. Parse semver and check increment relative to existing tags
      const newParsed = parseSemver(tag.name);
      if (!newParsed) {
        throw new Error(`Tag "${tag.name}" does not match semver format "vX.Y.Z" (e.g., "v1.2.3").`);
      }

      // If the tag already exists in local git tags with a different commit sha, reject duplication
      const duplicateCheck = spawnSync('git', ['rev-parse', tag.name], { cwd: dir });
      if (duplicateCheck.status === 0) {
        const existingSha = duplicateCheck.stdout.toString().trim();
        if (existingSha !== tag.sha) {
          throw new Error(`Tag "${tag.name}" already exists locally on a different commit (${existingSha.substring(0, 7)}).`);
        }
      }

      // Check highest existing version
      let highest = null;
      for (const existingTag of existingTags) {
        // Skip comparing against the tag being pushed if it already exists locally on the same SHA
        if (existingTag === tag.name) continue;

        const parsed = parseSemver(existingTag);
        if (!parsed) continue;
        if (!highest || compareSemver(parsed, highest) > 0) {
          highest = parsed;
        }
      }

      if (highest && compareSemver(newParsed, highest) <= 0) {
        throw new Error(`Tag "${tag.name}" is invalid. The version must be strictly greater than the highest existing tag "v${highest.major}.${highest.minor}.${highest.patch}".`);
      }

      e.logln(`✅ Tag "${tag.name}" validated successfully!`);
    }
  });
};

function parseSemver(tag) {
  const match = tag.match(/^v(\d+)\.(\d+)\.(\d+)(?:-([\w.-]+))?$/);
  if (!match) return null;
  return {
    major: parseInt(match[1], 10),
    minor: parseInt(match[2], 10),
    patch: parseInt(match[3], 10),
    prerelease: match[4] || ''
  };
}

function compareSemver(a, b) {
  if (a.major !== b.major) return a.major - b.major;
  if (a.minor !== b.minor) return a.minor - b.minor;
  if (a.patch !== b.patch) return a.patch - b.patch;
  return 0;
}
