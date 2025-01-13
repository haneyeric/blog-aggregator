-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT inserted_feed_follow.*, feeds.name as feed_name, users.name as user_name
FROM inserted_feed_follow
INNER JOIN feeds on inserted_feed_follow.feed_id = feeds.id
INNER JOIN users on inserted_feed_follow.user_id = users.id;


-- name: GetFeedFollowsForUser :many
SELECT feeds.name as feed_name, users.name as user_name
FROM feed_follows
JOIN users ON feed_follows.user_id = users.id
JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
USING feeds
WHERE feed_follows.feed_id = feeds.id
AND feeds.url = $1 AND feed_follows.user_id = $2;
