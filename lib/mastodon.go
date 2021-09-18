package blueskidgo

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

func assertionFromMastodon(url string, fieldCount int) (assertionFields []string, pid string, err error) {

	pid, err = processMastodonURL(url)
	if err != nil {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	assertionFields, err = processMastodonPost(string(bodyBytes), fieldCount)
	return
}

func processMastodonURL(url string) (string, error) {
	// mastodon URLs look like https://mastodon.cloud/@timbray/106939372963435956. First, let's pick out the username
	if !strings.HasPrefix(url, "https://mastodon.") {
		return "", errors.New("not a mastodon url")
	}
	pid := url[len("https://mastodno."):]
	slashAt := strings.Index(pid, "/@")
	if slashAt == -1 {
		return "", errors.New("can't parse URL")
	}
	pid = pid[slashAt + 2:]
	slashAt = strings.Index(pid, "/")
	if slashAt == -1 {
		return "", errors.New("can't parse URL")
	}

	return "mastodon.cloud@" + pid[:slashAt], nil
}

func processMastodonPost(body string, fieldCount int) (assertionFields []string, err error) {
	// best place to pull content out seems to be in a <div class="copy>
	divAt := strings.Index(body, `<div class='e-content'>`)
	if divAt == -1 {
		err = errors.New("can't find 'e-content' div")
		return
	}
	body = body[divAt:]
	divEnd := strings.Index(body, "</div>")
	if divEnd == -1 {
		err = errors.New("can't find e-content div end")
		return
	}
	body = body[:divEnd]
	assertionFields, err = findBlueskidAssertion(body, fieldCount)
	return
}
