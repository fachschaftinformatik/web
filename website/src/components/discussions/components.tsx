'use client';

import * as React from "react";
import Link from "next/link";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import Avatar from "@mui/material/Avatar";
import IconButton from "@mui/material/IconButton";
import Tooltip from "@mui/material/Tooltip";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import SendIcon from "@mui/icons-material/Send";
import EditIcon from "@mui/icons-material/Edit";
import ReplyIcon from "@mui/icons-material/Reply";
import ShareIcon from "@mui/icons-material/Share";
import MoreHorizIcon from "@mui/icons-material/MoreHoriz";
import ReportOutlinedIcon from "@mui/icons-material/ReportOutlined";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import { getAvatarUrl } from "@lib/images";
import type { DtoDiscussionPostResponse as ApiPost, DtoDiscussionCommentResponse as ApiComment, DtoUserResponse } from "@lib/api";

export type Program = string;
export type Vote = -1 | 0 | 1;

export type Comment = ApiComment;
export type Post = ApiPost & {
    comments: Comment[];
};

export type CommentAppearance = {
    surface: string;
    border: string;
    accent: string;
    textSecondary: string;
};

export const COMMENT_COLLAPSE_LIMIT = 6;
export const POSTS_PER_PAGE = 10;

export { FORUM_CATEGORIES, FORUM_TAGS } from "@lib/config";

export const isoToShort = (iso?: string) => {
    if (!iso)
        return "Unbekannt";
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
}

export type CommentNode = Comment & { children: CommentNode[] };

const toVote = (value: number | undefined | null): Vote => {
    if (value === 1 || value === -1) {
        return value;
    }

    return 0;
};

export function buildCommentTree(comments: Comment[]): CommentNode[] {
    if (!Array.isArray(comments)) return [];
    const map = new Map<string, CommentNode>();
    const roots: CommentNode[] = [];
    comments.forEach((c) => {
        if (c.id) map.set(String(c.id), { ...c, children: [] });
    });
    comments.forEach((c) => {
        if (!c.id) return;
        const node = map.get(String(c.id))!;
        const pid = c.parent_id ? String(c.parent_id) : null;
        if (!pid) roots.push(node);
        else {
            const parent = map.get(pid);
            if (parent) parent.children.push(node);
            else roots.push(node);
        }
    });
    return roots;
}

export function renderTextWithMentions(text: string) {
    if (!text) return "";
    const mentionRegex = /@[A-Za-z0-9._-]+/g;
    const nodes: React.ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;
    while ((match = mentionRegex.exec(text)) !== null) {
        const mIndex = match.index;
        if (mIndex > lastIndex) {
            nodes.push(text.slice(lastIndex, mIndex));
        }
        const mention = match[0];
        nodes.push(
            <Box
                component="span"
                key={`${mention}-${mIndex}`}
                sx={{
                    color: "primary.light",
                    fontWeight: 600,
                    cursor: "pointer",
                    transition: "color .2s ease",
                    "&:hover": { textDecoration: "underline", color: "primary.main" },
                }}
            >
                {mention}
            </Box>
        );
        lastIndex = mIndex + mention.length;
    }
    if (lastIndex < text.length) {
        nodes.push(text.slice(lastIndex));
    }
    return nodes.length ? nodes : [text];
}

export function CommentThread({
    node,
    onReply,
    onEdit,
    onVote,
    depth = 0,
    appearance,
    currentUser,
}: {
    node: CommentNode;
    onReply: (parentId: string | null, text: string) => void;
    onEdit: (commentId: string, text: string) => void;
    onVote: (commentId: string, vote: Vote) => void;
    depth?: number;
    appearance: CommentAppearance;
    currentUser: DtoUserResponse | null;
}) {
    const [replyOpen, setReplyOpen] = React.useState(false);
    const [editing, setEditing] = React.useState(false);
    const [repliesExpanded, setRepliesExpanded] = React.useState(false);
    const [editText, setEditText] = React.useState(node.text || "");
    const [copied, setCopied] = React.useState(false);
    const [menuAnchorEl, setMenuAnchorEl] = React.useState<null | HTMLElement>(null);

    const openMenu = Boolean(menuAnchorEl);
    const handleOpenMenu = (event: React.MouseEvent<HTMLElement>) => setMenuAnchorEl(event.currentTarget);
    const handleCloseMenu = () => setMenuAnchorEl(null);

    const isAuthor = String(currentUser?.id) === String(node.user_id);
    const canReply = !!currentUser;
    const canEdit = isAuthor || currentUser?.role === "admin" || currentUser?.role === "editor";
    const isEdited = node.updated_at && node.created_at && node.updated_at !== node.created_at;

    const netVotes = node.votes || 0;
    const userVote = toVote(node.user_vote);

    const totalDescendants = React.useMemo(() => {
        const count = (n: CommentNode): number =>
            n.children.reduce((acc, child) => acc + 1 + count(child), 0);
        return count(node);
    }, [node]);

    const handleShare = () => {
        const url = `${window.location.origin}${window.location.pathname}#${node.id}`;
        navigator.clipboard.writeText(url).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        });
    };

    React.useEffect(() => {
        const hash = window.location.hash.replace("#", "");
        if (hash === String(node.id)) {
            // Target found
        } else if (hash) {
            const hasTargetChild = (n: CommentNode): boolean =>
                n.children.some(c => String(c.id) === hash || hasTargetChild(c));

            if (hasTargetChild(node)) {
                setRepliesExpanded(true);
            }
        }
    }, [node]);

    return (
        <Box
            id={String(node.id)}
            sx={{
                mt: 1.5,
                pl: depth ? { xs: 0.75, sm: 2 } : 0,
                borderLeft: depth ? `1.5px solid ${appearance.border}` : "none",
                width: "100%",
                maxWidth: "100%",
                boxSizing: "border-box"
            }}
        >
            <Stack direction="row" spacing={{ xs: 1, sm: 2 }} alignItems="flex-start">
                <Avatar
                    src={getAvatarUrl(node.user_avatar_url)}
                    sx={{
                        width: depth ? { xs: 20, sm: 24 } : { xs: 28, sm: 32 },
                        height: depth ? { xs: 20, sm: 24 } : { xs: 28, sm: 32 },
                        bgcolor: appearance.accent,
                        fontSize: depth ? "0.6rem" : "0.8rem",
                        fontWeight: "bold",
                        flexShrink: 0
                    }}
                >
                    {node.user_name ? node.user_name[0].toUpperCase() : "A"}
                </Avatar>

                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Paper
                        elevation={0}
                        sx={{
                            p: { xs: 1.5, sm: 2 },
                            borderRadius: 3,
                            borderRadiusTopLeft: depth ? 3 : 0,
                            bgcolor: appearance.surface,
                            border: `1px solid ${appearance.border}`,
                            width: "100%",
                            boxSizing: "border-box",
                            overflow: "hidden"
                        }}
                    >
                        <Stack spacing={1}>
                            <Stack direction="row" justifyContent="space-between" alignItems="center">
                                <Stack direction="row" spacing={1} alignItems="center" sx={{ minWidth: 0 }}>
                                    <Typography
                                        component={Link}
                                        href={`/u/${node.user_id}`}
                                        variant="subtitle2"
                                        fontWeight={700}
                                        noWrap
                                        sx={{
                                            color: "inherit",
                                            textDecoration: "none",
                                            fontSize: { xs: '0.75rem', sm: '0.875rem' },
                                            "&:hover": { color: appearance.accent, cursor: "pointer" }
                                        }}
                                    >
                                        {node.user_name || "Anonym"}
                                    </Typography>
                                    <Typography variant="caption" sx={{ color: appearance.textSecondary, whiteSpace: 'nowrap', fontSize: { xs: '0.7rem', sm: '0.75rem' } }}>
                                        · {isoToShort(node.created_at ?? "")}
                                        {isEdited && " (Bearbeitet)"}
                                    </Typography>
                                </Stack>
                            </Stack>

                            {editing ? (
                                <Box sx={{ mt: 1 }}>
                                    <TextField
                                        fullWidth
                                        multiline
                                        minRows={2}
                                        value={editText}
                                        onChange={(e) => setEditText(e.target.value)}
                                        variant="outlined"
                                        sx={{
                                            "& .MuiOutlinedInput-root": {
                                                bgcolor: "background.paper",
                                                borderRadius: 2
                                            }
                                        }}
                                    />
                                    <Stack direction="row" spacing={1} sx={{ mt: 1 }} justifyContent="flex-end">
                                        <Button
                                            size="small"
                                            onClick={() => {
                                                setEditing(false);
                                                setEditText(node.text || "");
                                            }}
                                            sx={{ borderRadius: 2, fontSize: '0.75rem' }}
                                        >
                                            Abbrechen
                                        </Button>
                                        <Button
                                            size="small"
                                            variant="contained"
                                            onClick={() => {
                                                onEdit(String(node.id!), editText);
                                                setEditing(false);
                                            }}
                                            sx={{ borderRadius: 2, fontSize: '0.75rem' }}
                                        >
                                            Speichern
                                        </Button>
                                    </Stack>
                                </Box>
                            ) : (
                                <Typography
                                    variant="body2"
                                    sx={{
                                        whiteSpace: "pre-wrap",
                                        lineHeight: 1.6,
                                        wordBreak: "break-word",
                                        overflowWrap: "anywhere",
                                        fontSize: { xs: '0.8rem', sm: '0.875rem' }
                                    }}
                                >
                                    {renderTextWithMentions(node.text ?? "")}
                                </Typography>
                            )}
                        </Stack>
                    </Paper>

                    {/* Main Content Actions */}
                    <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap" sx={{ px: { xs: 1.5, sm: 2 } }}>
                        <Stack direction="row" alignItems="center" spacing={0}>
                            <IconButton
                                size="small"
                                onClick={() => onVote(String(node.id!), userVote === 1 ? 0 : 1)}
                                color={userVote === 1 ? "primary" : "default"}
                                sx={{ p: { xs: 0.25, sm: 0.5 } }}
                            >
                                <ThumbUpOutlined sx={{ fontSize: { xs: 14, sm: 16 } }} />
                            </IconButton>
                            <Typography variant="caption" sx={{ minWidth: { xs: 16, sm: 20 }, textAlign: 'center', fontWeight: 'bold', fontSize: { xs: '0.7rem', sm: '0.75rem' } }}>
                                {netVotes}
                            </Typography>
                            <IconButton
                                size="small"
                                onClick={() => onVote(String(node.id!), userVote === -1 ? 0 : -1)}
                                color={userVote === -1 ? "primary" : "default"}
                                sx={{ p: { xs: 0.25, sm: 0.5 } }}
                            >
                                <ThumbDownOutlined sx={{ fontSize: { xs: 14, sm: 16 } }} />
                            </IconButton>
                        </Stack>

                        <Button
                            size="small"
                            startIcon={<ReplyIcon sx={{ fontSize: { xs: 14, sm: 16 } }} />}
                            onClick={() => setReplyOpen((v) => !v)}
                            disabled={!canReply}
                            sx={{
                                color: appearance.textSecondary,
                                minWidth: 0,
                                px: { xs: 0.5, sm: 1 },
                                fontSize: { xs: '0.7rem', sm: '0.75rem' },
                                textTransform: 'none',
                                "&:hover": { color: appearance.accent, bgcolor: "transparent" }
                            }}
                        >
                            Antworten
                        </Button>
                        <Tooltip title={copied ? "Kopiert!" : "Link kopieren"}>
                            <Button
                                size="small"
                                startIcon={<ShareIcon sx={{ fontSize: { xs: 14, sm: 16 } }} />}
                                onClick={handleShare}
                                sx={{
                                    color: appearance.textSecondary,
                                    minWidth: 0,
                                    px: { xs: 0.5, sm: 1 },
                                    fontSize: { xs: '0.7rem', sm: '0.75rem' },
                                    textTransform: 'none',
                                    "&:hover": { color: appearance.accent, bgcolor: "transparent" }
                                }}
                            >
                                Teilen
                            </Button>
                        </Tooltip>
                        {!editing && (
                            <>
                                <IconButton
                                    size="small"
                                    onClick={handleOpenMenu}
                                    sx={{
                                        color: appearance.textSecondary,
                                        "&:hover": { color: appearance.accent, bgcolor: "transparent" }
                                    }}
                                >
                                    <MoreHorizIcon sx={{ fontSize: { xs: 16, sm: 20 } }} />
                                </IconButton>
                                <Menu
                                    anchorEl={menuAnchorEl}
                                    open={openMenu}
                                    onClose={handleCloseMenu}
                                    disableScrollLock
                                    PaperProps={{
                                        elevation: 2,
                                        sx: {
                                            borderRadius: 2,
                                            mt: 0.5,
                                            minWidth: 140,
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
                                    {canEdit && (
                                        <MenuItem
                                            onClick={() => {
                                                handleCloseMenu();
                                                setEditing(true);
                                            }}
                                        >
                                            <EditIcon sx={{ fontSize: 18 }} />
                                            <Typography variant="body2" sx={{ fontWeight: 500 }}>Bearbeiten</Typography>
                                        </MenuItem>
                                    )}
                                    <MenuItem disabled onClick={handleCloseMenu}>
                                        <ReportOutlinedIcon sx={{ fontSize: 18 }} />
                                        <Typography variant="body2" sx={{ fontWeight: 500 }}>Melden</Typography>
                                    </MenuItem>
                                </Menu>
                            </>
                        )}
                    </Stack>

                    {replyOpen && (
                        <Box sx={{ mt: 2, ml: { xs: 1.5, sm: 2 } }}>
                            <CommentsSection
                                comments={[]}
                                onAdd={(_, t) => {
                                    onReply(String(node.id ?? null), t);
                                    setReplyOpen(false);
                                }}
                                appearance={appearance}
                                currentUser={currentUser}
                                flat={true}
                                idPrefix={`reply-${node.id}`}
                                disableCollapse
                            />
                        </Box>
                    )}
                </Box>
            </Stack>

            {depth === 0 && totalDescendants > 0 && (
                <Box sx={{ mt: 1, ml: { xs: 3.5, sm: 5 } }}>
                    <Button
                        size="small"
                        onClick={() => setRepliesExpanded(!repliesExpanded)}
                        startIcon={repliesExpanded ? <ExpandLessIcon sx={{ fontSize: { xs: 16, sm: 20 } }} /> : <ExpandMoreIcon sx={{ fontSize: { xs: 16, sm: 20 } }} />}
                        sx={{
                            color: "primary.main",
                            fontWeight: 600,
                            fontSize: { xs: '0.75rem', sm: '0.8125rem' },
                            textTransform: 'none',
                            "&:hover": { bgcolor: "transparent", textDecoration: 'underline' }
                        }}
                    >
                        {totalDescendants} {totalDescendants === 1 ? "Antwort" : "Antworten"}
                    </Button>
                </Box>
            )}

            {(depth > 0 || repliesExpanded) && node.children.map((child) => (
                <CommentThread
                    key={child.id}
                    node={child}
                    onReply={onReply}
                    onEdit={onEdit}
                    onVote={onVote}
                    depth={depth + 1}
                    appearance={appearance}
                    currentUser={currentUser}
                />
            ))}
        </Box>
    );
}

export function CommentsSection({
    comments,
    onAdd,
    onEdit,
    onVote,
    appearance,
    currentUser,
    disabledHelper,
    flat = false,
    disableCollapse = false,
    idPrefix = "comment",
}: {
    comments: Comment[];
    onAdd: (parentId: string | null, text: string) => void;
    onEdit?: (commentId: string, text: string) => void;
    onVote?: (commentId: string, vote: Vote) => void;
    appearance: CommentAppearance;
    currentUser: DtoUserResponse | null;
    disabledHelper?: string;
    flat?: boolean;
    disableCollapse?: boolean;
    idPrefix?: string;
}) {
    const safeComments = React.useMemo(() => (Array.isArray(comments) ? comments : []), [comments]);
    const totalComments = safeComments.length;
    const [expanded, setExpanded] = React.useState(false);
    const collapseLimit = disableCollapse ? Number.MAX_SAFE_INTEGER : COMMENT_COLLAPSE_LIMIT;
    const canComment = !!currentUser;
    const visibleComments = React.useMemo(
        () =>
            expanded || totalComments <= collapseLimit
                ? safeComments
                : safeComments.slice(0, collapseLimit),
        [safeComments, expanded, totalComments, collapseLimit]
    );
    const hiddenCount = disableCollapse ? 0 : totalComments - visibleComments.length;
    const [sortOrder, setSortOrder] = React.useState<"desc" | "top">("desc");

    const sortedComments = React.useMemo(() => {
        const sorted = [...visibleComments];
        if (sortOrder === "top") {
            return sorted.sort((a, b) => (b.votes || 0) - (a.votes || 0));
        }
        return sorted.sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime());
    }, [visibleComments, sortOrder]);

    React.useEffect(() => {
        if (typeof window === 'undefined') return;
        const hash = window.location.hash.replace("#", "");
        if (hash && safeComments.some(c => String(c.id) === hash)) {
            setExpanded(true);
        }
    }, [safeComments]);

    const tree = React.useMemo(() => buildCommentTree(sortedComments), [sortedComments]);
    const [text, setText] = React.useState("");
    const handleAdd = React.useCallback(
        (parentId: string | null, t: string) => {
            if (!expanded && (hiddenCount > 0 || totalComments >= collapseLimit)) {
                setExpanded(true);
            }
            onAdd(parentId, t);
        },
        [expanded, hiddenCount, totalComments, onAdd, collapseLimit]
    );
    const renderNodes = React.useCallback(
        () =>
            tree.map((root) => (
                <CommentThread
                    key={root.id}
                    node={root}
                    onReply={(pid, t) => handleAdd(pid, t)}
                    onEdit={onEdit || (() => { })}
                    onVote={onVote || (() => { })}
                    appearance={appearance}
                    currentUser={currentUser}
                />
            )),
        [tree, handleAdd, appearance, currentUser, onEdit, onVote]
    );

    return (
        <Stack spacing={3}>
            {!flat && (
                <Paper
                    elevation={0}
                    sx={{
                        p: 2,
                        borderRadius: 3,
                        bgcolor: appearance.surface,
                        border: `1px solid ${appearance.border}`,
                    }}
                >
                    <Stack direction="row" spacing={2} alignItems="flex-start">
                        {currentUser && (
                            <Avatar
                                src={getAvatarUrl(currentUser.avatar_url)}
                                sx={{
                                    width: 40, height: 40,
                                    bgcolor: appearance.accent,
                                    fontSize: "1rem",
                                    fontWeight: "bold"
                                }}
                            />
                        )}
                        <Stack spacing={2} sx={{ width: "100%" }}>
                            <TextField
                                id={`${idPrefix}-input-contained`}
                                name="comment"
                                fullWidth
                                multiline
                                minRows={3}
                                placeholder={canComment ? "Kommentar hinzufügen..." : "Bitte einloggen, um zu kommentieren."}
                                value={text}
                                onChange={(e) => setText(e.target.value)}
                                disabled={!canComment}
                                variant="outlined"
                                sx={{
                                    "& .MuiOutlinedInput-root": {
                                        bgcolor: "background.paper",
                                        borderRadius: 2,
                                        "& fieldset": { borderColor: "transparent" },
                                        "&:hover fieldset": { borderColor: "transparent" },
                                        "&.Mui-focused fieldset": { borderColor: appearance.accent },
                                    },
                                }}
                            />
                            <Stack direction="row" justifyContent="end" alignItems="center">
                                <Button
                                    variant="contained"
                                    onClick={() => {
                                        if (!canComment) return;
                                        if (text.trim()) {
                                            handleAdd(null, text.trim());
                                            setText("");
                                        }
                                    }}
                                    disabled={!canComment || !text.trim()}
                                    endIcon={<SendIcon />}
                                    sx={{
                                        borderRadius: 2,
                                        px: 3,
                                        py: 1,
                                        textTransform: "none",
                                        fontWeight: 600,
                                        boxShadow: "none",
                                        "&:hover": { boxShadow: "0 4px 12px rgba(0,0,0,0.15)" }
                                    }}
                                >
                                    Posten
                                </Button>
                            </Stack>
                        </Stack>
                    </Stack>
                    {!canComment && disabledHelper && (
                        <Typography variant="caption" sx={{ color: appearance.textSecondary, mt: 1, display: "block" }}>
                            {disabledHelper}
                        </Typography>
                    )}
                </Paper>
            )}

            {/* Flat input for replies */}
            {flat && (
                <Stack spacing={2}>
                    <TextField
                        id={`${idPrefix}-input-flat`}
                        name="comment"
                        fullWidth
                        multiline
                        minRows={2}
                        placeholder={canComment ? "Antwort schreiben..." : "Bitte einloggen."}
                        value={text}
                        onChange={(e) => setText(e.target.value)}
                        disabled={!canComment}
                        variant="outlined"
                        sx={{
                            "& .MuiOutlinedInput-root": {
                                bgcolor: "background.paper",
                                borderRadius: 2,
                            },
                        }}
                    />
                    <Stack direction="row" justifyContent="flex-end">
                        <Button
                            variant="contained"
                            size="small"
                            onClick={() => {
                                if (!canComment) return;
                                if (text.trim()) {
                                    handleAdd(null, text.trim());
                                    setText("");
                                }
                            }}
                            disabled={!canComment || !text.trim()}
                            endIcon={<SendIcon />}
                            sx={{ borderRadius: 2, textTransform: "none" }}
                        >
                            Antworten
                        </Button>
                    </Stack>
                </Stack>
            )}

            {/* List */}
            <Box>
                {!flat && (
                    <Stack
                        direction={{ xs: "column", sm: "row" }}
                        justifyContent="space-between"
                        alignItems={{ xs: "flex-start", sm: "center" }}
                        spacing={2}
                        mb={3}
                    >
                        <Typography variant="h6" fontWeight={700}>
                            Kommentare ({totalComments})
                        </Typography>
                        <TextField
                            select
                            size="small"
                            value={sortOrder}
                            onChange={(e) => {
                                const value = e.target.value;
                                if (value === "top" || value === "desc") {
                                    setSortOrder(value);
                                }
                            }}
                            InputProps={{
                                sx: { borderRadius: 2, fontSize: '0.875rem' }
                            }}
                            sx={{ minWidth: 120, alignSelf: { xs: "flex-end", sm: "center" } }}
                        >
                            <MenuItem value="desc">Neueste</MenuItem>
                            <MenuItem value="top">Am besten bewertet</MenuItem>
                        </TextField>
                    </Stack>
                )}

                {tree.length ? (
                    <Stack spacing={3}>{renderNodes()}</Stack>
                ) : !flat ? (
                    <Paper
                        variant="outlined"
                        sx={{
                            p: 4,
                            textAlign: "center",
                            borderStyle: "dashed",
                            borderColor: appearance.border,
                            bgcolor: "transparent",
                            borderRadius: 3
                        }}
                    >
                        <Typography variant="body1" sx={{ color: appearance.textSecondary }}>
                            Für diesen Beitrag gibt es noch keine Kommentare.
                        </Typography>
                    </Paper>
                ) : null}

                {hiddenCount > 0 && !expanded && (
                    <Box sx={{ mt: 2, textAlign: "center" }}>
                        <Button onClick={() => setExpanded(true)} sx={{ textTransform: "none" }}>
                            Alle {totalComments} Kommentare anzeigen
                        </Button>
                    </Box>
                )}
            </Box>
        </Stack>
    );
}
