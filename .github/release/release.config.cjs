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
      prepareCmd: 'docker buildx build --platform linux/amd64 --sbom=true --build-arg VERSION=${nextRelease.gitTag} --label org.opencontainers.image.version=${nextRelease.gitTag} --tag ghcr.io/fachschaftinformatik/web:${nextRelease.gitTag} --cache-from type=gha,scope=app-amd64 --cache-to type=gha,mode=max,scope=app-amd64 --push .',
      successCmd: 'printf "released=true\\ntag=%s\\n" "${nextRelease.gitTag}" >> "$GITHUB_OUTPUT"',
    }],
  ],
};
