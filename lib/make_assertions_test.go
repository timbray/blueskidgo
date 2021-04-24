package blueskidgo

import (
	"encoding/json"
	"testing"
)

func TestNewGrantAssertionsResponse(t *testing.T) {
	_, problem, myFault := newGrantAssertionsResponse([]byte("Definitely not JSON"))
	if problem == "" {
		t.Error("Failed to detect invalid JSON")
	}
	if myFault {
		t.Error("Blamed self for invalid JSON")
	}

	goodRequest := `{
  "BID": "30900000021",
  "Granter": "twitter.com@tim",
  "Accepter": "reddit.com@tim"
}`
	missingField := `{
  "BID": "30900000021",
  "Accepter": "reddit.com@tim"
}`
	badBID := `{
  "BID": "notHex",
  "Granter": "twitter.com@tim",
  "Accepter": "reddit.com@tim"
}`
	bid := uint64(0x30900000021)

	_, problem, myFault = newGrantAssertionsResponse([]byte(missingField))
	if problem == "" {
		t.Error("Failed to detect missing field")
	}
	if myFault {
		t.Error("Blamed self for missing field")
	}

	_, problem, myFault = newGrantAssertionsResponse([]byte(badBID))
	if problem == "" {
		t.Error("Failed to detect bad BID")
	}
	if myFault {
		t.Error("Blamed self for bad BID")
	}

	bytes, problem, myFault := newGrantAssertionsResponse([]byte(goodRequest))
	if problem != "" {
		t.Error("Failed on good request: " + problem)
	}

	var resp grantAssertionsResponse
	err := json.Unmarshal(bytes, &resp)
	if err != nil {
		t.Error("got invalid JSON: " + problem)
	}

	gParts, err := findBlueskidAssertion(resp.GrantAssertion, 6)
	if err != nil {
		t.Error(err.Error())
	}
	aParts, err := findBlueskidAssertion(resp.AcceptAssertion, 6)
	if err != nil {
		t.Error(err.Error())
	}

	foundBid, err := checkGrantAssertionPair(gParts, "twitter.com@tim", aParts, "reddit.com@tim")
	if err != nil {
		t.Error("Assertion check failed: " + err.Error())
	}
	if bid != foundBid {
		t.Error("bid and foundBid differ")
	}

	var req grantAssertionsRequest
	err = json.Unmarshal([]byte(goodRequest), &req)
	if err != nil {
		t.Error("Can't parse goodRequest: " + err.Error())
	}
}
