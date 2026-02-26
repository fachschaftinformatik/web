'use client';

import * as React from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import IconButton from "@mui/material/IconButton";
import Chip from "@mui/material/Chip";
import InputAdornment from "@mui/material/InputAdornment";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Tooltip from "@mui/material/Tooltip";
import Pagination from "@mui/material/Pagination";
import Autocomplete from "@mui/material/Autocomplete";
import Avatar from "@mui/material/Avatar";
import Skeleton from "@mui/material/Skeleton";
import { alpha, useTheme } from "@mui/material/styles";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ShareIcon from "@mui/icons-material/Share";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import PushPinOutlinedIcon from "@mui/icons-material/PushPinOutlined";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import EditIcon from "@mui/icons-material/Edit";
import { Sidebar } from "@components/layout";
import { getAvatarUrl } from "@lib/images";

import {
  isoToShort,
  Post, Vote
} from "@components/discussions/components";
import type { SessionUser } from "@lib/types/session";
import { useSessionUser } from "@lib/hooks/useSessionUser";
import { useDiscussionList } from "@lib/hooks/useDiscussionList";

function PostItem({
  post,
  onVote,
  vote,
  onDelete,
  onTogglePin,
  user,
  canModerate,
  onShare,
}: {
  post: Post;
  onVote: (id: string, v: Vote) => void;
  vote: Vote;
  onDelete: (id: string) => void;
  onTogglePin: (id: string) => void;
  user: SessionUser | null;
  canModerate: boolean;
  onShare: (id: string) => void;
}) {
  const router = useRouter();
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const postId = String(post.id ?? "");
  const netVotes = (Number(post.votes) || 0);
  const canDelete = canModerate || (String(user?.id) === String(post.user_id));
  const [menuEl, setMenuEl] = React.useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuEl);

  const handleOpenMenu = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setMenuEl(e.currentTarget);
  };
  const handleCloseMenu = () => setMenuEl(null);

  const handleShare = (e: React.MouseEvent) => {
    e.stopPropagation();
    handleCloseMenu();
    onShare(postId);
  };

  const runMenuAction = (e: React.MouseEvent, action: () => void) => {
    e.stopPropagation();
    handleCloseMenu();
    action();
  };

  return (
    <Paper
      onClick={() => router.push(`/d/${postId}`)}
      elevation={0}
      sx={{
        p: 2,
        borderRadius: 3,
        cursor: "pointer",
        border: "1px solid",
        borderColor: theme.palette.divider,
        transition: "all 0.2s ease-in-out",
        bgcolor: post.type === "news"
          ? (isDark ? alpha(theme.palette.primary.main, 0.08) : alpha(theme.palette.primary.main, 0.04))
          : "var(--card-bg)",
        "&:hover": {
          transform: "translateY(-2px)",
          boxShadow: isDark ? "0 8px 16px rgba(0,0,0,0.4)" : "0 8px 16px rgba(15,110,46,0.08)",
          borderColor: theme.palette.primary.main,
          bgcolor: post.type === "news"
            ? (isDark ? alpha(theme.palette.primary.main, 0.12) : alpha(theme.palette.primary.main, 0.06))
            : (isDark ? alpha(theme.palette.primary.main, 0.04) : alpha(theme.palette.primary.main, 0.02)),
        },
      }}
    >
      <Stack direction="row" spacing={2} alignItems="flex-start">
        <Stack alignItems="center" spacing={0.5} sx={{ mt: 0.5 }}>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onVote(postId, 1);
            }}
            color={vote === 1 ? "primary" : "default"}
          >
            <ThumbUpOutlined fontSize="small" />
          </IconButton>
          <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
            {netVotes}
          </Typography>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onVote(postId, -1);
            }}
            color={vote === -1 ? "primary" : "default"}
          >
            <ThumbDownOutlined fontSize="small" />
          </IconButton>
        </Stack>

        <Stack spacing={1} sx={{ flex: 1 }}>
          <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
            <Stack direction="row" spacing={1} alignItems="center">
              <Typography variant="h6" sx={{ fontWeight: 700, lineHeight: 1.3 }}>
                {post.title}
              </Typography>
               {post.pinned === 1 && (
                 <PushPinOutlinedIcon fontSize="small" sx={{ color: 'primary.main' }} />
               )}
            </Stack>
            <IconButton size="small" onClick={handleOpenMenu}>
              <MoreVertIcon fontSize="small" />
            </IconButton>
            <Menu
              anchorEl={menuEl}
              open={openMenu}
              onClose={handleCloseMenu}
              onClick={(e) => e.stopPropagation()}
              disableScrollLock
              PaperProps={{
                elevation: 3,
                sx: {
                  borderRadius: 2,
                  mt: 0.5,
                  minWidth: 160,
                  '& .MuiMenuItem-root': {
                    px: 2,
                    py: 1,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1.5
                  }
                }
              }}
            >

              {String(user?.id) === String(post.user_id) && (
                <MenuItem
                  onClick={(e) => runMenuAction(e, () => router.push(`/d/${postId}/edit`))}
                >
                  <EditIcon sx={{ fontSize: 18 }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>Bearbeiten</Typography>
                </MenuItem>
              )}
              {canModerate && (
                <MenuItem
                  onClick={(e) => runMenuAction(e, () => onTogglePin(postId))}
                >
                  <PushPinOutlinedIcon sx={{ fontSize: 18 }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>{post.pinned === 1 ? "Entpinnen" : "Anpinnen"}</Typography>
                </MenuItem>
              )}
              <MenuItem
                onClick={handleShare}
              >
                <ShareIcon sx={{ fontSize: 18 }} />
                <Typography variant="body2" sx={{ fontWeight: 500 }}>Teilen</Typography>
              </MenuItem>
              {canDelete && (
                <MenuItem
                  onClick={(e) => runMenuAction(e, () => onDelete(postId))}
                  sx={{ color: "error.main" }}
                >
                  <DeleteOutlineIcon sx={{ fontSize: 18, color: "inherit" }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>Löschen</Typography>
                </MenuItem>
              )}
            </Menu>
          </Stack>

          <Typography
            variant="body2"
            sx={{
              color: theme.palette.text.secondary,
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
              overflow: "hidden",
            }}
          >
            {post.body}
          </Typography>

          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <Avatar
              src={getAvatarUrl(post.user_avatar_url)}
              sx={{
                width: 20,
                height: 20,
                fontSize: "0.65rem",
                bgcolor: theme.palette.primary.main,
                fontWeight: "bold"
              }}
            >
              {post.user_name ? post.user_name[0].toUpperCase() : "A"}
            </Avatar>
            <Typography variant="caption" sx={{ color: theme.palette.text.secondary }}>
              von <Typography component={Link} href={`/u/${post.user_id}`} onClick={(e) => e.stopPropagation()} variant="caption" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.user_name || "Anonym"}</Typography> · {isoToShort(post.created_at ?? "")}
              {post.updated_at && post.created_at && post.updated_at !== post.created_at && " (bearbeitet)"}
              {" · "}
              {post.comment_count} {post.comment_count === 1 ? "Kommentar" : "Kommentare"}
            </Typography>
          </Stack>

          <Stack direction="row" spacing={0.5} flexWrap="wrap">
            {(post.programs || []).map((program) => (
              <Chip
                key={String(program.id)}
                label={program.name}
                size="small"
                variant="outlined"
              />
            ))}
            {(post.tags || []).slice(0, 3).map((tag) => (
              <Chip
                key={tag}
                label={tag}
                size="small"
                variant="outlined"
              />
            ))}
          </Stack>
        </Stack>
      </Stack>
    </Paper>
  );
}



export default function DiscussionsPage() {
  const { user } = useSessionUser();
  const router = useRouter();

  const {
    posts,
    pageCount,
    loading,
    q,
    sort,
    activeProgramFilters,
    page,
    apiPrograms,
    canModerate,
    setQ,
    setSortOrder,
    setActiveProgramFilters,
    setPage,
    votePost,
    togglePin,
    deletePost,
    normalizeVote,
  } = useDiscussionList({ user });

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const targetId = params.get("postId") || params.get("post");
    if (targetId) {
      router.replace(`/d/${targetId}`);
    }
  }, [router]);

  const handleDelete = async (id: string) => {
    if (!confirm("Wirklich löschen?")) return;
    await deletePost(id);
  };

  const handleOpenCreate = () => router.push("/d/new");

  const handleShare = (id: string) => {
    const url = `${window.location.origin}/d/${id}`;
    navigator.clipboard.writeText(url);
  };

  return (
    <Sidebar user={user} title="Diskussionen" maxWidth="lg">
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 4 }}>
          <Box>
            <Typography variant="h4" fontWeight={700} gutterBottom>Diskussionen</Typography>
            <Typography variant="body1" color="text.secondary">Diskutiere mit deinen Kommilitonen über Studium, Campus und Events.</Typography>
          </Box>
        </Box>

        {loading && posts.length === 0 ? (
          <Stack spacing={2}>
            <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
              <Stack spacing={2}>
                <Skeleton variant="rectangular" height={40} sx={{ borderRadius: 2 }} />
                <Box sx={{ display: 'flex', gap: 2 }}>
                  <Skeleton variant="rectangular" height={40} sx={{ flex: 2, borderRadius: 2 }} />
                  <Skeleton variant="rectangular" height={40} sx={{ flex: 1, borderRadius: 2 }} />
                </Box>
              </Stack>
            </Paper>
            {[1, 2, 3, 4, 5].map(i => (
              <Paper key={i} elevation={0} sx={{ p: 2, borderRadius: 3, border: "1px solid", borderColor: "divider" }}>
                <Stack direction="row" spacing={2}>
                  <Stack alignItems="center" spacing={1}>
                    <Skeleton variant="circular" width={32} height={32} />
                    <Skeleton variant="text" width={20} />
                    <Skeleton variant="circular" width={32} height={32} />
                  </Stack>
                  <Stack spacing={1} sx={{ flex: 1 }}>
                    <Skeleton variant="text" width="60%" height={32} />
                    <Skeleton variant="text" width="90%" height={20} />
                    <Skeleton variant="text" width="40%" height={20} />
                    <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                      <Skeleton variant="circular" width={20} height={20} />
                      <Skeleton variant="text" width="30%" />
                    </Stack>
                  </Stack>
                </Stack>
              </Paper>
            ))}
          </Stack>
        ) : (
          <>
            <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
              <Stack spacing={2}>
                <Stack direction="row" spacing={2}>
                  <TextField
                    fullWidth
                    placeholder="Suche in Titel, Text oder Tags…"
                    value={q}
                    onChange={e => setQ(e.target.value)}
                    size="small"
                    InputProps={{
                      startAdornment: <InputAdornment position="start"><SearchIcon color="primary" /></InputAdornment>,
                      sx: { borderRadius: 2 }
                    }}
                  />
                  <Tooltip title="Beitrag erstellen">
                    <IconButton
                      onClick={handleOpenCreate}
                      sx={{
                        bgcolor: 'primary.main',
                        color: 'primary.contrastText',
                        borderRadius: 2,
                        width: 40,
                        height: 40,
                        '&:hover': { bgcolor: 'primary.dark' }
                      }}
                    >
                      <AddIcon />
                    </IconButton>
                  </Tooltip>
                </Stack>
                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', width: '100%', alignItems: 'center' }}>
                  <Autocomplete
                    multiple
                    options={apiPrograms}
                    getOptionLabel={o => o.name || ""}
                    value={apiPrograms.filter(p => activeProgramFilters.includes(p.name || ""))}
                    isOptionEqualToValue={(o, v) => String(o.id) === String(v.id)}
                    onChange={(_, n) => setActiveProgramFilters(n.map(v => v.name || ""))}
                    renderInput={p => <TextField {...p} label="Studiengang" size="small" placeholder="Wählen..." />}
                    size="small"
                    sx={{ flex: 2, minWidth: 300 }}
                  />
                  <TextField
                    select
                    label="Filter"
                    value={sort}
                    onChange={(e) => {
                      setSortOrder(e.target.value);
                    }}
                    size="small"
                    sx={{ flex: 1, minWidth: 150 }}
                  >
                    <MenuItem value="relevant">Relevant</MenuItem>
                    <MenuItem value="new">Neuste</MenuItem>
                    <MenuItem value="votes">Top</MenuItem>
                  </TextField>
                </Box>
              </Stack>
            </Paper>

            <Stack spacing={2} sx={{ opacity: loading ? 0.6 : 1, transition: 'opacity 0.2s' }}>
              {posts.map(p => (
                <PostItem
                  key={String(p.id)}
                  post={p}
                  vote={normalizeVote(p.user_vote)}
                  onVote={votePost}
                  onDelete={handleDelete}
                   onTogglePin={togglePin}
                  user={user}
                  canModerate={canModerate}
                  onShare={handleShare}
                />
              ))}
              {posts.length === 0 && !loading && (
                <Stack spacing={1.5}>
                  <Typography color="text.secondary">Keine Treffer. Suchbegriff oder Filter anpassen.</Typography>
                </Stack>
              )}
            </Stack>

            {pageCount > 1 && (
              <Box display="flex" justifyContent="center" mt={4}>
                <Pagination
                  count={pageCount}
                  page={page}
                  onChange={(_, v) => setPage(v)}
                  color="primary"
                  shape="rounded"
                  variant="outlined"
                  disabled={loading}
                />
              </Box>
            )}
          </>
        )}

      </Box>
    </Sidebar>
  );
}
