import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import IconButton from "@mui/material/IconButton";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Tooltip from "@mui/material/Tooltip";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";
import Paper from "@mui/material/Paper";
import { useTheme } from "@mui/material/styles";
import DelIcon from "@mui/icons-material/DeleteOutline";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ShareIcon from "@mui/icons-material/Share";

import Page from "@components/Page";
import Back from "@components/Back";
import { useAuth } from "@lib/auth";
import {
  getDiscussionsByPostId, getDiscussionsByPostIdComments,
  postDiscussionsByPostIdComments, postDiscussionsByPostIdVote,
  postDiscussionsCommentsByCommentIdVote, deleteDiscussionsByPostId,
  putDiscussionsCommentsByCommentId
} from "@lib/api";
import { Post, Comment, CommentsSection, isoToShort, Vote } from "./DiscussionComponents";
import { getAvatarUrl } from "@lib/images";

export default function DiscussionDetail() {
  const { postId } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const theme = useTheme();
  const [post, setPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!postId) return;
    Promise.all([
      getDiscussionsByPostId({ path: { id: postId } }),
      getDiscussionsByPostIdComments({ path: { id: postId } })
    ]).then(([{ data: p }, { data: c }]) => {
      if (p) setPost({ ...p, comments: (c || []) as Comment[] });
      setLoading(false);
    });
  }, [postId]);

  const vote = async (v: Vote) => {
    if (!post || !user || !postId) return;
    const target = v === post.user_vote ? 0 : v;
    await postDiscussionsByPostIdVote({ path: { id: postId }, body: { vote: target as never } });
    setPost({ ...post, user_vote: target, votes: Number(post.votes) - (post.user_vote || 0) + target });
  };

  const addComment = async (pid: string | null, text: string) => {
    if (!postId || !user) return;
    const { data } = await postDiscussionsByPostIdComments({ path: { id: postId }, body: { parent_id: pid || undefined, text } });
    if (data && post) setPost({ ...post, comments: [...post.comments, { ...data, user_name: user.name || 'Anonym', user_avatar_url: user.avatar_url || '', user_id: user.id } as Comment], comment_count: (post.comment_count || 0) + 1 });
  };

  if (loading) return <Page title="Laden..."><CircularProgress /></Page>;
  if (!post) return <Page title="Fehler"><Alert severity="error">Nicht gefunden</Alert></Page>;

  return (
    <Page title={post.title || 'Diskussion'} maxWidth="md">
      <Back to="/discussions" />
      <Paper variant="outlined" sx={{ p: { xs: 2, md: 4 }, borderRadius: 3 }}>
        <Stack direction="row" spacing={3}>
          <Stack alignItems="center" spacing={0.5} sx={{ bgcolor: 'action.hover', borderRadius: 2, p: 0.5 }}>
            <IconButton size="small" onClick={() => vote(1)} color={post.user_vote === 1 ? "primary" : "default"}><ThumbUpOutlined fontSize="small" /></IconButton>
            <Typography variant="subtitle2" fontWeight={700}>{post.votes}</Typography>
            <IconButton size="small" onClick={() => vote(-1)} color={post.user_vote === -1 ? "primary" : "default"}><ThumbDownOutlined fontSize="small" /></IconButton>
          </Stack>
          <Box sx={{ flex: 1 }}>
            <Stack direction="row" justifyContent="space-between">
              <Typography variant="h4" fontWeight={700}>{post.title}</Typography>
              <Stack direction="row" spacing={1}>
                <IconButton onClick={() => { navigator.clipboard.writeText(window.location.href); setCopied(true); setTimeout(() => setCopied(false), 2000); }}><Tooltip title={copied ? "Kopiert!" : "Teilen"}><ShareIcon fontSize="small" /></Tooltip></IconButton>
                {(user?.role === 'admin' || user?.id === post.user_id) && <IconButton color="error" onClick={async () => { if (confirm("Löschen?")) { await deleteDiscussionsByPostId({ path: { id: postId! } }); navigate("/discussions"); } }}><DelIcon fontSize="small" /></IconButton>}
              </Stack>
            </Stack>
            <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 3 }}>
              <Avatar src={getAvatarUrl(post.user_avatar_url)} sx={{ width: 24, height: 24 }} />
              <Typography variant="body2" color="text.secondary">von <b>{post.user_name}</b> · {isoToShort(post.created_at || '')}</Typography>
            </Stack>
            <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap', mb: 3 }}>{post.body}</Typography>
            <Stack direction="row" spacing={1} sx={{ mb: 3 }}>{post.tags?.map(t => <Chip key={t} label={t} size="small" variant="outlined" />)}</Stack>
            <Divider sx={{ mb: 3 }} />
            <CommentsSection comments={post.comments} onAdd={addComment} currentUser={user} appearance={{ surface: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : '#f5f5f5', border: theme.palette.divider, accent: theme.palette.primary.main, textSecondary: theme.palette.text.secondary }}
              onEdit={async (id, text) => { await putDiscussionsCommentsByCommentId({ path: { id }, body: { text } }); if (post) setPost({ ...post, comments: post.comments.map(c => String(c.id) === id ? { ...c, text } : c) }); }}
              onVote={async (id, v) => { await postDiscussionsCommentsByCommentIdVote({ path: { id }, body: { vote: v } }); if (post) setPost({ ...post, comments: post.comments.map(c => String(c.id) === id ? { ...c, user_vote: v, votes: (c.votes || 0) - (c.user_vote || 0) + v } : c) }); }}
            />
          </Box>
        </Stack>
      </Paper>
    </Page>
  );
}
