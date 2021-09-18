package blueskidgo

import (
	"os"
	"testing"
)

func TestProcessTumblrURL(t *testing.T) {

	// busted URL
	_, err := processTumblrURL("http://tumblr.com/")
	if err == nil {
		t.Error("accepted bad URL")
	}
	_, err = processTumblrURL("https://tumblr.com")
	if err == nil {
		t.Error("accepted URL with no username")
	}
	pid, err := processTumblrURL("https://t-runic.tumblr.com/post/662425486899691520/blueskid-assertion")
	if err != nil {
		t.Error("problem in procTURL: " + err.Error())
	}
	if pid != "tumblr.com@t-runic" {
		t.Error("Wrong PID")
	}
}

func TestProcTumblrPost(t *testing.T) {
	body, err := readTestTumblr(t)
	if err == nil {
		parts, e2 := processTumblrPost(body, 6)
		if e2 != nil {
			t.Error("Processing error: " + e2.Error())
		}
		// 🥁A
		// 🎸55555
		// 🎸2021-09-15T05:25:47Z.A
		// 🎸MCowBQYDK2VwAyEAX51DzwGncOsU87Y4xVoiFlNLLH8FTgSSIPG3ZutQbGc=
		// 🎸W+kWrsb6WS1y2DPwYbQUtRSDm/b78WE98H6wifrgwSuYjgiWl7kOkVn4xbXcAbdYzGVd51zz+zao2FFk68E+AQ==
		// 🎸twitter.com@timbray🥁
		if parts[0] != "A" ||
			parts[1] != "55555" ||
			parts[2] != "2021-09-15T05:25:47Z.A" ||
			parts[3] != "MCowBQYDK2VwAyEAX51DzwGncOsU87Y4xVoiFlNLLH8FTgSSIPG3ZutQbGc=" ||
			parts[4] != "W+kWrsb6WS1y2DPwYbQUtRSDm/b78WE98H6wifrgwSuYjgiWl7kOkVn4xbXcAbdYzGVd51zz+zao2FFk68E+AQ==" ||
			parts[5] != "twitter.com@timbray" {
			t.Error("Bad tumblr assertion")
		}
	}
}

func readTestTumblr(t *testing.T) (string, error) {
	f, err := os.Open("../testdata/tumblr-post.txt")
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
