package blueskidgo

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type tweetData struct {
	AuthorId string `json:"author_id"`
	ID       string `json:"id"`
	Text     string `json:"text"`
}
type tweetUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}
type tweetIncludes struct {
	Users []tweetUser `json:"users"`
}
type tweet struct {
	Data     tweetData     `json:"data"`
	Includes tweetIncludes `json:"includes"`
}

func assertionFromTwitter(url string, fieldCount int) (assertionFields []string, pid string, err error) {
	tweet, err := fetchTweet(url)
	if err != nil {
		return
	}
	assertionFields, pid, err = assertionFromTweet(tweet, url, fieldCount) // making it easier to test
	return
}

// given a tweet instance, extract the blueskid assertion and the author and return both
func assertionFromTweet(tweet *tweet, url string, fieldCount int) (assertionFields []string, pid string, err error) {

	// verify it's really a twitter URL
	twitterPrefix := "https://twitter.com/"
	if !strings.HasPrefix(url, twitterPrefix) {
		err = errors.New("not a twitter URL")
		return
	}

	// pull out the twitter handle
	usernameFromUrl := url[len(twitterPrefix):]
	slashAt := strings.Index(usernameFromUrl, "/")

	// if there isn't one, this isn't really a Tweet
	if slashAt == -1 {
		err = errors.New("not a tweet URL")
		return
	}
	usernameFromUrl = usernameFromUrl[:slashAt]

	// sanity check that the official poster matches the username in the twitter
	if usernameFromUrl != tweet.Includes.Users[0].Username {
		err = errors.New("username in tweet doesn't match username in URL")
		return
	}

	assertionFields, err = findBlueskidAssertion(tweet.Data.Text, fieldCount)
	if err != nil {
		return
	}

	pid = "twitter.com@" + usernameFromUrl
	return
}

// retrieve a tweet and parse it into JSON
// Note: Must have an approved developer account and make bearer token available
func fetchTweet(url string) (*tweet, error) {

	// the API wants the post ID number not the original URL
	uParts := strings.Split(url, "/")
	trailer := uParts[len(uParts)-1]
	match, _ := regexp.MatchString("^[0-9]*$", trailer)
	if !match {
		return nil, errors.New("malformed URL, should end with digit sequence")
	}

	query := "https://api.twitter.com/2/tweets/" + trailer + "?expansions=author_id&tweet.fields=author_id"

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, errors.New("Failed to create request: " + err.Error())
	}
	token := os.Getenv("TWITTER_BEARER_TOKEN")
	if token == "" {
		return nil, errors.New("TWITTER_BEARER_TOKEN not set in environment")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("GET failed: " + err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err == nil {
		_ = resp.Body.Close()
	} else {
		return nil, err
	}

	var t tweet
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, errors.New("error parsing twitter JSON: " + err.Error())
	}
	return &t, nil
}
