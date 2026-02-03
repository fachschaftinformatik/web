import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import MenuItem from "@mui/material/MenuItem";
import Autocomplete from "@mui/material/Autocomplete";
import Avatar from "@mui/material/Avatar";
import Skeleton from "@mui/material/Skeleton";
import { alpha, useTheme } from "@mui/material/styles";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import PinIcon from "@mui/icons-material/PushPinOutlined";
import DelIcon from "@mui/icons-material/DeleteOutline";

import Page from "@components/Page";
import Pagination from "@components/Pagination";
import { useAuth } from "@lib/auth";
import { getDiscussions, getPrograms, postDiscussionsByPostIdVote, deleteDiscussionsByPostId } from "@lib/api";
import type { DtoDiscussionPostResponse as Post, DtoProgramResponse as Prog } from "@lib/api";
import { getAvatarUrl } from "@lib/images";
import { POSTS_PER_PAGE, Vote } from "./DiscussionComponents";

const PAGE_SIZE = POSTS_PER_PAGE;

export default function Discussions() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const theme = useTheme();
  const [posts, setPosts] = useState<Post[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [progs, setProgs] = useState<Prog[]>([]);
  const [q, setQ] = useState("");
  const [sort, setSort] = useState("new");
  const [selProgs, setSelProgs] = useState<string[]>([]);
  const [page, setPage] = useState(1);

  useEffect(() => { getPrograms().then(({ data }) => data && setProgs(data)); }, []);

  const fetch = useCallback(async () => {
    void Promise.resolve().then(() => setLoading(true));
    const { data, response } = await getDiscussions({ query: { 
      limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE, query: q || undefined, 
      sort: sort === 'votes' ? 'votes' : undefined,
      program_id: selProgs.length ? progs.find(p => p.name === selProgs[0])?.id : undefined
    }});
    if (data) { setPosts(data); setCount(parseInt(response.headers.get("X-Total-Count") || "0")); }
    setLoading(false);
  }, [page, q, sort, selProgs, progs]);

  useEffect(() => {
    void Promise.resolve().then(() => fetch());
  }, [fetch]);

  const vote = async (id: string, v: number) => {
    if (!user) return;
    await postDiscussionsByPostIdVote({ path: { id }, body: { vote: v as Vote } });
    setPosts(prev => prev.map(p => String(p.id) === id ? { ...p, user_vote: v, votes: (Number(p.votes)) - (p.user_vote || 0) + v } : p));
  };

  const del = async (id: string) => {
    if (confirm("Löschen?")) {
      await deleteDiscussionsByPostId({ path: { id } });
      setPosts(prev => prev.filter(p => String(p.id) !== id));
    }
  };

  const PostItem = ({ p }: { p: Post }) => (
    <Paper onClick={() => navigate(`/d/${p.id}`)} variant="outlined" sx={{ p: 2, borderRadius: 3, cursor: 'pointer', '&:hover': { borderColor: 'primary.main' }, bgcolor: p.type === 'news' ? alpha(theme.palette.primary.main, 0.05) : 'background.paper' }}>
      <Stack direction="row" spacing={2}>
        <Stack alignItems="center">
          <IconButton size="small" onClick={e => { e.stopPropagation(); vote(String(p.id), p.user_vote === 1 ? 0 : 1); }} color={p.user_vote === 1 ? "primary" : "default"}><ThumbUpOutlined fontSize="small" /></IconButton>
          <Typography variant="subtitle2">{p.votes}</Typography>
          <IconButton size="small" onClick={e => { e.stopPropagation(); vote(String(p.id), p.user_vote === -1 ? 0 : -1); }} color={p.user_vote === -1 ? "primary" : "default"}><ThumbDownOutlined fontSize="small" /></IconButton>
        </Stack>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
            <Typography variant="h6" fontWeight={700} sx={{ 
              display: '-webkit-box',
              WebkitLineClamp: 1,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              wordBreak: 'break-word'
            }}>
              {p.title} {p.pinned === 1 && <PinIcon fontSize="small" color="primary" sx={{ verticalAlign: 'middle', ml: 0.5 }} />}
            </Typography>
            { (user?.role === 'admin' || user?.id === p.user_id) && <IconButton size="small" onClick={e => { e.stopPropagation(); del(String(p.id!)); }} sx={{ ml: 1, flexShrink: 0 }}><DelIcon fontSize="small" /></IconButton> }
          </Stack>
          <Typography variant="body2" color="text.secondary" sx={{ 
            mb: 1,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            wordBreak: 'break-word'
          }}>
            {p.body}
          </Typography>
          <Stack direction="row" spacing={1} alignItems="center">
            <Avatar src={getAvatarUrl(p.user_avatar_url)} sx={{ width: 20, height: 20 }} />
            <Typography variant="caption">von <b>{p.user_name}</b> · {p.comment_count} {p.comment_count === 1 ? 'Kommentar' : 'Kommentare'}</Typography>
          </Stack>
        </Box>
      </Stack>
    </Paper>
  );

  return (
    <Page title="Beiträge" description="Hier könnt ihr euch mit eurem Kommilitonen direkt austauschen. Außerdem veröffentlicht die FSV hier anstehende Events und weitere Infos.">
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => navigate("/d/new")}>Neu</Button>
      </Box>

      <Paper variant="outlined" sx={{ p: 2, mb: 3 }}>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
          <TextField fullWidth placeholder="Suchen..." value={q} onChange={e => setQ(e.target.value)} size="small" slotProps={{ input: { startAdornment: <SearchIcon color="primary" /> } }} />
          <Autocomplete multiple options={progs.map(p => p.name || '')} value={selProgs} onChange={(_, v) => setSelProgs(v)} renderInput={p => <TextField {...p} label="Studiengang" size="small" />} sx={{ flex: 1 }} />
          <TextField select label="Sortierung" value={sort} onChange={e => setSort(e.target.value)} size="small" sx={{ minWidth: 120 }}>
            <MenuItem value="new">Neueste</MenuItem><MenuItem value="votes">Top</MenuItem>
          </TextField>
        </Stack>
      </Paper>

      <Stack spacing={2}>
        {loading ? [1,2,3].map(i => <Skeleton key={i} variant="rectangular" height={100} sx={{ borderRadius: 3 }} />) : 
          posts.map(p => <PostItem key={String(p.id)} p={p} />)}
      </Stack>

      <Pagination count={Math.ceil(count / PAGE_SIZE)} page={page} onChange={(_, p) => setPage(p)} disabled={loading} />
    </Page>
  );
}
