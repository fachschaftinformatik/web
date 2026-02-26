import type { Vote } from '@components/discussions/components';

export const normalizeVote = (value: number | null | undefined): Vote =>
  value === 1 || value === -1 ? value : 0;

export const resolveNextVote = (currentVote: Vote, requestedVote: Vote): Vote =>
  currentVote === requestedVote ? 0 : requestedVote;

export const applyVote = (score: number | null | undefined, currentVote: Vote, nextVote: Vote): number =>
  Number(score || 0) - currentVote + nextVote;
