package blueskidgo

import (
	"os"
	"testing"
)

func TestProcessMastodonURL(t *testing.T) {

	// busted URL
	_, err := processMastodonURL("http://mastodon/")
	if err == nil {
		t.Error("accepted bad URL")
	}
	_, err = processMastodonURL("https://mastodon.cloud/x")
	if err == nil {
		t.Error("accepted URL with no username")
	}
	pid, err := processMastodonURL("https://mastodon.cloud/@timbray/106939372963435956")
	if err != nil {
		t.Error("problem in procTURL: " + err.Error())
	}
	if pid != "mastodon.cloud@timbray" {
		t.Error("Wrong PID")
	}
}

func TestProcMastodonPost(t *testing.T) {
	body, err := readTestMastodon(t)
	if err == nil {
		parts, e2 := processMastodonPost(body, 6)
		if e2 != nil {
			t.Error("Processing error: " + e2.Error())
		}
		// 🥁A
		// 🎸55555
		// 🎸2021-09-15T05:21:34Z.A
		// 🎸MCowBQYDK2VwAyEA6HTjajtqXj8uLnQnG2bg+01RxnvgCdqTsypD+B2h6bA=
		// 🎸/g+5fI40iowqGibPkuUfd5oJJLYy9ERetRM1Sw/ZdmEylI8JKj55O0WcjaLBfENfQNUTovjRLFfxZYkZfIGvBg==
		// 🎸twitter.com@timbray🥁"
		if parts[0] != "A" ||
			parts[1] != "55555" ||
			parts[2] != "2021-09-15T05:21:34Z.A" ||
			parts[3] != "MCowBQYDK2VwAyEA6HTjajtqXj8uLnQnG2bg+01RxnvgCdqTsypD+B2h6bA=" ||
			parts[4] != "/g+5fI40iowqGibPkuUfd5oJJLYy9ERetRM1Sw/ZdmEylI8JKj55O0WcjaLBfENfQNUTovjRLFfxZYkZfIGvBg==" ||
			parts[5] != "twitter.com@timbray" {
			t.Error("Bad mastodon assertion")
		}
	}
}

func readTestMastodon(t *testing.T) (string, error) {
	f, err := os.Open("../testdata/mastodon-post.txt")
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
	return string(b), nil
}
