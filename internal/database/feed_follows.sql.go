// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: feed_follows.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, created_at, updated_at, user_id, feed_id)
SELECT inserted_feed_follow.id, inserted_feed_follow.created_at, inserted_feed_follow.updated_at, inserted_feed_follow.user_id, inserted_feed_follow.feed_id,
feeds.name AS feed_name,
users.name AS user_name
FROM inserted_feed_follow
INNER JOIN public.feeds AS feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN public.users AS users ON users.id = inserted_feed_follow.user_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
}

type CreateFeedFollowRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
	FeedName  string
	UserName  string
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (CreateFeedFollowRow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	var i CreateFeedFollowRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.FeedID,
		&i.FeedName,
		&i.UserName,
	)
	return i, err
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name, users.name AS user_name
FROM feed_follows
INNER JOIN public.feeds AS feeds ON feeds.id = feed_id
INNER JOIN public.users AS users ON users.id = feed_follows.user_id
W




feed_follows.user_id = $1
`

type GetFeedFollowsForUserRow struct {
	FeedName string
	UserName string
}

func (q *Queries) GetFeedFollowsForUser(ctx context.Context, userID uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(&i.FeedName, &i.UserName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const unfollowFeed = `-- name: UnfollowFeed :exec
WITH current_user_feed_follow AS (
SELECT feed_follows.id, feed_follows.created_at, feed_follows.updated_at, feed_follows.user_id, feed_follows.feed_id, feeds.url AS feed_url
FROM feed_follows
INNER JOIN public.feeds AS feeds ON feeds.url = $1
)
DELETE FROM feed_follows WHERE id IN (SELECT id FROM current_user_feed_follow)
`

func (q *Queries) UnfollowFeed(ctx context.Context, url string) error {
	_, err := q.db.ExecContext(ctx, unfollowFeed, url)
	return err
}
