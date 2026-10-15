import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, relative } from 'node:path';
import { createRequire } from 'node:module';
import { test } from 'node:test';

const require = createRequire(import.meta.url);
const { getRootDirs } = require('@next/eslint-plugin-next/dist/utils/get-root-dirs');

test('Next ESLint root-directory discovery works with the tinyglobby override', () => {
  const cwd = mkdtempSync(join(tmpdir(), 'eslint-root-dirs-'));
  try {
    for (const directory of ['apps/web', 'apps/admin', 'packages/shared']) {
      mkdirSync(join(cwd, directory), { recursive: true });
    }
    const discover = rootDir => getRootDirs({
      cwd,
      settings: { next: { rootDir } },
    }).map(path => relative(cwd, path).replaceAll('\\', '/')).sort();
    const pattern = suffix => `${cwd.replaceAll('\\', '/')}/${suffix}`;

    assert.deepEqual(getRootDirs({ cwd, settings: {} }), [cwd]);
    assert.deepEqual(discover(pattern('apps/*')), ['apps/admin', 'apps/web']);
    assert.deepEqual(discover(pattern('apps/{admin,web}')), ['apps/admin', 'apps/web']);
    assert.deepEqual(discover([pattern('apps/web'), pattern('packages/*')]), [
      'apps/web', 'packages/shared',
    ]);
    assert.deepEqual(discover(pattern('missing/*')), []);
    assert.deepEqual(discover(pattern('apps/*').replaceAll('/', '\\')), [
      'apps/admin', 'apps/web',
    ]);
  } finally {
    rmSync(cwd, { recursive: true, force: true });
  }
});
