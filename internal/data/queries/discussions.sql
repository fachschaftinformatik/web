-- name: CreateDiscussionPost :one
INSERT INTO discussion_posts (
    id, title, body, user_id, type, pinned, event_date, location, image_url
) VALUES (
    sqlc.arg(id), sqlc.arg(title), sqlc.arg(body), sqlc.arg(user_id),
    COALESCE(sqlc.arg(type), 'discussion'),
    COALESCE(sqlc.arg(pinned), 0),
    sqlc.arg(event_date),
    sqlc.arg(location),
    sqlc.arg(image_url)
)
RETURNING *;

-- name: AddProgramToPost :exec
INSERT INTO discussion_post_programs (post_id, program_id) VALUES (sqlc.arg(post_id), sqlc.arg(program_id));

-- name: AddTagToPost :exec
INSERT INTO discussion_post_tags (post_id, tag) VALUES (sqlc.arg(post_id), sqlc.arg(tag));

-- name: AddLinkToPost :one
INSERT INTO discussion_post_links (post_id, url, label) VALUES (sqlc.arg(post_id), sqlc.arg(url), sqlc.arg(label)) RETURNING *;

-- name: GetDiscussionPost :one
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as user_avatar_url,
       (SELECT COUNT(*) FROM discussion_comments WHERE post_id = p.id AND deleted_at IS NULL) as comment_count,
       (SELECT CAST(COALESCE(SUM(vote), 0) AS INTEGER) FROM discussion_votes WHERE discussion_votes.post_id = p.id) as votes,
       CAST(COALESCE((SELECT vote FROM discussion_votes WHERE discussion_votes.post_id = p.id AND discussion_votes.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM discussion_posts p
JOIN users u ON p.user_id = u.id
WHERE p.id = sqlc.arg(id) AND p.deleted_at IS NULL
LIMIT 1;

-- name: GetPostPrograms :many
SELECT p.id, p.name FROM programs p
JOIN discussion_post_programs dpp ON p.id = dpp.program_id
WHERE dpp.post_id = sqlc.arg(post_id);

-- name: GetPostTags :many
SELECT tag FROM discussion_post_tags WHERE post_id = sqlc.arg(post_id);

-- name: GetPostLinks :many
SELECT * FROM discussion_post_links WHERE post_id = sqlc.arg(post_id);

-- name: ListDiscussionPosts :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as user_avatar_url,
       (SELECT COUNT(*) FROM discussion_comments WHERE post_id = p.id AND deleted_at IS NULL) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM discussion_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM discussion_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM discussion_posts p
JOIN users u ON p.user_id = u.id
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.deleted_at IS NULL
  AND (sqlc.narg(program_id) IS NULL OR p.id IN (SELECT post_id FROM discussion_post_programs WHERE program_id = sqlc.narg(program_id)))
ORDER BY
    p.pinned DESC,
    p.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListDiscussionPostsTop :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as user_avatar_url,
       (SELECT COUNT(*) FROM discussion_comments WHERE post_id = p.id AND deleted_at IS NULL) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM discussion_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM discussion_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM discussion_posts p
JOIN users u ON p.user_id = u.id
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.deleted_at IS NULL
  AND (sqlc.narg(program_id) IS NULL OR p.id IN (SELECT post_id FROM discussion_post_programs WHERE program_id = sqlc.narg(program_id)))
ORDER BY
    p.pinned DESC,
    votes DESC,
    p.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateDiscussionPost :one
UPDATE discussion_posts
SET title = COALESCE(sqlc.narg(title), title),
    body = COALESCE(sqlc.narg(body), body),
    pinned = COALESCE(sqlc.narg(pinned), pinned),
    event_date = COALESCE(sqlc.narg(event_date), event_date),
    location = COALESCE(sqlc.narg(location), location),
    image_url = COALESCE(sqlc.narg(image_url), image_url)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: ClearPostPrograms :exec
DELETE FROM discussion_post_programs WHERE post_id = sqlc.arg(post_id);

-- name: ClearPostTags :exec
DELETE FROM discussion_post_tags WHERE post_id = sqlc.arg(post_id);

-- name: ClearPostLinks :exec
DELETE FROM discussion_post_links WHERE post_id = sqlc.arg(post_id);

-- name: DeleteDiscussionPost :exec
UPDATE discussion_posts
SET deleted_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id);

-- name: CreateDiscussionComment :one
INSERT INTO discussion_comments (
    id, post_id, user_id, parent_id, text
) VALUES (
    sqlc.arg(id), sqlc.arg(post_id), sqlc.arg(user_id), sqlc.arg(parent_id), sqlc.arg(text)
)
RETURNING *;

-- name: GetDiscussionComment :one
SELECT * FROM discussion_comments WHERE id = sqlc.arg(id) AND deleted_at IS NULL LIMIT 1;

-- name: ListDiscussionComments :many
SELECT c.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as user_avatar_url,
       (SELECT CAST(COALESCE(SUM(vote), 0) AS INTEGER) FROM discussion_comment_votes v WHERE v.comment_id = c.id) as votes,
       CAST(COALESCE((SELECT vote FROM discussion_comment_votes v WHERE v.comment_id = c.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM discussion_comments c
JOIN users u ON c.user_id = u.id
WHERE c.post_id = sqlc.arg(post_id) AND c.deleted_at IS NULL
ORDER BY c.created_at ASC;

-- name: SearchDiscussionPosts :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as user_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as user_avatar_url,
       (SELECT COUNT(*) FROM discussion_comments WHERE post_id = p.id AND deleted_at IS NULL) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM discussion_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM discussion_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM discussion_posts p
JOIN users u ON p.user_id = u.id
WHERE (
  lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'
) AND p.deleted_at IS NULL
ORDER BY 
    p.pinned DESC,
    p.created_at DESC
LIMIT 10;

-- name: CountDiscussionPosts :one
SELECT COUNT(*) FROM discussion_posts p
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.deleted_at IS NULL
  AND (sqlc.narg(program_id) IS NULL OR p.id IN (SELECT post_id FROM discussion_post_programs WHERE program_id = sqlc.narg(program_id)));

-- name: UpdateDiscussionComment :one
UPDATE discussion_comments
SET text = sqlc.arg(text)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: DeleteDiscussionComment :exec
UPDATE discussion_comments
SET deleted_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id);

-- name: UpsertDiscussionVote :one
INSERT INTO discussion_votes (
    post_id, user_id, vote
) VALUES (
    sqlc.arg(post_id), sqlc.arg(user_id), sqlc.arg(vote)
)
ON CONFLICT (post_id, user_id) DO UPDATE SET
    vote = excluded.vote
RETURNING *;

-- name: UpsertDiscussionCommentVote :one
INSERT INTO discussion_comment_votes (
    comment_id, user_id, vote
) VALUES (
    sqlc.arg(comment_id), sqlc.arg(user_id), sqlc.arg(vote)
)
ON CONFLICT (comment_id, user_id) DO UPDATE SET
    vote = excluded.vote
RETURNING *;
