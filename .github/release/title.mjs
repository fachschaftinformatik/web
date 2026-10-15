export function validateTitle(title) {
  return typeof title === 'string' &&
    /^(feat|fix|docs|refactor|test|perf|build|ci|chore|style|revert)(\([a-z0-9][a-z0-9./_-]*\))?!?: [^\r\n]+$/.test(title) &&
    title === title.toLowerCase();
}

if (process.env.PR_TITLE !== undefined && !validateTitle(process.env.PR_TITLE)) {
  console.error('Use a lowercase Conventional Commit title, e.g. fix(deps): bump next from 16.4.1 to 16.4.2');
  process.exitCode = 1;
}
