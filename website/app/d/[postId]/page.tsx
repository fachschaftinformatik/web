'use client';

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
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
import {
    CommentsSection, isoToShort
} from "@components/discussions/components";
import { getAvatarUrl } from "@lib/images";
import { useSessionUser } from "@lib/hooks/useSessionUser";
import { useDiscussionPost } from "@lib/hooks/useDiscussionPost";

export default function ViewPost({ params }: { params: Promise<{ postId: string }> }) {
    const { postId } = React.use(params);
    const router = useRouter();
    const { user } = useSessionUser();
    const theme = useTheme();
    const isDark = theme.palette.mode === "dark";
    const [copied, setCopied] = useState(false);

    const {
        post,
        loading,
        error,
        isAdmin,
        netVotes,
        userVote,
        votePost,
        addComment,
        editComment,
        voteComment,
        deletePost,
    } = useDiscussionPost({
        postIdParam: postId,
        user,
        accentColor: theme.palette.primary.main,
    });

    const handleShare = () => {
        const url = window.location.origin + window.location.pathname;
        navigator.clipboard.writeText(url).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        });
    };

    const handleDelete = async () => {
        if (!confirm("Beitrag wirklich löschen?")) return;

        const didDelete = await deletePost();
        if (didDelete) {
            router.push("/discussions");
        }
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
                    onClick={() => router.push("/discussions")}
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

    return (
        <Sidebar user={user} title="Diskussion" maxWidth="md">
            <Box pb={6}>
                <Button
                    startIcon={<ArrowBackRounded />}
                    onClick={() => router.push("/discussions")}
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
                            <IconButton size="small" onClick={() => votePost(1)} color={userVote === 1 ? "primary" : "default"}>
                                <ThumbUpOutlined fontSize="small" />
                            </IconButton>
                            <Typography variant="subtitle2" fontWeight={700} sx={{ px: { xs: 1, sm: 0 } }}>{netVotes}</Typography>
                            <IconButton size="small" onClick={() => votePost(-1)} color={userVote === -1 ? "primary" : "default"}>
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
                                            von <Typography component={Link} href={`/u/${post.user_id}`} variant="body2" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.user_name || "Anonym"}</Typography> · {isoToShort(post.created_at || "")}
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
                                                <IconButton onClick={() => router.push(`/d/${post.id}/edit`)}>
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
                                    onAdd={addComment}
                                    onEdit={editComment}
                                    appearance={{
                                        surface: isDark ? "rgba(255,255,255,0.05)" : "#f5f5f5",
                                        border: isDark ? "rgba(255,255,255,0.1)" : "#e0e0e0",
                                        accent: theme.palette.primary.main,
                                        textSecondary: theme.palette.text.secondary
                                    }}
                                    onVote={voteComment}
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
