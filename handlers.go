package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/haneyeric/blog-aggregator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("missing arguments")
	}
	username := cmd.args[0]
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		fmt.Println("User not found")
		os.Exit(1)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", s.cfg.CurrentUserName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("missing arguments")
	}
	username := cmd.args[0]
	createParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}
	user, err := s.db.CreateUser(context.Background(), createParams)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				fmt.Println("Duplicate user")
				os.Exit(1)
			}
		}
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User created: %s\n", user.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		fmt.Printf("Error deleting users: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Users deleted, goodbye")
	os.Exit(0)
	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Error fetching users: %s\n", err)
		os.Exit(1)
	}

	for _, user := range users {
		line := user.Name
		if line == s.cfg.CurrentUserName {
			line += " (current)"
		}
		fmt.Println(line)
	}
	return nil

}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("missing arguments")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		UserID:    user.ID,
		Url:       url,
	}
	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		fmt.Printf("Error creating feed: %s\n", err)
		return err
	}

	fmt.Printf("Following feed: %s\n", feed.Name)

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		fmt.Printf("Error creating feed follow: %s\n", err)
		return err
	}

	fmt.Printf("Feed added: %s - URL: %s", feed.Name, feed.Url)

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Error getting feeds: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(feeds)
	os.Exit(0)
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("missing arguments")
	}

	url := cmd.args[0]
	feedId, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		fmt.Printf("Error getting feed: %s\n", err)
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    feedId,
		UserID:    user.ID,
	}

	feeds, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		fmt.Printf("Error creating feed follow: %s\n", err)
		return err
	}

	fmt.Printf("Following %q as %q\n", feeds[0].FeedName, feeds[0].UserName)

	return nil
}

func handlerFollowing(s *state, _ command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("Error fetching follows: %s\n", err)
		return err
	}
	for _, follow := range follows {
		fmt.Println(follow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("missing arguments")
	}

	url := cmd.args[0]

	params := database.DeleteFeedFollowParams{
		Url:    url,
		UserID: user.ID,
	}

	err := s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		fmt.Printf("Error deleting follow: %s\n", err)
		return err
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("Couldn't get next feeds to fetch", err)
		return
	}
	log.Println("Found a feed to fetch!")
	scrapeFeed(s.db, feed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}
	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			FeedID:    feed.ID,
			Title:     item.Title,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", user.Name)
	fmt.Printf("* LastFetchedAt: %v\n", feed.LastFetchedAt.Time)
}
