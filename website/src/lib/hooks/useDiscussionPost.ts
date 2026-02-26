import { useCallback, useEffect, useState } from 'react';
import type { Comment, Post, Vote } from '@components/discussions/components';
import type { SessionUser } from '@lib/types/session';
import {
  addDiscussionComment,
  deleteDiscussionById,
  getDiscussionById,
  getDiscussionComments,
  updateDiscussionComment,
  voteDiscussionComment,
  voteDiscussionPost,
} from '@lib/adapters/discussions';
import { toCommentId, toPostId } from '@lib/types/domain';
import { applyVote, normalizeVote, resolveNextVote } from '@lib/discussions/vote';

type UseDiscussionPostOptions = {
  postIdParam: string;
  user: SessionUser | null;
  accentColor: string;
};

export const useDiscussionPost = ({ postIdParam, user, accentColor }: UseDiscussionPostOptions) => {
  const [post, setPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const postId = toPostId(postIdParam);
  const isAdmin = user?.role === 'admin' || user?.role === 'editor';

  useEffect(() => {
    if (!postId) {
      setError('Ungültige Beitrags-ID.');
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    Promise.all([getDiscussionById(postId), getDiscussionComments(postId)])
      .then(([{ data: postData, error: postError }, { data: commentsData }]) => {
        if (postError || !postData) {
          setError('Beitrag nicht gefunden oder Fehler beim Laden.');
          return;
        }

        const comments: Comment[] = commentsData || [];
        setPost({ ...postData, comments });

        const hash = window.location.hash.replace('#', '');
        if (!hash) {
          return;
        }

        setTimeout(() => {
          const element = document.getElementById(hash);
          if (!element) {
            return;
          }

          element.scrollIntoView({ behavior: 'smooth', block: 'center' });
          const textBox = element.querySelector<HTMLElement>('.MuiPaper-root');
          if (!textBox) {
            return;
          }

          textBox.style.transition = 'all 0.5s ease';
          const originalBorder = textBox.style.borderColor;
          const originalShadow = textBox.style.boxShadow;

          textBox.style.borderColor = accentColor;
          textBox.style.boxShadow = `0 0 0 2px ${accentColor}33`;

          setTimeout(() => {
            textBox.style.borderColor = originalBorder;
            textBox.style.boxShadow = originalShadow;
          }, 3000);
        }, 800);
      })
      .catch((fetchError) => {
        console.error(fetchError);
        setError('Ein unerwarteter Fehler ist aufgetreten.');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [accentColor, postId]);

  const votePost = useCallback(async (vote: Vote) => {
    if (!post || !user || !postId) {
      return;
    }

    const currentVote = normalizeVote(post.user_vote);
    const nextVote = resolveNextVote(currentVote, vote);

    setPost((prev) => prev
      ? { ...prev, user_vote: nextVote, votes: applyVote(prev.votes, normalizeVote(prev.user_vote), nextVote) }
      : null);

    try {
      await voteDiscussionPost(postId, nextVote);
    } catch (voteError) {
      console.error('Vote failed', voteError);
    }
  }, [post, postId, user]);

  const addComment = useCallback(async (parentId: string | null, text: string) => {
    if (!postId || !user) {
      return;
    }

    const { data } = await addDiscussionComment(postId, {
      parent_id: parentId ? String(parentId) : undefined,
      text,
    });

    if (!data) {
      return;
    }

    const newComment: Comment = {
      ...data,
      user_name: user.name || user.email || 'Anonym',
      user_avatar_url: user.avatar_url || '',
      user_id: user.id,
    };

    setPost((prev) => prev
      ? {
        ...prev,
        comments: [...prev.comments, newComment],
        comment_count: (prev.comment_count || 0) + 1,
      }
      : null);
  }, [postId, user]);

  const editComment = useCallback(async (commentIdValue: string, text: string) => {
    if (!user) {
      return;
    }

    const commentId = toCommentId(commentIdValue);
    if (!commentId) {
      return;
    }

    const { data } = await updateDiscussionComment(commentId, { text });
    if (!data) {
      return;
    }

    setPost((prev) => {
      if (!prev) {
        return null;
      }

      const updatedComments = prev.comments.map((comment) => (
        String(comment.id) === commentIdValue
          ? { ...comment, text: data.text || '', updated_at: data.updated_at || '' }
          : comment
      ));

      return { ...prev, comments: updatedComments };
    });
  }, [user]);

  const voteComment = useCallback(async (commentIdValue: string, vote: Vote) => {
    if (!post || !user) {
      return;
    }

    const commentId = toCommentId(commentIdValue);
    if (!commentId) {
      return;
    }

    try {
      await voteDiscussionComment(commentId, vote);

      setPost((prev) => {
        if (!prev) {
          return null;
        }

        const updatedComments = prev.comments.map((comment) => {
          if (String(comment.id) !== commentIdValue) {
            return comment;
          }

          const oldUserVote = normalizeVote(comment.user_vote);
          const newVotes = applyVote(comment.votes, oldUserVote, vote);

          return { ...comment, user_vote: vote, votes: newVotes };
        });

        return { ...prev, comments: updatedComments };
      });
    } catch (voteError) {
      console.error('Comment vote failed', voteError);
    }
  }, [post, user]);

  const deletePost = useCallback(async () => {
    if (!postId || (!isAdmin && String(post?.user_id) !== String(user?.id))) {
      return false;
    }

    await deleteDiscussionById(postId);
    return true;
  }, [isAdmin, post?.user_id, postId, user?.id]);

  return {
    post,
    loading,
    error,
    isAdmin,
    netVotes: Number(post?.votes) || 0,
    userVote: normalizeVote(post?.user_vote),
    votePost,
    addComment,
    editComment,
    voteComment,
    deletePost,
  };
};
