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
import LinkIcon from "@mui/icons-material/Link";

import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import {
    getForumPostsById,
    getForumPostsByIdComments,
    postForumPostsByIdComments,
    postForumPostsByIdVote,
    postForumCommentsByIdVote,
    deleteForumPostsById,
    putForumCommentsById,
    getAuthCsrf
} from "@lib/api";
import {
    Post, Comment, Vote, PROGRAM_META_MAP, isoToShort, CommentsSection, Program
} from "../components";

const safeParseArray = (jsonString: any): any[] => {
    try {
        const parsed = JSON.parse(jsonString as string);
        return Array.isArray(parsed) ? parsed : [];
    } catch {
        return [];
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
        // Wrap in microtask to avoid sync setState warning
        Promise.resolve().then(() => setLoading(true));
        Promise.all([
            getForumPostsById({ path: { id } }),
            getForumPostsByIdComments({ path: { id } })
        ]).then(([{ data: postData, error: postError }, { data: commentsData }]) => {
            if (postError || !postData) {
                setError("Beitrag nicht gefunden oder Fehler beim Laden.");
            } else {
                const parsedPost: Post = {
                    ...postData,
                    programs: safeParseArray(postData.programs) as Program[],
                    tags: safeParseArray(postData.tags) as string[],
                    links: safeParseArray(postData.links) as string[],
                    comments: []
                };

                const comments = (commentsData || []) as Comment[];
                setPost({ ...parsedPost, comments });
            }
        }).catch(err => {
            console.error(err);
            setError("Ein ungewarteter Fehler ist aufgetreten.");
        }).finally(() => setLoading(false));
    }, [id]);

    const handleVote = async (vote: Vote) => {
        if (!post || !user || !id) return;

        const oldVote = (post.user_vote as Vote) || 0;
        const oldVotes = Number(post.votes) || 0;

        const targetVote = vote === oldVote ? 0 : vote; // Toggle if same

        // Calculate new net votes locally
        // If oldVote was 1 and target is 0, net -1.
        // If oldVote was 1 and target is -1, net -2.
        // If oldVote was 0 and target is 1, net +1.
        let diff = 0;
        if (targetVote === 1) {
            if (oldVote === 1) diff = 0; // shouldn't happen with toggle logic
            else if (oldVote === -1) diff = 2;
            else diff = 1;
        } else if (targetVote === -1) {
            if (oldVote === 1) diff = -2;
            else if (oldVote === -1) diff = 0;
            else diff = -1;
        } else { // target 0
            if (oldVote === 1) diff = -1;
            else if (oldVote === -1) diff = 1;
        }

        const newVotes = oldVotes + diff;

        setPost(prev => prev ? ({ ...prev, user_vote: targetVote, votes: newVotes }) : null);

        try {
            const { data: csrfData } = await getAuthCsrf();
            if (csrfData?.csrf) {
                await postForumPostsByIdVote({
                    path: { id },
                    body: { vote: targetVote },
                    headers: { "X-CSRF-Token": csrfData.csrf }
                });
            }
        } catch (err) {
            console.error("Vote failed", err);
        }
    };

    const handleAddComment = async (parentId: string | null, text: string) => {
        if (!id || !user) return;
        const { data: csrfData } = await getAuthCsrf();
        if (!csrfData?.csrf) return;

        const { data } = await postForumPostsByIdComments({
            path: { id },
            body: { parent_id: parentId ?? undefined, text },
            headers: { "X-CSRF-Token": csrfData.csrf }
        });
        if (data) {
            const newComment: Comment = {
                ...data,
                author_name: user.name || user.email
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
        const { data: csrfData } = await getAuthCsrf();
        if (!csrfData?.csrf) return;

        const { data } = await putForumCommentsById({
            path: { id: commentId },
            body: { text },
            headers: { "X-CSRF-Token": csrfData.csrf }
        });

        if (data) {
            setPost(prev => {
                if (!prev) return null;
                const newComments = prev.comments.map(c =>
                    c.id === commentId ? { ...c, text: data.text, updated_at: data.updated_at } : c
                );
                return { ...prev, comments: newComments };
            });
        }
    };

    const handleVoteComment = async (commentId: string, vote: Vote) => {
        if (!post || !user) return;
        const { data: csrfData } = await getAuthCsrf();
        if (!csrfData?.csrf) return;

        try {
            await postForumCommentsByIdVote({
                path: { id: commentId },
                body: { vote },
                headers: { "X-CSRF-Token": csrfData.csrf }
            });

            // Update local state
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
        if (!id || !isAdmin && post?.author_id !== user?.id) return;
        if (!confirm("Beitrag wirklich löschen?")) return;

        const { data: csrfData } = await getAuthCsrf();
        if (!csrfData?.csrf) return;

        await deleteForumPostsById({
            path: { id },
            headers: { "X-CSRF-Token": csrfData.csrf }
        });
        navigate("/forum");
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
                <Button sx={{ mt: 2 }} startIcon={<ArrowBackIcon />} onClick={() => navigate("/forum")}>Zurück zum Forum</Button>
            </Sidebar>
        );
    }

    const netVotes = (Number(post.votes) || 0);
    const userVote = (post.user_vote as Vote) || 0;

    return (
        <Sidebar user={user} title={post.title} maxWidth="md">
            <Box pb={6}>
                <Button startIcon={<ArrowBackIcon />} onClick={() => navigate("/forum")} sx={{ mb: 2, color: "text.secondary" }}>
                    Zurück zum Forum
                </Button>

                <Paper sx={{ p: { xs: 2.5, md: 4 }, borderRadius: 3 }}>
                    <Stack direction="row" spacing={3} alignItems="flex-start">
                        <Stack alignItems="center" spacing={0.5} sx={{ bgcolor: "action.hover", borderRadius: 2, p: 0.5 }}>
                            <IconButton size="small" onClick={() => handleVote(1)} color={userVote === 1 ? "primary" : "default"}>
                                <ThumbUpOutlined fontSize="small" />
                            </IconButton>
                            <Typography variant="subtitle2" fontWeight={700}>{netVotes}</Typography>
                            <IconButton size="small" onClick={() => handleVote(-1)} color={userVote === -1 ? "primary" : "default"}>
                                <ThumbDownOutlined fontSize="small" />
                            </IconButton>
                        </Stack>

                        <Stack spacing={3} sx={{ flex: 1, width: "100%" }}>
                            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                                <Stack spacing={1}>
                                    <Typography variant="h4" fontWeight={700}>
                                        {post.title}
                                    </Typography>
                                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                                        <Avatar
                                            src={post.author_avatar_url || undefined}
                                            sx={{
                                                width: 24,
                                                height: 24,
                                                fontSize: "0.75rem",
                                                bgcolor: theme.palette.primary.main,
                                                fontWeight: "bold"
                                            }}
                                        />
                                        <Typography variant="body2" color="text.secondary">
                                            von <Typography component={Link} to={`/user/${post.author_id}`} variant="body2" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.author_name || "Anonym"}</Typography> · {isoToShort(post.created_at)}
                                        </Typography>
                                        {post.pinned === 1 && (
                                            <Chip label="Angepinnt" size="small" icon={<PushPinOutlinedIcon />} />
                                        )}
                                    </Stack>
                                </Stack>

                                {(isAdmin || user?.id === post.author_id) && (
                                    <Stack direction="row" spacing={1}>
                                        <Tooltip title="Bearbeiten">
                                            <IconButton onClick={() => navigate(`/forum/${post.id}/edit`)}>
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

                            {post.image_url && (
                                <Box component="img" src={post.image_url} sx={{ width: "100%", maxHeight: 500, objectFit: "contain", borderRadius: 2 }} />
                            )}

                            {post.links && post.links.length > 0 && (
                                <Stack spacing={1}>
                                    <Typography variant="subtitle2" fontWeight={700}>Links</Typography>
                                    {post.links.map((link, i) => (
                                        <Button key={i} href={link} target="_blank" startIcon={<LinkIcon />} sx={{ justifyContent: "flex-start" }}>
                                            {link}
                                        </Button>
                                    ))}
                                </Stack>
                            )}

                            <Stack direction="row" spacing={1} flexWrap="wrap">
                                {post.programs.map((p) => (
                                    <Chip key={p} label={PROGRAM_META_MAP[p]?.shortLabel || p} variant="outlined" size="small" />
                                ))}
                                {post.tags.map(t => (
                                    <Chip key={t} label={t} variant="outlined" size="small" />
                                ))}
                            </Stack>

                            <Divider />

                            <Box>
                                {/* <Typography variant="h6" gutterBottom>Kommentare ({post.comments.length})</Typography> */}
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
