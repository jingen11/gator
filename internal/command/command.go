package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jingen11/gator/internal/database"
	"github.com/jingen11/gator/internal/rss"
	"github.com/jingen11/gator/internal/state"
)

type Command struct {
	Name      string
	Arguments []string
}

func scrapeFeeds(s *state.State) error {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.Db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: feed.ID,
	})
	if err != nil {
		return err
	}
	fetched, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, v := range fetched.Channel.Item {
		publishAt, err := time.Parse(time.RFC1123, v.PubDate)
		if err != nil {
			fmt.Println(err)
		}
		_, err = s.Db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     v.Title,
			Url:       v.Link,
			Description: sql.NullString{
				String: v.Description,
				Valid:  true,
			},
			PublishedAt: publishAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func MiddlewareLoggedIn(handler func(*state.State, *Command, database.User) error) func(*state.State, *Command) error {
	return func(s *state.State, c *Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Conf.CurrentUserName)
		if err != nil {
			return err
		}
		err = handler(s, c, user)
		if err != nil {
			return err
		}
		return nil
	}
}

func HandlerLogin(s *state.State, cmd *Command) error {
	if len(cmd.Arguments) != 1 {
		return errors.New("Invalid argument")
	}

	username := cmd.Arguments[0]

	_, err := s.Db.GetUser(context.Background(), username)

	if err != nil {
		return err
	}
	err = s.Conf.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Println("user has been set")

	return nil
}

func HandlerRegister(s *state.State, cmd *Command) error {
	if len(cmd.Arguments) != 1 {
		return errors.New("Invalid argument")
	}
	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Arguments[0],
	})
	if err != nil {
		return err
	}
	err = s.Conf.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("user created %v", user)
	return nil
}

func HandlerReset(s *state.State, cmd *Command) error {
	err := s.Db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("db reset")
	return nil
}

func HandlerListUsers(s *state.State, cmd *Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	currentUser := s.Conf.CurrentUserName

	for _, u := range users {
		if u.Name == currentUser {
			fmt.Printf("* %s (current)\n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}
	return nil
}

func HandlerAggregate(s *state.State, cmd *Command) error {
	if len(cmd.Arguments) != 1 {
		return errors.New("Invalid argument")
	}
	duration, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		fmt.Printf("Collecting feeds every %dm%ds\n", int(duration.Minutes()), int(duration.Seconds())%60)
		scrapeFeeds(s)
	}

	return nil
}

func HandlerAddFeed(s *state.State, cmd *Command, user database.User) error {
	if len(cmd.Arguments) != 2 {
		return errors.New("Invalid argument")
	}
	_, err := rss.FetchFeed(context.Background(), cmd.Arguments[1])
	if err != nil {
		return err
	}
	feed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Arguments[0],
		Url:       cmd.Arguments[1],
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}
	s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	fmt.Printf("%v", feed)
	return nil
}

func HandlerListFeeds(s *state.State, cmd *Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, f := range feeds {
		fmt.Printf("* %v", f)
	}
	return nil
}

func HandlerFollow(s *state.State, cmd *Command, user database.User) error {
	if len(cmd.Arguments) != 1 {
		return errors.New("Invalid argument")
	}
	url := cmd.Arguments[0]

	feed, err := s.Db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}
	s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	return nil
}

func HandlerFollowing(s *state.State, cmd *Command, user database.User) error {
	feeds, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("feeds: %v", feeds)
	return nil
}

func HandlerUnfollow(s *state.State, cmd *Command, user database.User) error {
	if len(cmd.Arguments) != 1 {
		return errors.New("Invalid argument")
	}
	url := cmd.Arguments[0]

	s.Db.UnfollowFeed(context.Background(), url)
	return nil
}

func HandlerBrowse(s *state.State, cmd *Command, user database.User) error {
	if len(cmd.Arguments) != 1 {
		cmd.Arguments = append(cmd.Arguments, "2")
	}
	limit, err := strconv.Atoi(cmd.Arguments[0])
	if err != nil {
		return err
	}
	posts, err := s.Db.GetPosts(context.Background(), int32(limit))

	for _, v := range posts {
		fmt.Println(v)
	}
	return nil
}
