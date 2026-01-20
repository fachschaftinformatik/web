import { useEffect, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
    Avatar, Box, Stack, Typography, Button, IconButton, Chip, Divider, Tooltip, CircularProgress, Alert, Paper
} from "@mui/material";
import { useTheme } from "@mui/material/styles";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import EditIcon from "@mui/icons-material/Edit";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import PushPinOutlinedIcon from "@mui/icons-material/PushPinOutlined";

import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import {
    getForumPostsById,
    getForumPostsByIdComments,
    postForumPostsByIdComments,
    postForumPostsByIdVote,
    postForumCommentsByIdVote,
    deleteForumPostsById,
    putForumCommentsById
} from "@lib/api";
import {
    Post, Comment, Vote, CommentsSection, isoToShort
} from "../components";

const isoToShort = (iso?: string) => {
    if (!iso) return "Unbekannt";
    try {
        const d = new Date(iso);
        if (isNaN(d.getTime())) return "Unbekanntes Datum";
        return d.toLocaleString("de-DE", {
            year: "numeric",
            month: "short",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
        });
    } catch {
        return "Datum Fehler";
    }
};

export default function ViewPost() {
    const { id } = useParams();
    const navigate = useNavigate();
    const { user } = useAuth();
    const theme = useTheme();
    const isDark = theme.palette.mode === "dark";

    const [post, setPost] = useState<Post | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const isAdmin = user?.role === "admin" || user?.role === "editor";

    useEffect(() => {
        if (!id) return;
        Promise.all([
            getForumPostsById({ path: { id } }),
            getForumPostsByIdComments({ path: { id } })
        ]).then(([{ data: postData, error: postError }, { data: commentsData }]) => {
            if (postError || !postData) {
                setError("Beitrag nicht gefunden oder Fehler beim Laden.");
            } else {
                const parsedPost: Post = {
                    ...postData,
                    programs: postData.programs || [],
                    tags: postData.tags || [],
                    comments: []
                };

                const comments = (commentsData || []) as Comment[];
                setPost({ ...parsedPost, comments });
            }
        }).catch(err => {
            console.error(err);
            setError("Ein unerwarteter Fehler ist aufgetreten.");
        }).finally(() => setLoading(false));
    }, [id]);

    const handleVote = async (vote: Vote) => {
        if (!post || !user || !id) return;

        const oldVote = (post.user_vote as Vote) || 0;
        const oldVotes = Number(post.votes) || 0;
        const targetVote = vote === oldVote ? 0 : vote;

        let diff = 0;
        if (targetVote === 1) {
            if (oldVote === 1) diff = 0;
            else if (oldVote === -1) diff = 2;
            else diff = 1;
        } else if (targetVote === -1) {
            if (oldVote === 1) diff = -2;
            else if (oldVote === -1) diff = 0;
            else diff = -1;
        } else {
            if (oldVote === 1) diff = -1;
            else if (oldVote === -1) diff = 1;
        }

        const newVotes = oldVotes + diff;
        setPost(prev => prev ? ({ ...prev, user_vote: targetVote, votes: newVotes }) : null);

        try {
            await postForumPostsByIdVote({
                path: { id },
                body: { vote: targetVote }
            });
        } catch (err) {
            console.error("Vote failed", err);
        }
    };

    const handleAddComment = async (parentId: string | null, text: string) => {
        if (!id || !user) return;

        const { data } = await postForumPostsByIdComments({
            path: { id },
            body: { parent_id: parentId ?? undefined, text }
        });
        if (data) {
            const newComment: Comment = {
                ...data,
                author_name: user.name || user.email,
                author_avatar_url: user.avatar_url || "",
                author_id: user.id
            };
            setPost(prev => prev ? ({
                ...prev,
                comments: [...prev.comments, newComment],
                comment_count: (prev.comment_count || 0) + 1
            }) : null);
        }
    };

    const handleEditComment = async (commentId: string, text: string) => {
        if (!id || !user) return;

        const { data } = await putForumCommentsById({
            path: { id: commentId },
            body: { text }
        });

        if (data) {
            setPost(prev => {
                if (!prev) return null;
                const newComments = prev.comments.map(c =>
                    c.id === commentId ? { ...c, text: data.text || "", updated_at: data.updated_at || "" } : c
                );
                return { ...prev, comments: newComments };
            });
        }
    };

    const handleVoteComment = async (commentId: string, vote: Vote) => {
        if (!post || !user) return;

        try {
            await postForumCommentsByIdVote({
                path: { id: commentId },
                body: { vote }
            });

            setPost(prev => {
                if (!prev) return null;
                const newComments = prev.comments.map(c => {
                    if (c.id === commentId) {
                        const oldUserVote = (c.user_vote || 0) as Vote;
                        const oldVotes = c.votes || 0;
                        const newVotes = oldVotes - oldUserVote + vote;
                        return { ...c, user_vote: vote, votes: newVotes };
                    }
                    return c;
                });
                return { ...prev, comments: newComments };
            });
        } catch (err) {
            console.error("Comment vote failed", err);
        }
    };

    const handleDelete = async () => {
        if (!id || (!isAdmin && post?.author_id !== user?.id)) return;
        if (!confirm("Beitrag wirklich löschen?")) return;

        await deleteForumPostsById({
            path: { id }
        });
        navigate("/discussions");
    };

    if (loading) {
        return (
            <Sidebar user={user} title="Forum" maxWidth="md">
                <Box display="flex" justifyContent="center" py={8}><CircularProgress /></Box>
            </Sidebar>
        );
    }

    if (error || !post) {
        return (
            <Sidebar user={user} title="Forum" maxWidth="md">
                <Alert severity="error">{error || "Beitrag nicht gefunden"}</Alert>
                <Button sx={{ mt: 2 }} startIcon={<ArrowBackIcon />} onClick={() => navigate("/discussions")}>Zurück zum Forum</Button>
            </Sidebar>
        );
    }

    const netVotes = (Number(post.votes) || 0);
    const userVote = (post.user_vote as Vote) || 0;

    return (
        <Sidebar user={user} title="Forum" maxWidth="md">
            <Box pb={6}>
                <Button startIcon={<ArrowBackIcon />} onClick={() => navigate("/discussions")} sx={{ mb: 2, color: "text.secondary" }}>
                    Zurück zum Forum
                </Button>

                <Paper sx={{ p: { xs: 2, sm: 2.5, md: 4 }, borderRadius: 3 }}>
                    <Stack direction={{ xs: "column", sm: "row" }} spacing={{ xs: 2, sm: 3 }} alignItems="flex-start">
                        {/* Voting Section */}
                        <Stack
                            direction={{ xs: "row", sm: "column" }}
                            alignItems="center"
                            spacing={0.5}
                            sx={{
                                bgcolor: "action.hover",
                                borderRadius: 2,
                                p: 0.5,
                                width: { xs: "fit-content", sm: "auto" },
                                alignSelf: { xs: "center", sm: "flex-start" }
                            }}
                        >
                            <IconButton size="small" onClick={() => handleVote(1)} color={userVote === 1 ? "primary" : "default"}>
                                <ThumbUpOutlined fontSize="small" />
                            </IconButton>
                            <Typography variant="subtitle2" fontWeight={700} sx={{ px: { xs: 1, sm: 0 } }}>{netVotes}</Typography>
                            <IconButton size="small" onClick={() => handleVote(-1)} color={userVote === -1 ? "primary" : "default"}>
                                <ThumbDownOutlined fontSize="small" />
                            </IconButton>
                        </Stack>

                        <Stack spacing={3} sx={{ flex: 1, width: "100%" }}>
                            <Stack direction={{ xs: "column", sm: "row" }} justifyContent="space-between" alignItems="flex-start" spacing={2}>
                                <Stack spacing={1}>
                                    <Typography variant="h4" fontWeight={700} sx={{ fontSize: { xs: '1.75rem', md: '2.125rem' } }}>
                                        {post.title}
                                    </Typography>
                                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                                        <Avatar
                                            key={post.author_avatar_url || post.author_id}
                                            src={post.author_avatar_url || undefined}
                                            sx={{
                                                width: 24,
                                                height: 24,
                                                fontSize: "0.75rem",
                                                bgcolor: theme.palette.primary.main,
                                                fontWeight: "bold"
                                            }}
                                        >
                                            {post.author_name ? post.author_name[0].toUpperCase() : "A"}
                                        </Avatar>
                                        <Typography variant="body2" color="text.secondary">
                                            von <Typography component={Link} to={`/user/${post.author_id}`} variant="body2" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.author_name || "Anonym"}</Typography> · {isoToShort(post.created_at || "")}
                                        </Typography>
                                        {post.pinned === 1 && (
                                            <Chip label="Angepinnt" size="small" icon={<PushPinOutlinedIcon />} variant="outlined" sx={{ fontWeight: 600 }} />
                                        )}
                                    </Stack>
                                </Stack>

                                {(isAdmin || user?.id === post.author_id) && (
                                    <Stack direction="row" spacing={1}>
                                        <Tooltip title="Bearbeiten">
                                            <IconButton onClick={() => navigate(`/discussions/${post.id}/edit`)}>
                                                <EditIcon />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="Löschen">
                                            <IconButton onClick={handleDelete} color="error">
                                                <DeleteOutlineIcon />
                                            </IconButton>
                                        </Tooltip>
                                    </Stack>
                                )}
                            </Stack>

                            <Divider />

                            <Typography variant="body1" sx={{ whiteSpace: "pre-wrap" }}>
                                {post.body}
                            </Typography>


                            <Stack direction="row" spacing={1} flexWrap="wrap">
                                {post.programs.map((p) => (
                                    <Chip key={p} label={p} variant="outlined" size="small" />
                                ))}
                                {post.tags.map(t => (
                                    <Chip key={t} label={t} variant="outlined" size="small" />
                                ))}
                            </Stack>

                            <Divider />

                            <Box>
                                <CommentsSection
                                    comments={post.comments}
                                    onAdd={handleAddComment}
                                    onEdit={handleEditComment}
                                    appearance={{
                                        surface: isDark ? "rgba(255,255,255,0.05)" : "#f5f5f5",
                                        border: isDark ? "rgba(255,255,255,0.1)" : "#e0e0e0",
                                        accent: theme.palette.primary.main,
                                        textSecondary: theme.palette.text.secondary
                                    }}
                                    onVote={handleVoteComment}
                                    currentUser={user}
                                />
                            </Box>

                        </Stack>
                    </Stack>
                </Paper>
            </Box>
        </Sidebar >
    );
}
