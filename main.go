package main

import (
	"flag"
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func main() {
	consumerKey := flag.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flag.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flag.String("access-token", "", "Twitter Access Token")
	accessSecret := flag.String("access-secret", "", "Twitter Access Secret")

	flag.Parse()

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Println("Consumer key/secret and Access token/secret required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)

	client, err := newClient(config, token)

	if err != nil {
		log.Fatalln(err)
	}

	client.batchDelete(0)
}

type client struct {
	tc *twitter.Client
}

func newClient(config *oauth1.Config, token *oauth1.Token) (*client, error) {

	httpClient := config.Client(oauth1.NoContext, token)

	tc := twitter.NewClient(httpClient)

	_, _, err := tc.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	})

	client := &client{tc: tc}

	return client, err
}

func (c *client) batchDelete(id int64) {
	maxID := id
	tweets, err := c.timeline(id)

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Found %v tweets", len(tweets))

	for _, tweet := range tweets {
		c.delete(tweet)
		maxID = tweet.ID
	}

	if maxID != id {
		c.batchDelete(maxID)
	}
}

func (c *client) timeline(maxID int64) ([]twitter.Tweet, error) {
	userTimelineParams := &twitter.UserTimelineParams{
		ExcludeReplies:  twitter.Bool(false),
		IncludeRetweets: twitter.Bool(true),
		Count:           200,
	}

	if maxID > 0 {
		userTimelineParams.MaxID = maxID
	}

	tweets, _, err := c.tc.Timelines.UserTimeline(userTimelineParams)

	return tweets, err
}

func (c *client) delete(tweet twitter.Tweet) {
	log.Printf("Deleting tweet %v:", tweet.ID)

	_, _, err := c.tc.Statuses.Destroy(tweet.ID, nil)

	if err != nil {
		log.Fatalf("Error deleting tweet %v: %v", tweet.ID, err)
	}
}
