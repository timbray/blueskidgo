package blueskidgo

import (
	"errors"
	"html"
	"io"
	"net/http"
	"strings"
)

func assertionFromTumblr (url string, fieldCount int) (assertionFields []string, pid string, err error) {

	pid, err = processTumblrURL(url)
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

	assertionFields, err = processTumblrPost(string(bodyBytes), fieldCount)
	return
}

func processTumblrURL(url string) (string, error) {
	// tumblr URLs are of the form https://$PID.tumblr.com/whatever. First, let's pick out the username
	if !strings.HasPrefix(url, "https://") {
		return "", errors.New("not an HTTPS url")
	}
	pid := url[len("https://"):]
	dotAt := strings.Index(pid, ".")
	if !strings.HasPrefix(pid[dotAt:], ".tumblr.com/") {
		return "", errors.New("not a Tumblr URL")
	}
	return "tumblr.com@" + pid[:dotAt], nil
}

func processTumblrPost(body string, fieldCount int) (assertionFields []string, err error) {
	// best place to pull content out seems to be in a <div class="copy>
	divAt := strings.Index(body, ` <div class="copy">`)
	if divAt == -1 {
		err = errors.New("can't find 'copy' div")
		return
	}
	body = body[divAt:]
	divEnd := strings.Index(body, "</div>")
	if divEnd == -1 {
		err = errors.New("can't find copy div end")
		return
	}
	body = body[:divEnd]
	body = html.UnescapeString(body)
	assertionFields, err = findBlueskidAssertion(body, fieldCount)
	return
}
