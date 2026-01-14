import * as React from "react";
import {
    Box, Stack, Paper, Typography, TextField, Button, Avatar, IconButton
} from "@mui/material";
import SendIcon from "@mui/icons-material/Send";
import EditIcon from "@mui/icons-material/Edit";
import ReplyIcon from "@mui/icons-material/Reply";

import type {
    AuthPostResponse as ApiPost,
    AuthCommentResponse as ApiComment,
    AuthUserResponse
} from "@lib/api";
import { Link } from "react-router-dom";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import MenuItem from "@mui/material/MenuItem";

export const PROGRAM_CATALOG = [
    { id: "inf-bsc", label: "Informatik (B.Sc.)", shortLabel: "INF B.Sc.", level: "Bachelor" },
    { id: "winf-bsc", label: "Wirtschaftsinformatik (B.Sc.)", shortLabel: "WINF B.Sc.", level: "Bachelor" },
    { id: "med-bsc", label: "Informatik und Design (B.Sc.)", shortLabel: "IuD B.Sc.", level: "Bachelor" },
    { id: "inf-msc", label: "Informatik (M.Sc.)", shortLabel: "INF M.Sc.", level: "Master" },
    { id: "winf-msc", label: "Wirtschaftsinformatik (M.Sc.)", shortLabel: "WINF M.Sc.", level: "Master" },
    { id: "med-msc", label: "Informatik und Design (M.Sc.)", shortLabel: "IuD M.Sc.", level: "Master" },
] as const;

export type ProgramMeta = typeof PROGRAM_CATALOG[number];
export type Program = ProgramMeta["id"];
export type Vote = -1 | 0 | 1;

export type Comment = ApiComment & {
    author_name?: string;
    votes?: number;
    user_vote?: number;
};
export type Post = Omit<ApiPost, "programs" | "tags" | "links"> & {
    programs: Program[];
    tags: string[];
    links: string[];
    comments: Comment[];
};

export type CommentAppearance = {
    surface: string;
    border: string;
    accent: string;
    textSecondary: string;
};

export const PROGRAM_META_MAP: Record<Program, ProgramMeta> = PROGRAM_CATALOG.reduce((acc, meta) => {
    acc[meta.id] = meta;
    return acc;
}, {} as Record<Program, ProgramMeta>);

export const COMMENT_COLLAPSE_LIMIT = 6;
export const POSTS_PER_PAGE = 20;

export const FORUM_CATEGORIES = [
    "Frage",
    "Diskussion",
    "Lerngruppe",
    "Feedback",
    "Sonstiges"
] as const;

export const FORUM_TAGS = [
    "Hilfe benötigt",
    "Gelöst",
    "Wichtig",
    "Klausur",
    "Mathe",
    "Programmieren",
    "Hardware",
    "Empfehlung",
    "Organisatorisches"
] as const;

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
export function buildCommentTree(comments: Comment[]): CommentNode[] {
    if (!Array.isArray(comments)) return [];
    const map = new Map<string, CommentNode>();
    const roots: CommentNode[] = [];
    comments.forEach((c) => {
        if (c.id) map.set(c.id, { ...c, children: [] });
    });
    comments.forEach((c) => {
        if (!c.id) return;
        const node = map.get(c.id)!;
        const pid = c.parent_id ?? null;
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
            nodes.push(text.slice(lastIndex, mIndex) as React.ReactNode);
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
        nodes.push(text.slice(lastIndex) as React.ReactNode);
    }
    return nodes.length ? nodes : [text as React.ReactNode];
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
    currentUser: AuthUserResponse | null;
}) {
    const [replyOpen, setReplyOpen] = React.useState(false);
    const [editing, setEditing] = React.useState(false);
    const [repliesExpanded, setRepliesExpanded] = React.useState(false);
    const [editText, setEditText] = React.useState(node.text || "");

    const isAuthor = currentUser?.id === node.author_id;
    const canReply = !!currentUser;
    const canEdit = isAuthor || currentUser?.role === "admin" || currentUser?.role === "editor";
    const isEdited = node.updated_at && node.created_at && node.updated_at !== node.created_at;

    const netVotes = node.votes || 0;
    const userVote = (node.user_vote || 0) as Vote;

    return (
        <Box
            sx={{
                mt: 2,
                pl: depth ? 2 : 0,
                borderLeft: depth ? `2px solid ${appearance.border}` : "none",
            }}
        >
            <Stack direction="row" spacing={2} alignItems="flex-start">
                <Avatar
                    src={node.author_avatar_url || undefined}
                    sx={{
                        width: depth ? 24 : 32,
                        height: depth ? 24 : 32,
                        bgcolor: appearance.accent,
                        fontSize: "0.8rem",
                        fontWeight: "bold"
                    }}
                />

                <Box sx={{ width: "100%" }}>
                    <Paper
                        elevation={0}
                        sx={{
                            p: 2,
                            borderRadius: 3,
                            borderRadiusTopLeft: depth ? 3 : 0,
                            bgcolor: appearance.surface,
                            border: `1px solid ${appearance.border}`,
                        }}
                    >
                        <Stack spacing={1}>
                            <Stack direction="row" justifyContent="space-between" alignItems="center">
                                <Stack direction="row" spacing={1} alignItems="center">
                                    <Typography
                                        component={Link}
                                        to={`/user/${node.author_id}`}
                                        variant="subtitle2"
                                        fontWeight={700}
                                        sx={{
                                            color: "inherit",
                                            textDecoration: "none",
                                            "&:hover": { color: appearance.accent, cursor: "pointer" }
                                        }}
                                    >
                                        {node.author_name || "Anonym"}
                                    </Typography>
                                    <Typography variant="caption" sx={{ color: appearance.textSecondary }}>
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
                                            sx={{ borderRadius: 2 }}
                                        >
                                            Abbrechen
                                        </Button>
                                        <Button
                                            size="small"
                                            variant="contained"
                                            onClick={() => {
                                                onEdit(node.id!, editText);
                                                setEditing(false);
                                            }}
                                            sx={{ borderRadius: 2 }}
                                        >
                                            Speichern
                                        </Button>
                                    </Stack>
                                </Box>
                            ) : (
                                <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", lineHeight: 1.6 }}>
                                    {renderTextWithMentions(node.text ?? "")}
                                </Typography>
                            )}
                        </Stack>
                    </Paper>

                    {/* Action Bar */}
                    <Stack direction="row" spacing={1} sx={{ mt: 0.5, ml: 1 }} alignItems="center">
                        <Stack direction="row" alignItems="center" spacing={0}>
                            <IconButton
                                size="small"
                                onClick={() => onVote(node.id!, userVote === 1 ? 0 : 1)}
                                color={userVote === 1 ? "primary" : "default"}
                                sx={{ p: 0.5 }}
                            >
                                <ThumbUpOutlined sx={{ fontSize: 16 }} />
                            </IconButton>
                            <Typography variant="caption" sx={{ minWidth: 20, textAlign: 'center', fontWeight: 'bold' }}>
                                {netVotes}
                            </Typography>
                            <IconButton
                                size="small"
                                onClick={() => onVote(node.id!, userVote === -1 ? 0 : -1)}
                                color={userVote === -1 ? "primary" : "default"}
                                sx={{ p: 0.5 }}
                            >
                                <ThumbDownOutlined sx={{ fontSize: 16 }} />
                            </IconButton>
                        </Stack>

                        <Button
                            size="small"
                            startIcon={<ReplyIcon sx={{ fontSize: 16 }} />}
                            onClick={() => setReplyOpen((v) => !v)}
                            disabled={!canReply}
                            sx={{
                                color: appearance.textSecondary,
                                minWidth: 0,
                                px: 1,
                                textTransform: 'none',
                                "&:hover": { color: appearance.accent, bgcolor: "transparent" }
                            }}
                        >
                            Antworten
                        </Button>
                        {canEdit && !editing && (
                            <Button
                                size="small"
                                startIcon={<EditIcon sx={{ fontSize: 16 }} />}
                                onClick={() => setEditing(true)}
                                sx={{
                                    color: appearance.textSecondary,
                                    minWidth: 0,
                                    px: 1,
                                    textTransform: 'none',
                                    "&:hover": { color: appearance.accent, bgcolor: "transparent" }
                                }}
                            >
                                Bearbeiten
                            </Button>
                        )}
                    </Stack>

                    {replyOpen && (
                        <Box sx={{ mt: 2, ml: 0 }}>
                            <CommentsSection
                                comments={[]} // Reuse input component logic basically
                                onAdd={(_, t) => {
                                    onReply((node.id ?? null) as string | null, t);
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

            {node.children.length > 0 && (
                <Box sx={{ mt: 1, ml: 1 }}>
                    <Button
                        size="small"
                        onClick={() => setRepliesExpanded(!repliesExpanded)}
                        startIcon={repliesExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                        sx={{
                            color: "primary.main",
                            fontWeight: 600,
                            textTransform: 'none',
                            "&:hover": { bgcolor: "transparent", textDecoration: 'underline' }
                        }}
                    >
                        {node.children.length} {node.children.length === 1 ? "Antwort" : "Antworten"}
                    </Button>
                </Box>
            )}

            {repliesExpanded && node.children.map((child) => (
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
    currentUser: AuthUserResponse | null;
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
        // Default to desc (newest)
        return sorted.sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime());
    }, [visibleComments, sortOrder]);

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
            {/* Input Area */}
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
                                src={currentUser.avatar_url || undefined}
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
                    <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
                        <Typography variant="h6" fontWeight={700}>
                            Kommentare ({totalComments})
                        </Typography>
                        <TextField
                            select
                            size="small"
                            value={sortOrder}
                            onChange={(e) => setSortOrder(e.target.value as "top" | "desc")}
                            InputProps={{
                                sx: { borderRadius: 2, fontSize: '0.875rem' }
                            }}
                            sx={{ minWidth: 120 }}
                        >
                            <MenuItem value="desc">Neueste</MenuItem>
                            <MenuItem value="top">Top (am besten bewertet)</MenuItem>
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
