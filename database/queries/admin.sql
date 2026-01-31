-- name: GetAdminStats :one
SELECT
    CAST((SELECT COUNT(*) FROM users WHERE deleted_at IS NULL) AS INTEGER) as user_count,
    CAST(((SELECT COUNT(*) FROM discussion_posts WHERE deleted_at IS NULL) + (SELECT COUNT(*) FROM discussion_comments WHERE deleted_at IS NULL)) AS INTEGER) as post_count,
    CAST((SELECT COUNT(*) FROM archive WHERE deleted_at IS NULL) AS INTEGER) as archive_count,
    CAST((SELECT COUNT(*) FROM events WHERE deleted_at IS NULL) AS INTEGER) as event_count,
    CAST((SELECT COUNT(DISTINCT name) FROM modules WHERE deleted_at IS NULL) AS INTEGER) as module_count,
    CAST((SELECT COUNT(*) FROM programs WHERE deleted_at IS NULL) AS INTEGER) as program_count,
    CAST((SELECT COUNT(*) FROM activities WHERE deleted_at IS NULL) AS INTEGER) as activity_count,
    CAST((SELECT COUNT(*) FROM sessions) AS INTEGER) as session_count;

-- name: GetDailyUserGrowth :many
SELECT 
    date(created_at) as date,
    COUNT(*) as count
FROM users
WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
GROUP BY date
ORDER BY date ASC;

-- name: GetDailyActivityTrend :many
SELECT 
    date(created_at) as date,
    COUNT(*) as count
FROM activities
WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
GROUP BY date
ORDER BY date ASC;

-- name: GetDailySessionTrend :many
SELECT 
    date(created_at) as date,
    COUNT(*) as count
FROM sessions
WHERE created_at >= date('now', '-90 days')
GROUP BY date
ORDER BY date ASC;

-- name: GetDailyExamGrowth :many
SELECT 
    date(created_at) as date,
    COUNT(*) as count
FROM archive
WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
GROUP BY date
ORDER BY date ASC;

-- name: GetDailyDiscussionGrowth :many
SELECT 
    date,
    CAST(COALESCE(SUM(count), 0) AS INTEGER) as count
FROM (
    SELECT date(created_at) as date, COUNT(*) as count 
    FROM discussion_posts 
    WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
    GROUP BY date
    UNION ALL
    SELECT date(created_at) as date, COUNT(*) as count 
    FROM discussion_comments 
    WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
    GROUP BY date
)
GROUP BY date
ORDER BY date ASC;

-- name: GetDailyModuleGrowth :many
SELECT 
    first_seen as date,
    COUNT(*) as count
FROM (
    SELECT date(MIN(created_at)) as first_seen
    FROM modules
    WHERE deleted_at IS NULL
    GROUP BY name
)
WHERE first_seen >= date('now', '-90 days')
GROUP BY first_seen
ORDER BY first_seen ASC;

-- name: GetDailyProgramGrowth :many
SELECT 
    date(created_at) as date,
    COUNT(*) as count
FROM programs
WHERE created_at >= date('now', '-90 days') AND deleted_at IS NULL
GROUP BY date
ORDER BY date ASC;

-- name: GetUserProgramDistribution :many
SELECT 
    p.name as program_name,
    COUNT(u.id) as user_count
FROM programs p
JOIN users u ON u.program_id = p.id
WHERE u.deleted_at IS NULL AND p.deleted_at IS NULL
GROUP BY p.name;
