import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { test } from 'node:test';
import { analyzeCommits } from '@semantic-release/commit-analyzer';
import { validateTitle } from '../title.mjs';
import config from '../release.config.cjs';

const require = createRequire(import.meta.url);
const braces = require('braces');
const analyzerOptions = config.plugins[0][1];

for (const [messages, expected] of [
  [['fix: correct navigation'], 'patch'],
  [['perf: cache gallery resources'], 'patch'],
  [['feat: add event albums'], 'minor'],
  [['feat!: replace the authentication api'], 'major'],
  [['chore(deps): bump next from 16.4.1 to 16.4.2'], 'patch'],
  [['chore(deps)!: bump a breaking dependency'], 'major'],
  [['fix: update api\n\nBREAKING CHANGE: remove the old endpoint'], 'major'],
  [['ci: configure semantic-release', 'docs: explain releases'], null],
  [['chore(ci): bump semantic-release from 25.0.8 to 25.0.9'], null],
  [['fix: correct navigation', 'feat: add albums'], 'minor'],
]) {
  test(`release classification: ${messages.join('; ')}`, async () => {
    const result = await analyzeCommits(analyzerOptions, {
      cwd: new URL('..', import.meta.url).pathname,
      commits: messages.map(message => ({ message, hash: 'test' })),
      logger: { log() {} },
    });
    assert.equal(result, expected);
  });
}

test('PR titles accept the supported release conventions', () => {
  for (const title of ['feat!: replace auth', 'chore(deps): bump next from 1 to 2', 'ci: add release checks']) {
    assert.equal(validateTitle(title), true, title);
  }
  for (const title of ['Update packages', 'fix: Uppercase title', 'fix: bad\nsecond line', 'fix: ', 'fix:']) {
    assert.equal(validateTitle(title), false, title);
  }
});

test('patched braces rejects deeply nested patterns before recursive processing', () => {
  const patterns = [
    '{'.repeat(2000) + 'a,b' + '}'.repeat(2000),
    '('.repeat(2000) + 'a' + ')'.repeat(2000),
    '{('.repeat(1000) + 'a,b' + ')}'.repeat(1000),
    '{'.repeat(2000) + 'a',
  ];
  for (const pattern of patterns) {
    for (const method of [braces, braces.parse, braces.compile, braces.expand, braces.stringify]) {
      assert.throws(() => method(pattern), /exceeds max depth/);
    }
  }
  assert.deepEqual(braces.expand('release/{api,website}/v{1..2}'), [
    'release/api/v1', 'release/api/v2', 'release/website/v1', 'release/website/v2',
  ]);
});

test('patched braces bounds recursion for directly supplied ASTs', () => {
  for (const method of [braces.compile, braces.expand, braces.stringify]) {
    let ast = { type: 'text', value: 'a' };
    for (let i = 0; i < 2000; i++) ast = { type: 'brace', nodes: [ast] };
    assert.throws(() => method({ type: 'root', nodes: [ast] }), /exceeds max depth/);
  }
});
