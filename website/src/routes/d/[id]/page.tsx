import { useEffect, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Tooltip from "@mui/material/Tooltip";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";
import Paper from "@mui/material/Paper";
import { useTheme } from "@mui/material/styles";
import ArrowBackRounded from "@mui/icons-material/ArrowBackRounded";
import EditIcon from "@mui/icons-material/Edit";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ShareIcon from "@mui/icons-material/Share";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import {
    getDiscussionsByPostId,
    getDiscussionsByPostIdComments,
    postDiscussionsByPostIdComments,
    postDiscussionsByPostIdVote,
    postDiscussionsCommentsByCommentIdVote,
    deleteDiscussionsByPostId,
    putDiscussionsCommentsByCommentId
} from "@lib/api";
import {
    Post, Comment, Vote, CommentsSection, isoToShort
} from "../components";
import { getAvatarUrl } from "@lib/images";

export default function ViewPost() {
    const { postId } = useParams<{ postId: string }>();
    const navigate = useNavigate();
    const { user } = useAuth();
    const theme = useTheme();
    const isDark = theme.palette.mode === "dark";

    const [post, setPost] = useState<Post | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);

    const isAdmin = user?.role === "admin" || user?.role === "editor";

    useEffect(() => {
        if (!postId) return;
        Promise.all([
            getDiscussionsByPostId({ path: { postId } }),
            getDiscussionsByPostIdComments({ path: { postId } })
        ]).then(([{ data: postData, error: postError }, { data: commentsData }]) => {
            if (postError || !postData) {
                setError("Beitrag nicht gefunden oder Fehler beim Laden.");
            } else {
                const parsedPost: Post = {
                    ...postData,
                    comments: []
                };

                const comments = (commentsData || []) as Comment[];
                setPost({ ...parsedPost, comments });

                // Scroll to comment if hash is present
                const hash = window.location.hash.replace("#", "");
                if (hash) {
                    setTimeout(() => {
                        const element = document.getElementById(hash);
                        if (element) {
                            element.scrollIntoView({ behavior: "smooth", block: "center" });
                            
                            // Find the text box (Paper) for a more targeted highlight
                            const textBox = element.querySelector(".MuiPaper-root") as HTMLElement;
                            if (textBox) {
                                textBox.style.transition = "all 0.5s ease";
                                const originalBorder = textBox.style.borderColor;
                                const originalShadow = textBox.style.boxShadow;
                                
                                textBox.style.borderColor = theme.palette.primary.main;
                                textBox.style.boxShadow = `0 0 0 2px ${theme.palette.primary.main}33`; // 33 is ~20% alpha
                                
                                setTimeout(() => {
                                    textBox.style.borderColor = originalBorder;
                                    textBox.style.boxShadow = originalShadow;
                                }, 3000);
                            }
                        }
                    }, 800);
                }
            }
        }).catch(err => {
            console.error(err);
            setError("Ein unerwarteter Fehler ist aufgetreten.");
        }).finally(() => setLoading(false));
    }, [postId, theme.palette.action.selected, theme.palette.primary.main]);

    const handleShare = () => {
        const url = window.location.origin + window.location.pathname;
        navigator.clipboard.writeText(url).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        });
    };

    const handleVote = async (vote: Vote) => {
        if (!post || !user || !postId) return;

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
            await postDiscussionsByPostIdVote({
                path: { postId },
                body: { vote: targetVote }
            });
        } catch (err) {
            console.error("Vote failed", err);
        }
    };

    const handleAddComment = async (parentId: string | null, text: string) => {
        if (!postId || !user) return;

        const { data } = await postDiscussionsByPostIdComments({
            path: { postId },
            body: { parent_id: parentId ? String(parentId) : undefined, text }
        });
        if (data) {
            const newComment: Comment = {
                ...data,
                user_name: user.name || user.email || "Anonym",
                user_avatar_url: user.avatar_url || "",
                user_id: user.id
            };
            setPost(prev => prev ? ({
                ...prev,
                comments: [...prev.comments, newComment],
                comment_count: (prev.comment_count || 0) + 1
            }) : null);
        }
    };

    const handleEditComment = async (commentId: string, text: string) => {
        if (!postId || !user) return;

        const { data } = await putDiscussionsCommentsByCommentId({
            path: { commentId },
            body: { text }
        });

        if (data) {
            setPost(prev => {
                if (!prev) return null;
                const newComments = prev.comments.map(c =>
                    String(c.id) === commentId ? { ...c, text: data.text || "", updated_at: data.updated_at || "" } : c
                );
                return { ...prev, comments: newComments };
            });
        }
    };

    const handleVoteComment = async (commentId: string, vote: Vote) => {
        if (!post || !user) return;

        try {
            await postDiscussionsCommentsByCommentIdVote({
                path: { commentId },
                body: { vote }
            });

            setPost(prev => {
                if (!prev) return null;
                const newComments = prev.comments.map(c => {
                    if (String(c.id) === commentId) {
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
        if (!postId || (!isAdmin && String(post?.user_id) !== String(user?.id))) return;
        if (!confirm("Beitrag wirklich löschen?")) return;

        await deleteDiscussionsByPostId({
            path: { postId }
        });
        navigate("/discussions");
    };

    if (loading) {
        return (
            <Sidebar user={user} title="Diskussion" maxWidth="md">
                <Box display="flex" justifyContent="center" py={8}><CircularProgress /></Box>
            </Sidebar>
        );
    }

    if (error || !post) {
        return (
            <Sidebar user={user} title="Diskussion" maxWidth="md">
                <Alert severity="error">{error || "Beitrag nicht gefunden"}</Alert>
                <Button
                    startIcon={<ArrowBackRounded />}
                    onClick={() => navigate("/discussions")}
                    sx={{
                        mt: 2,
                        color: 'text.secondary',
                        textTransform: 'none',
                        '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                    }}
                >
                    Zurück zum Forum
                </Button>
            </Sidebar>
        );
    }

    const netVotes = (Number(post.votes) || 0);
    const userVote = (post.user_vote as Vote) || 0;

    return (
        <Sidebar user={user} title="Diskussion" maxWidth="md">
            <Box pb={6}>
                <Button
                    startIcon={<ArrowBackRounded />}
                    onClick={() => navigate("/discussions")}
                    sx={{
                        mb: 2,
                        color: 'text.secondary',
                        textTransform: 'none',
                        '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                    }}
                >
                    Zurück zur Übersicht
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
                                            key={post.user_avatar_url || String(post.user_id)}
                                            src={getAvatarUrl(post.user_avatar_url)}
                                            sx={{
                                                width: 24,
                                                height: 24,
                                                fontSize: "0.75rem",
                                                bgcolor: theme.palette.primary.main,
                                                fontWeight: "bold"
                                            }}
                                        >
                                            {post.user_name ? post.user_name[0].toUpperCase() : "A"}
                                        </Avatar>
                                        <Typography variant="body2" color="text.secondary">
                                            von <Typography component={Link} to={`/u/${post.user_id}`} variant="body2" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.user_name || "Anonym"}</Typography> · {isoToShort(post.created_at || "")}
                                        </Typography>
                                    </Stack>
                                </Stack>

                                <Stack direction="row" spacing={1}>
                                    <Tooltip title={copied ? "Kopiert!" : "Link kopieren"}>
                                        <IconButton onClick={handleShare}>
                                            <ShareIcon fontSize="small" />
                                        </IconButton>
                                    </Tooltip>
                                    {(isAdmin || String(user?.id) === String(post.user_id)) && (
                                        <>
                                            <Tooltip title="Bearbeiten">
                                                <IconButton onClick={() => navigate(`/d/${post.id}/edit`)}>
                                                    <EditIcon fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip title="Löschen">
                                                <IconButton onClick={handleDelete} color="error">
                                                    <DeleteOutlineIcon fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                        </>
                                    )}
                                </Stack>
                            </Stack>

                            <Divider />

                            <Typography variant="body1" sx={{ whiteSpace: "pre-wrap" }}>
                                {post.body}
                            </Typography>


                            <Stack direction="row" spacing={1} flexWrap="wrap">
                                {(post.programs || []).map((p) => (
                                    <Chip key={String(p.id)} label={p.name} variant="outlined" size="small" />
                                ))}
                                {(post.tags || []).map(t => (
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
