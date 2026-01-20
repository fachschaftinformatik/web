-- name: CreateForumPost :one
INSERT INTO forum_posts (
    id, title, body, author_id, type, pinned, event_date, location, programs, tags, image_url, links
) VALUES (
    sqlc.arg(id), sqlc.arg(title), sqlc.arg(body), sqlc.arg(author_id),
    COALESCE(sqlc.arg(type), 'forum'),
    COALESCE(sqlc.arg(pinned), 0),
    sqlc.arg(event_date),
    sqlc.arg(location),
    COALESCE(sqlc.arg(programs), '[]'),
    COALESCE(sqlc.arg(tags), '[]'),
    sqlc.arg(image_url),
    COALESCE(sqlc.arg(links), '[]')
)
RETURNING *;

-- name: GetForumPost :one
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as author_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as author_avatar_url,
       (SELECT COUNT(*) FROM forum_comments WHERE post_id = p.id AND active = 1) as comment_count,
       (SELECT CAST(COALESCE(SUM(vote), 0) AS INTEGER) FROM forum_votes WHERE forum_votes.post_id = p.id) as votes,
       CAST(COALESCE((SELECT vote FROM forum_votes WHERE forum_votes.post_id = p.id AND forum_votes.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM forum_posts p
JOIN users u ON p.author_id = u.id
WHERE p.id = sqlc.arg(id)
LIMIT 1;

-- name: ListForumPosts :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as author_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as author_avatar_url,
       (SELECT COUNT(*) FROM forum_comments WHERE post_id = p.id AND active = 1) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM forum_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM forum_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM forum_posts p
JOIN users u ON p.author_id = u.id
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.active = 1
ORDER BY
    p.pinned DESC,
    p.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListForumPostsTop :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as author_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as author_avatar_url,
       (SELECT COUNT(*) FROM forum_comments WHERE post_id = p.id AND active = 1) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM forum_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM forum_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM forum_posts p
JOIN users u ON p.author_id = u.id
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.active = 1
ORDER BY
    p.pinned DESC,
    votes DESC,
    p.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateForumPost :one
UPDATE forum_posts
SET title = COALESCE(sqlc.narg(title), title),
    body = COALESCE(sqlc.narg(body), body),
    pinned = COALESCE(sqlc.narg(pinned), pinned),
    event_date = COALESCE(sqlc.narg(event_date), event_date),
    location = COALESCE(sqlc.narg(location), location),
    image_url = COALESCE(sqlc.narg(image_url), image_url),
    links = COALESCE(sqlc.narg(links), links),
    programs = COALESCE(sqlc.narg(programs), programs),
    tags = COALESCE(sqlc.narg(tags), tags),
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteForumPost :exec
UPDATE forum_posts
SET active = 0
WHERE id = sqlc.arg(id);

-- name: CreateForumComment :one
INSERT INTO forum_comments (
    id, post_id, author_id, parent_id, text
) VALUES (
    sqlc.arg(id), sqlc.arg(post_id), sqlc.arg(author_id), sqlc.arg(parent_id), sqlc.arg(text)
)
RETURNING *;

-- name: GetForumComment :one
SELECT * FROM forum_comments WHERE id = sqlc.arg(id) LIMIT 1;

-- name: ListForumComments :many
SELECT c.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as author_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as author_avatar_url,
       (SELECT CAST(COALESCE(SUM(vote), 0) AS INTEGER) FROM forum_comment_votes v WHERE v.comment_id = c.id) as votes,
       CAST(COALESCE((SELECT vote FROM forum_comment_votes v WHERE v.comment_id = c.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM forum_comments c
JOIN users u ON c.author_id = u.id
WHERE c.post_id = sqlc.arg(post_id) AND c.active = 1
ORDER BY c.created_at ASC;

-- name: SearchForumPosts :many
SELECT p.*,
       CAST(CASE WHEN u.private = 1 THEN 'Anonym' ELSE u.name END AS TEXT) as author_name,
       CAST(CASE WHEN u.private = 1 THEN '' ELSE COALESCE(u.avatar_url, '') END AS TEXT) as author_avatar_url,
       (SELECT COUNT(*) FROM forum_comments WHERE post_id = p.id AND active = 1) as comment_count,
       (SELECT CAST(COALESCE(SUM(v.vote), 0) AS INTEGER) FROM forum_votes v WHERE v.post_id = p.id) as votes,
       CAST(COALESCE((SELECT v.vote FROM forum_votes v WHERE v.post_id = p.id AND v.user_id = sqlc.narg(current_user_id)), 0) AS INTEGER) as user_vote
FROM forum_posts p
JOIN users u ON p.author_id = u.id
WHERE (
  lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR
  lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'
) AND p.active = 1
ORDER BY 
    p.pinned DESC,
    p.created_at DESC
LIMIT 10;

-- name: CountForumPosts :one
SELECT COUNT(*) FROM forum_posts p
WHERE (sqlc.narg(type) IS NULL OR p.type = sqlc.narg(type))
  AND (sqlc.narg(query) IS NULL OR (lower(p.title) LIKE '%' || lower(sqlc.arg(query)) || '%' OR lower(p.body) LIKE '%' || lower(sqlc.arg(query)) || '%'))
  AND p.active = 1;

-- name: UpdateForumComment :one
UPDATE forum_comments
SET text = sqlc.arg(text),
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteForumComment :exec
UPDATE forum_comments
SET active = 0
WHERE id = sqlc.arg(id);

-- name: UpsertForumVote :one
INSERT INTO forum_votes (
    post_id, user_id, vote
) VALUES (
    sqlc.arg(post_id), sqlc.arg(user_id), sqlc.arg(vote)
)
ON CONFLICT (post_id, user_id) DO UPDATE SET
    vote = excluded.vote
RETURNING *;

-- name: UpsertForumCommentVote :one
INSERT INTO forum_comment_votes (
    comment_id, user_id, vote
) VALUES (
    sqlc.arg(comment_id), sqlc.arg(user_id), sqlc.arg(vote)
)
ON CONFLICT (comment_id, user_id) DO UPDATE SET
    vote = excluded.vote
RETURNING *;
