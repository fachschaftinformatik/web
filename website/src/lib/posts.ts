import type { DtoDiscussionPostResponse as ApiPost, DtoDiscussionCommentResponse as ApiComment } from "@lib/api";

export type Vote = -1 | 0 | 1;
export type Comment = ApiComment;
export type Post = ApiPost & { comments?: Comment[]; };

export const isoToShort = (iso?: string) => {
    if (!iso) return "Unbekannt";
    try {
        const d = new Date(iso);
        if (isNaN(d.getTime())) return "Unbekannt";
        return d.toLocaleString("de-DE", { year: "numeric", month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
    } catch { return "Unbekannt"; }
}
