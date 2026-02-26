import {
  deleteDiscussionsByPostId,
  getDiscussionsByPostId,
  getDiscussionsByPostIdComments,
  postDiscussionsByPostIdComments,
  postDiscussionsByPostIdVote,
  postDiscussionsCommentsByCommentIdVote,
  putDiscussionsByPostId,
  putDiscussionsCommentsByCommentId,
  type DeleteDiscussionsByPostIdData,
  type GetDiscussionsByPostIdCommentsData,
  type GetDiscussionsByPostIdData,
  type PostDiscussionsByPostIdCommentsData,
  type PostDiscussionsByPostIdVoteData,
  type PostDiscussionsCommentsByCommentIdVoteData,
  type PutDiscussionsByPostIdData,
  type PutDiscussionsCommentsByCommentIdData,
} from '@lib/api';
import { unwrapId, type CommentId, type PostId } from '@lib/types/domain';

const toPostPath = (postId: PostId): GetDiscussionsByPostIdData['path'] =>
  ({ postId: unwrapId(postId) } as unknown as GetDiscussionsByPostIdData['path']);

const toCommentPath = (commentId: CommentId): PutDiscussionsCommentsByCommentIdData['path'] =>
  ({ commentId: unwrapId(commentId) } as unknown as PutDiscussionsCommentsByCommentIdData['path']);

export const getDiscussionById = (postId: PostId) =>
  getDiscussionsByPostId({ path: toPostPath(postId) });

export const getDiscussionComments = (postId: PostId) =>
  getDiscussionsByPostIdComments({ path: toPostPath(postId) as unknown as GetDiscussionsByPostIdCommentsData['path'] });

export const voteDiscussionPost = (postId: PostId, vote: -1 | 0 | 1) =>
  postDiscussionsByPostIdVote({
    path: toPostPath(postId) as unknown as PostDiscussionsByPostIdVoteData['path'],
    body: { vote },
  });

export const addDiscussionComment = (postId: PostId, body: { parent_id?: string; text: string }) =>
  postDiscussionsByPostIdComments({
    path: toPostPath(postId) as unknown as PostDiscussionsByPostIdCommentsData['path'],
    body,
  });

export const updateDiscussionPost = (
  postId: PostId,
  body: Partial<PutDiscussionsByPostIdData['body']> & { pinned?: boolean },
) =>
  putDiscussionsByPostId({
    path: toPostPath(postId) as unknown as PutDiscussionsByPostIdData['path'],
    body: body as PutDiscussionsByPostIdData['body'],
  });

export const updateDiscussionComment = (commentId: CommentId, body: { text: string }) =>
  putDiscussionsCommentsByCommentId({
    path: toCommentPath(commentId),
    body,
  });

export const voteDiscussionComment = (commentId: CommentId, vote: -1 | 0 | 1) =>
  postDiscussionsCommentsByCommentIdVote({
    path: toCommentPath(commentId) as unknown as PostDiscussionsCommentsByCommentIdVoteData['path'],
    body: { vote },
  });

export const deleteDiscussionById = (postId: PostId) =>
  deleteDiscussionsByPostId({
    path: toPostPath(postId) as unknown as DeleteDiscussionsByPostIdData['path'],
  });
