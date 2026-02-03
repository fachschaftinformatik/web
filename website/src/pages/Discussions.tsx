import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import Stack from "@mui/material/Stack";
import Paper from "@mui/material/Paper";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import InputAdornment from "@mui/material/InputAdornment";
import MenuItem from "@mui/material/MenuItem";
import Autocomplete from "@mui/material/Autocomplete";
import Skeleton from "@mui/material/Skeleton";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";

import Page from "@components/Page";
import Pagination from "@components/Pagination";
import PostItem from "@components/PostItem";
import { useAuth } from "@lib/auth";
import { getDiscussions, getPrograms, postDiscussionsByPostIdVote, deleteDiscussionsByPostId } from "@lib/api";
import type { DtoDiscussionPostResponse as Post, DtoProgramResponse as Prog } from "@lib/api";
import { POSTS_PER_PAGE, Vote } from "./DiscussionComponents";

const PAGE_SIZE = POSTS_PER_PAGE;

export default function Discussions() {
  const { user } = useAuth();
  const navigate = useNavigate();
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

  return (
    <Page title="Beiträge" description="Hier könnt ihr euch mit eurem Kommilitonen direkt austauschen. Außerdem veröffentlicht die FSV hier anstehende Events und weitere Infos.">
      <Paper variant="outlined" sx={{ p: 2, mb: 3 }}>
        <Stack spacing={2}>
          <Stack direction="row" spacing={2} alignItems="stretch">
            <TextField 
              fullWidth 
              placeholder="Beitrag finden..." 
              value={q} 
              onChange={e => setQ(e.target.value)} 
              size="small" 
              slotProps={{ 
                input: { 
                  startAdornment: (
                    <InputAdornment position="start" sx={{ mr: 1 }}>
                      <SearchIcon color="primary" />
                    </InputAdornment>
                  ),
                  sx: { height: 48, '& input': { height: '100%', py: 0 } }
                } 
              }} 
            />
            <Button variant="contained" startIcon={<AddIcon />} onClick={() => navigate("/d/new")} sx={{ whiteSpace: 'nowrap', height: 48 }}>Neu</Button>
          </Stack>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems="stretch">
            <Autocomplete 
              multiple 
              limitTags={1}
              options={progs.map(p => p.name || '')} 
              value={selProgs} 
              onChange={(_, v) => setSelProgs(v)} 
              renderInput={p => <TextField {...p} label="Studiengang" size="small" />} 
              sx={{ 
                flex: 1,
                '& .MuiInputBase-root': {
                  minHeight: 48,
                  maxHeight: 48,
                  display: 'flex',
                  alignItems: 'center',
                  overflowY: 'auto',
                  '&::-webkit-scrollbar': { display: 'none' },
                  scrollbarWidth: 'none',
                  msOverflowStyle: 'none'
                },
                '& .MuiInputLabel-root': {
                  transform: 'translate(14px, 14px) scale(1)',
                  '&.MuiInputLabel-shrink': {
                    transform: 'translate(14px, -9px) scale(0.75)'
                  }
                }
              }} 
            />
            <TextField 
              select 
              label="Sortierung" 
              value={sort} 
              onChange={e => setSort(e.target.value)} 
              size="small" 
              sx={{ minWidth: 150 }}
              slotProps={{ input: { sx: { height: 48 } } }}
            >
              <MenuItem value="new">Neueste</MenuItem>
              <MenuItem value="votes">Top</MenuItem>
            </TextField>
          </Stack>
        </Stack>
      </Paper>

      <Stack spacing={2}>
        {loading ? [1,2,3].map(i => <Skeleton key={i} variant="rectangular" height={140} sx={{ borderRadius: 3 }} />) : 
          posts.map(p => <PostItem key={String(p.id)} p={p} user={user} onVote={vote} onDelete={del} />)}
      </Stack>

      <Pagination count={Math.ceil(count / PAGE_SIZE)} page={page} onChange={(_, p) => setPage(p)} disabled={loading} />
    </Page>
  );
}
