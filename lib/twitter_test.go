package blueskidgo

import (
	"encoding/json"
	"os"
	"testing"
)

// const URL = "https://twitter.com/ArtisanPortents/status/1435031524039036933"
const URL = "https://twitter.com/ArtisanPortents/status/1436831923330977798"

func TestFetchAssertion(t *testing.T) {
	// OK, I'm going out to call the real live API in a unit test. So sue me.
	tweet, err := fetchTweet(URL)
	if err != nil {
		t.Error("GET: " + err.Error())
	}

	wanted, err := readTestTweet(t)
	if err != nil {
		t.Error("Can't read: " + err.Error())
	}

	if tweet.Data.AuthorId != wanted.Data.AuthorId {
		t.Error("Wrong AuthorID")
	}
	if tweet.Data.ID != wanted.Data.ID {
		t.Error("Wrong data.ID")
	}
	if tweet.Data.Text != wanted.Data.Text {
		t.Error("Wrong text")
	}
	if tweet.Includes.Users[0].ID != wanted.Includes.Users[0].ID {
		t.Error("wrong Includes ID")
	}
	if tweet.Includes.Users[0].Name != wanted.Includes.Users[0].Name {
		t.Error("wrong Includes Name")
	}
	if tweet.Includes.Users[0].Username != wanted.Includes.Users[0].Username {
		t.Error("wrong Includes Unsername")
	}
}

func TestAssertionFromTweet(t *testing.T) {

	tweet, err := readTestTweet(t)
	if err != nil {
		t.Error("Can't read: " + err.Error())
	}

	_, pid, err := assertionFromTweet(tweet, URL, 6)
	if err != nil {
		t.Error("aFT failed: " + err.Error())
	}

	if pid != "twitter.com@ArtisanPortents" {
		t.Error("PID mismatch")
	}

	_, _, err = assertionFromTweet(tweet, "https://splitter.com/", 6)
	if err == nil {
		t.Error("Missed non-tweet URL")
	}
	_, _, err = assertionFromTweet(tweet, "https://twitter.com/blob", 6)
	if err == nil {
		t.Error("Missed non-tweet URL")
	}
	_, _, err = assertionFromTweet(tweet, "https://twitter.com", 6)
	if err == nil {
		t.Error("Missed non-tweet URL")
	}

	rememberUsername := tweet.Includes.Users[0].Username
	tweet.Includes.Users[0].Username = "notAP"
	_, _, err = assertionFromTweet(tweet, URL, 6)
	if err == nil {
		t.Error("Missed username mismatch")
	}
	tweet.Includes.Users[0].Username = rememberUsername
	badURL := "https://twitter.com/notAP/status/1435031524039036933"
	_, _, err = assertionFromTweet(tweet, badURL, 6)
	if err == nil {
		t.Error("Missed username mismatch")
	}

	tweet.Data.Text = "No assertion here"
	_, _, err = assertionFromTweet(tweet, URL, 6)
	if err == nil {
		t.Error("should have detected missing assertin")
	}
}

func readTestTweet(t *testing.T) (*tweet, error) {
	f, err := os.Open("../testdata/tweet.json")
	if err != nil {
		t.Error("Can't open file " + err.Error())
	}
	info, err := f.Stat()
	if err != nil {
		t.Error("Can't stat file" + err.Error())
	}
	var b []byte
	b = make([]byte, info.Size(), info.Size())
	_, err = f.Read(b)
	if err != nil {
		t.Error("Can't read file: " + err.Error())
	}
	var wanted tweet
	err = json.Unmarshal(b, &wanted)
	return &wanted, nil
}
