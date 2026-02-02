import * as React from "react";
import { Link as RouterLink } from "react-router-dom";
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
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import { getAvatarUrl } from "@lib/images";
import type { DtoDiscussionPostResponse as ApiPost, DtoDiscussionCommentResponse as ApiComment, DtoUserResponse } from "@lib/api";

export type Program = string;
export type Vote = -1 | 0 | 1;
export type Comment = ApiComment;
export type Post = ApiPost & { comments: Comment[]; };
export type CommentAppearance = { surface: string; border: string; accent: string; textSecondary: string; };
export const COMMENT_COLLAPSE_LIMIT = 6;
export const POSTS_PER_PAGE = 10;
export { FORUM_CATEGORIES, FORUM_TAGS } from "@internals/data";

export const isoToShort = (iso?: string) => {
    if (!iso) return "Unbekannt";
    try {
        const d = new Date(iso);
        if (isNaN(d.getTime())) return "Unbekannt";
        return d.toLocaleString("de-DE", { year: "numeric", month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
    } catch { return "Unbekannt"; }
}

export type CommentNode = Comment & { children: CommentNode[] };
export function buildCommentTree(comments: Comment[]): CommentNode[] {
    if (!Array.isArray(comments)) return [];
    const map = new Map<string, CommentNode>();
    const roots: CommentNode[] = [];
    comments.forEach((c) => { if (c.id) map.set(String(c.id), { ...c, children: [] }); });
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
        if (mIndex > lastIndex) nodes.push(text.slice(lastIndex, mIndex) as React.ReactNode);
        const mention = match[0];
        nodes.push(<Box component="span" key={`${mention}-${mIndex}`} sx={{ color: "primary.light", fontWeight: 600, cursor: "pointer", "&:hover": { textDecoration: "underline", color: "primary.main" } }}>{mention}</Box>);
        lastIndex = mIndex + mention.length;
    }
    if (lastIndex < text.length) nodes.push(text.slice(lastIndex) as React.ReactNode);
    return nodes.length ? nodes : [text as React.ReactNode];
}

export function CommentThread({ node, onReply, onEdit, onVote, depth = 0, appearance, currentUser }: { node: CommentNode; onReply: (parentId: string | null, text: string) => void; onEdit: (commentId: string, text: string) => void; onVote: (commentId: string, vote: Vote) => void; depth?: number; appearance: CommentAppearance; currentUser: DtoUserResponse | null; }) {
    const [replyOpen, setReplyOpen] = React.useState(false);
    const [editing, setEditing] = React.useState(false);
    const [repliesExpanded, setRepliesExpanded] = React.useState(false);
    const [editText, setEditText] = React.useState(node.text || "");
    const [copied, setCopied] = React.useState(false);
    const [menuAnchorEl, setMenuAnchorEl] = React.useState<null | HTMLElement>(null);

    return (
        <Box id={String(node.id)} sx={{ mt: 1.5, pl: depth ? { xs: 0.75, sm: 2 } : 0, borderLeft: depth ? `1.5px solid ${appearance.border}` : "none", width: "100%", maxWidth: "100%", boxSizing: "border-box" }}>
            <Stack direction="row" spacing={{ xs: 1, sm: 2 }} alignItems="flex-start">
                <Avatar src={getAvatarUrl(node.user_avatar_url)} sx={{ width: depth ? { xs: 20, sm: 24 } : { xs: 28, sm: 32 }, height: depth ? { xs: 20, sm: 24 } : { xs: 28, sm: 32 }, bgcolor: appearance.accent, fontSize: depth ? "0.6rem" : "0.8rem", fontWeight: "bold", flexShrink: 0 }}>{node.user_name?.[0]?.toUpperCase() || 'A'}</Avatar>
                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Paper elevation={0} sx={{ p: { xs: 1.5, sm: 2 }, borderRadius: 3, bgcolor: appearance.surface, border: `1px solid ${appearance.border}`, width: "100%", boxSizing: "border-box", overflow: "hidden" }}>
                        <Stack spacing={1}>
                            <Stack direction="row" spacing={1} alignItems="center" sx={{ minWidth: 0 }}>
                                <Typography component={RouterLink} to={`/u/${node.user_id}`} variant="subtitle2" fontWeight={700} noWrap sx={{ color: "inherit", textDecoration: "none", fontSize: { xs: '0.75rem', sm: '0.875rem' }, "&:hover": { color: appearance.accent } }}>{node.user_name || "Anonym"}</Typography>
                                <Typography variant="caption" sx={{ color: appearance.textSecondary, whiteSpace: 'nowrap', fontSize: { xs: '0.7rem', sm: '0.75rem' } }}> · {isoToShort(node.created_at ?? "")}</Typography>
                            </Stack>
                            {editing ? <Box sx={{ mt: 1 }}>
                                <TextField fullWidth multiline minRows={2} value={editText} onChange={(e) => setEditText(e.target.value)} variant="outlined" sx={{ "& .MuiOutlinedInput-root": { bgcolor: "background.paper", borderRadius: 2 } }} />
                                <Stack direction="row" spacing={1} sx={{ mt: 1 }} justifyContent="flex-end">
                                    <Button size="small" onClick={() => { setEditing(false); setEditText(node.text || ""); }}>Abbrechen</Button>
                                    <Button size="small" variant="contained" onClick={() => { onEdit(String(node.id!), editText); setEditing(false); }}>Speichern</Button>
                                </Stack>
                            </Box> : <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", lineHeight: 1.6, wordBreak: "break-word", overflowWrap: "anywhere", fontSize: { xs: '0.8rem', sm: '0.875rem' } }}>{renderTextWithMentions(node.text ?? "")}</Typography>}
                        </Stack>
                    </Paper>
                    <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap" sx={{ px: { xs: 1.5, sm: 2 } }}>
                        <Stack direction="row" alignItems="center">
                            <IconButton size="small" onClick={() => onVote(String(node.id!), node.user_vote === 1 ? 0 : 1)} color={node.user_vote === 1 ? "primary" : "default"}><ThumbUpOutlined sx={{ fontSize: 16 }} /></IconButton>
                            <Typography variant="caption" sx={{ minWidth: 20, textAlign: 'center', fontWeight: 'bold' }}>{node.votes}</Typography>
                            <IconButton size="small" onClick={() => onVote(String(node.id!), node.user_vote === -1 ? 0 : -1)} color={node.user_vote === -1 ? "primary" : "default"}><ThumbDownOutlined sx={{ fontSize: 16 }} /></IconButton>
                        </Stack>
                        <Button size="small" startIcon={<ReplyIcon sx={{ fontSize: 16 }} />} onClick={() => setReplyOpen(!replyOpen)} disabled={!currentUser} sx={{ color: appearance.textSecondary, textTransform: 'none' }}>Antworten</Button>
                        <Tooltip title={copied ? "Kopiert!" : "Teilen"}><Button size="small" startIcon={<ShareIcon sx={{ fontSize: 16 }} />} onClick={() => { navigator.clipboard.writeText(`${window.location.origin}${window.location.pathname}#${node.id}`).then(() => { setCopied(true); setTimeout(() => setCopied(false), 2000); }); }} sx={{ color: appearance.textSecondary, textTransform: 'none' }}>Teilen</Button></Tooltip>
                        {(currentUser?.role === 'admin' || String(currentUser?.id) === String(node.user_id)) && !editing && <IconButton size="small" onClick={(e) => setMenuAnchorEl(e.currentTarget)}><MoreHorizIcon sx={{ fontSize: 20 }} /></IconButton>}
                        <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={() => setMenuAnchorEl(null)}>
                            <MenuItem onClick={() => { setMenuAnchorEl(null); setEditing(true); }}><EditIcon sx={{ fontSize: 18, mr: 1 }} /> Bearbeiten</MenuItem>
                        </Menu>
                    </Stack>
                    {replyOpen && <Box sx={{ mt: 2, ml: 2 }}><CommentsSection comments={[]} onAdd={(_, t) => { onReply(String(node.id!), t); setReplyOpen(false); }} appearance={appearance} currentUser={currentUser} flat disableCollapse /></Box>}
                </Box>
            </Stack>
            {depth === 0 && node.children.length > 0 && <Box sx={{ mt: 1, ml: 5 }}><Button size="small" onClick={() => setRepliesExpanded(!repliesExpanded)} startIcon={repliesExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />} sx={{ textTransform: 'none' }}>{node.children.length} Antworten</Button></Box>}
            {(depth > 0 || repliesExpanded) && node.children.map(c => <CommentThread key={c.id} node={c} onReply={onReply} onEdit={onEdit} onVote={onVote} depth={depth + 1} appearance={appearance} currentUser={currentUser} />)}
        </Box>
    );
}

export function CommentsSection({ comments, onAdd, onEdit, onVote, appearance, currentUser, flat = false, disableCollapse = false }: { comments: Comment[]; onAdd: (parentId: string | null, text: string) => void; onEdit?: (commentId: string, text: string) => void; onVote?: (commentId: string, vote: Vote) => void; appearance: CommentAppearance; currentUser: DtoUserResponse | null; flat?: boolean; disableCollapse?: boolean; }) {
    const [expanded, setExpanded] = React.useState(false);
    const [text, setText] = React.useState("");
    const visible = expanded || disableCollapse || comments.length <= COMMENT_COLLAPSE_LIMIT ? comments : comments.slice(0, COMMENT_COLLAPSE_LIMIT);
    const tree = React.useMemo(() => buildCommentTree(visible), [visible]);

    return (
        <Stack spacing={3}>
            {!flat && <Paper elevation={0} sx={{ p: 2, borderRadius: 3, bgcolor: appearance.surface, border: `1px solid ${appearance.border}` }}>
                <Stack spacing={2}>
                    <TextField fullWidth multiline minRows={3} placeholder={currentUser ? "Kommentar hinzufügen..." : "Bitte einloggen"} value={text} onChange={e => setText(e.target.value)} disabled={!currentUser} />
                    <Stack direction="row" justifyContent="end"><Button variant="contained" onClick={() => { onAdd(null, text); setText(""); }} disabled={!text.trim()} endIcon={<SendIcon />}>Posten</Button></Stack>
                </Stack>
            </Paper>}
            {flat && <Stack spacing={2}>
                <TextField fullWidth multiline minRows={2} placeholder="Antwort schreiben..." value={text} onChange={e => setText(e.target.value)} />
                <Stack direction="row" justifyContent="end"><Button variant="contained" size="small" onClick={() => { onAdd(null, text); setText(""); }} disabled={!text.trim()}>Antworten</Button></Stack>
            </Stack>}
            <Box>
                {!flat && <Typography variant="h6" fontWeight={700} sx={{ mb: 2 }}>Kommentare ({comments.length})</Typography>}
                <Stack spacing={2}>{tree.map(root => <CommentThread key={root.id} node={root} onReply={onAdd} onEdit={onEdit || (() => { })} onVote={onVote || (() => { })} appearance={appearance} currentUser={currentUser} />)}</Stack>
                {comments.length > visible.length && <Button onClick={() => setExpanded(true)} sx={{ mt: 2, width: '100%' }}>Alle Kommentare anzeigen</Button>}
            </Box>
        </Stack>
    );
}
