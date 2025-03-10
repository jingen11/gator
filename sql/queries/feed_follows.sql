-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT inserted_feed_follow.*,
feeds.name AS feed_name,
users.name AS user_name
FROM inserted_feed_follow
INNER JOIN public.feeds AS feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN public.users AS users ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name, users.name AS user_name
FROM feed_follows
INNER JOIN public.feeds AS feeds ON feeds.id = feed_id
INNER JOIN public.users AS users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;

-- name: UnfollowFeed :exec
WITH current_user_feed_follow AS (
SELECT feed_follows.*, feeds.url AS feed_url
FROM feed_follows
INNER JOIN public.feeds AS feeds ON feeds.url = $1
)
DELETE FROM feed_follows WHERE id IN (SELECT id FROM current_user_feed_follow);
