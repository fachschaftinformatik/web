module.exports = {
  branches: ['main'],
  tagFormat: 'v${version}',
  plugins: [
    ['@semantic-release/commit-analyzer', {
      preset: 'conventionalcommits',
      releaseRules: [
        { breaking: true, release: 'major' },
        { type: 'chore', scope: 'deps', release: 'patch' },
      ],
    }],
    ['@semantic-release/github', {
      releaseBodyTemplate: '<%= "" %>',
      successComment: false,
      failComment: false,
      failTitle: false,
      releasedLabels: false,
    }],
    ['@semantic-release/exec', {
      execCwd: '../..',
      prepareCmd: 'RELEASE_TAG=${nextRelease.gitTag} docker buildx bake --file bake.hcl --push',
      successCmd: 'printf "released=true\\ntag=%s\\n" "${nextRelease.gitTag}" >> "$GITHUB_OUTPUT"',
    }],
  ],
};
