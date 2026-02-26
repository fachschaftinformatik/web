import { useCallback, useEffect, useState } from 'react';
import { getDiscussions, getPrograms, type DtoProgramResponse } from '@lib/api';
import { deleteDiscussionById, updateDiscussionPost, voteDiscussionPost } from '@lib/adapters/discussions';
import type { Post, Vote } from '@components/discussions/components';
import { POSTS_PER_PAGE } from '@components/discussions/components';
import { toPostId } from '@lib/types/domain';
import type { SessionUser } from '@lib/types/session';
import { applyVote, normalizeVote, resolveNextVote } from '@lib/discussions/vote';

export type DiscussionSort = 'new' | 'votes' | 'relevant';

type UseDiscussionListOptions = {
  user: SessionUser | null;
};

export const useDiscussionList = ({ user }: UseDiscussionListOptions) => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [q, setQ] = useState('');
  const [sort, setSort] = useState<DiscussionSort>('new');
  const [activeProgramFilters, setActiveProgramFilters] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [apiPrograms, setApiPrograms] = useState<DtoProgramResponse[]>([]);

  const isAdmin = user?.role === 'admin';
  const canModerate = isAdmin || user?.role === 'editor';

  const fetchPosts = useCallback(async () => {
    setLoading(true);
    try {
      const offset = (page - 1) * POSTS_PER_PAGE;
      const selectedProgramId = activeProgramFilters.length > 0
        ? apiPrograms.find((program) => program.name === activeProgramFilters[0])?.id
        : undefined;

      const { data, response } = await getDiscussions({
        query: {
          limit: POSTS_PER_PAGE,
          offset,
          query: q.trim() || undefined,
          sort: sort === 'votes' ? 'votes' : undefined,
          program_id: selectedProgramId,
        },
      });

      if (data) {
        const parsed: Post[] = data.map((post) => ({
          ...post,
          comments: [],
        }));
        setPosts(parsed);

        const total = Number.parseInt(response.headers.get('X-Total-Count') || '0', 10);
        setTotalCount(total);
      }
    } catch (error) {
      console.error('Fetch error:', error);
    } finally {
      setLoading(false);
    }
  }, [activeProgramFilters, apiPrograms, page, q, sort]);

  useEffect(() => {
    const fetchPrograms = async () => {
      try {
        const { data } = await getPrograms();
        if (data) {
          setApiPrograms(data);
        }
      } catch (error) {
        console.error('Failed to fetch programs:', error);
      }
    };

    void fetchPrograms();
  }, []);

  useEffect(() => {
    void fetchPosts();
  }, [fetchPosts]);

  useEffect(() => {
    setPage(1);
  }, [activeProgramFilters, q, sort]);

  const votePost = useCallback(async (id: string, newVote: Vote) => {
    if (!user) {
      return;
    }

    const postId = toPostId(id);
    if (!postId) {
      return;
    }

    const target = posts.find((post) => String(post.id) === id);
    if (!target) {
      return;
    }

    const currentVote = normalizeVote(target.user_vote);
    const nextVote = resolveNextVote(currentVote, newVote);

    try {
      await voteDiscussionPost(postId, nextVote);
      setPosts((prev) => prev.map((post) => {
        if (String(post.id) !== id) {
          return post;
        }

        return {
          ...post,
          user_vote: nextVote,
          votes: applyVote(post.votes, normalizeVote(post.user_vote), nextVote),
        };
      }));
    } catch (error) {
      console.error(error);
    }
  }, [posts, user]);

  const togglePin = useCallback(async (id: string) => {
    if (!canModerate) {
      return;
    }

    const post = posts.find((item) => String(item.id) === id);
    if (!post) {
      return;
    }

    const postId = toPostId(id);
    if (!postId) {
      return;
    }

    try {
      await updateDiscussionPost(postId, { pinned: !post.pinned });
      setPosts((prev) => prev.map((item) => String(item.id) === id
        ? { ...item, pinned: post.pinned ? 0 : 1 }
        : item));
    } catch (error) {
      console.error(error);
    }
  }, [canModerate, posts]);

  const deletePost = useCallback(async (id: string) => {
    const target = posts.find((post) => String(post.id) === id);
    const canDelete = canModerate || String(user?.id) === String(target?.user_id);
    if (!canDelete) {
      return;
    }

    const postId = toPostId(id);
    if (!postId) {
      return;
    }

    try {
      await deleteDiscussionById(postId);
      setPosts((prev) => prev.filter((post) => String(post.id) !== id));
    } catch (error) {
      console.error(error);
    }
  }, [canModerate, posts, user?.id]);

  const pageCount = Math.ceil(totalCount / POSTS_PER_PAGE);

  const setSortOrder = useCallback((value: string) => {
    if (value === 'new' || value === 'votes' || value === 'relevant') {
      setSort(value);
    }
  }, []);

  return {
    posts,
    totalCount,
    pageCount,
    loading,
    q,
    sort,
    activeProgramFilters,
    page,
    apiPrograms,
    isAdmin,
    canModerate,
    setQ,
    setSortOrder,
    setActiveProgramFilters,
    setPage,
    votePost,
    togglePin,
    deletePost,
    normalizeVote,
  };
};
