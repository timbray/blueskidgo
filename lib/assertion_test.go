package blueskidgo

import (
	"fmt"
	"strings"
	"testing"
)

func TestFindBlueskidAssertion(t *testing.T) {
	text := "🥁G🎸30900000021🎸2021-09-09T05:47:35Z.G🎸MCowBQYDK2VwAyEAG4Hs/FA/ylsiR2+Gmg58ZTS68gz0/ZuH3dgn/kF/YJ0=🎸SUNjhrp4yTublRn/7ytrDeicaJ62WbnmBbmOBKIWhoJVik/ICIgX9UWU3aYZpIDIo9HbSS73nKF5rfN8gQN8CQ==🎸reddit.com@tim🥁"
	a, err := findBlueskidAssertion(text, 6)
	if err != nil {
		t.Error("fba failed: " + err.Error())
	}
	if a[0] != "G" ||
		a[1] != "30900000021" ||
		a[2] != "2021-09-09T05:47:35Z.G" ||
		a[3] != "MCowBQYDK2VwAyEAG4Hs/FA/ylsiR2+Gmg58ZTS68gz0/ZuH3dgn/kF/YJ0=" ||
		a[4] != "SUNjhrp4yTublRn/7ytrDeicaJ62WbnmBbmOBKIWhoJVik/ICIgX9UWU3aYZpIDIo9HbSS73nKF5rfN8gQN8CQ==" ||
		a[5] != "reddit.com@tim" {
		t.Error("Incorrect parse")
	}
	text = "Some random stuff" + text + " still more"
	a, err = findBlueskidAssertion(text, 6)
	if err != nil {
		t.Error("fba failed: " + err.Error())
	}
	if a[0] != "G" ||
		a[1] != "30900000021" ||
		a[2] != "2021-09-09T05:47:35Z.G" ||
		a[3] != "MCowBQYDK2VwAyEAG4Hs/FA/ylsiR2+Gmg58ZTS68gz0/ZuH3dgn/kF/YJ0=" ||
		a[4] != "SUNjhrp4yTublRn/7ytrDeicaJ62WbnmBbmOBKIWhoJVik/ICIgX9UWU3aYZpIDIo9HbSS73nKF5rfN8gQN8CQ==" ||
		a[5] != "reddit.com@tim" {
		t.Error("Incorrect parse")
	}
}

func TestGoodAssertions(t *testing.T) {

	var bid uint64
	bid = uint64((33 << 32) | 33)
	for i := 0; i < 100; i++ {
		grant, accept, err := generateGrantAssertions(bid, "twitter.com@tim", "reddit.com@tim")
		if err != nil {
			t.Error("Generate failed: " + err.Error())
		}
		gFields, err := findBlueskidAssertion(grant, 6)
		if err != nil {
			t.Error(err.Error())
		}
		aFields, err := findBlueskidAssertion(accept, 6)
		if err != nil {
			t.Error(err.Error())
		}

		reportedBID, err := checkGrantAssertionPair(gFields, "twitter.com@tim", aFields, "reddit.com@tim")
		if err != nil {
			t.Error("Check failed: " + err.Error())
		}
		if reportedBID != bid {
			t.Error("Wrong BID in grant")
		}
	}
}

func TestBadAssertions(t *testing.T) {
	bid := (777 << 32) | 33
	grant, accept, err := generateGrantAssertions(uint64(bid), "twitter.com@tim", "reddit.com@tim")
	if err != nil {
		t.Error("Generate failed: " + err.Error())
	}
	fmt.Println("G: " + grant + "\nA: " + accept)

	var failer Failer
	failer.t = t

	var gParts, aParts []string

	// insufficient fields
	gParts, aParts = split(grant, accept)
	failer.failParts(gParts[:3], aParts, "accepted insufficient grant fields")
	failer.failParts(gParts, aParts[:4], "accepted insufficient accept fields")

	// bad dopcode
	gParts, aParts = split(grant, accept)
	gParts[0] = "X"
	failer.failParts(gParts, aParts, "accepted bad grant opcode")
	gParts[0] = "G"
	aParts[0] = "zerch"
	failer.failParts(gParts, aParts, "accepted bad accept opcode")

	// borked BID
	gParts, aParts = split(grant, accept)
	gParts[BID] = "foo"
	failer.failParts(gParts, aParts, "accepted bad grant BID")
	gParts, aParts = split(grant, accept)
	aParts[BID] = "003h"
	failer.failParts(gParts, aParts, "accepted bad grant BID")

	// borked nonce suffix
	gParts, aParts = split(grant, accept)
	gParts[ClaimNonce] = strings.ReplaceAll(gParts[ClaimNonce], ".G", "Xr")
	failer.failParts(gParts, aParts, "accepted bad grant nonce suffix")
	gParts[ClaimNonce] = strings.ReplaceAll(gParts[ClaimNonce], "Xr", ".A")
	failer.failParts(gParts, aParts, "accepted bad grant nonce suffix")
	gParts, aParts = split(grant, accept)
	aParts[ClaimNonce] = strings.ReplaceAll(gParts[ClaimNonce], ".A", "=5")
	failer.failParts(gParts, aParts, "accepted bad accept nonce suffix")
	aParts[ClaimNonce] = strings.ReplaceAll(gParts[ClaimNonce], "=5", ".G")
	failer.failParts(gParts, aParts, "accepted bad accept nonce suffix")

	// borked nonce timestamp
	gParts, aParts = split(grant, accept)
	gParts[ClaimNonce] = "z" + gParts[ClaimNonce]
	failer.failParts(gParts, aParts, "accepted bogus grant timestamp")
	gParts, aParts = split(grant, accept)
	aParts[ClaimNonce] = "foo.A"
	failer.failParts(gParts, aParts, "accepted bogus accept timestamp")

	// borked key
	gParts, aParts = split(grant, accept)
	gParts[ClaimKey] = "really-not-a-key"
	failer.failParts(gParts, aParts, "accepted bogus grant key")
	gParts, aParts = split(grant, accept)
	aParts[ClaimKey] = "*" + aParts[ClaimNonce]
	failer.failParts(gParts, aParts, "accepted bogus accept key")

	// non-base64 sig
	gParts, aParts = split(grant, accept)
	gParts[ClaimSig] = "!" + gParts[ClaimSig]
	failer.failParts(gParts, aParts, "accepted bogus grant key")
	gParts, aParts = split(grant, accept)
	aParts[ClaimSig] = "&&"
	failer.failParts(gParts, aParts, "accepted bogus accept key")

	// can't verify
	gParts, aParts = split(grant, accept)
	bytes := []byte(gParts[ClaimSig])
	bytes[0] += 1
	gParts[ClaimSig] = string(bytes)
	failer.failParts(gParts, aParts, "accepted broken grant key")
	gParts, aParts = split(grant, accept)
	bytes = []byte(aParts[ClaimSig])
	bytes[0] += 1
	aParts[ClaimSig] = string(bytes)
	failer.failParts(gParts, aParts, "accepted broken accept key")
	gParts, aParts = split(grant, accept)
	gParts[ClaimSig] = aParts[ClaimSig]
	failer.failParts(gParts, aParts, "accepted broken grant key")

	// different pubkeys
	_, a2, err := generateGrantAssertions(uint64(bid), "twitter.com@tim", "reddit.com@tim")
	failer.failStrings(grant, a2, "accepted different pubkeys")

	// same nonce
	// - can't test this, this program insists that the nonce end in .G and .A respectively

	// different BIDs
	gParts, aParts = split(grant, accept)
	aParts[BID] = "33"
	failer.failParts(gParts, aParts, "accepted different BIDs")
}

func split(g string, a string) ([]string, []string) {
	return strings.Split(g, Guitar), strings.Split(a, Guitar)
}

type Failer struct {
	t       *testing.T
	verbose bool
}

func (f *Failer) failParts(gp []string, ap []string, msg string) {
	_, err := checkGrantAssertionPair(gp, "twitter.com@tim", ap, "reddit.com@tim")
	if err == nil {
		f.t.Error("Should fail: " + msg)
	} else {
		if f.verbose {
			fmt.Println("E: " + msg + ": " + err.Error())
		}
	}
}

func (f *Failer) failStrings(g string, a string, msg string) {
	gp, err := findBlueskidAssertion(g, 6)
	if err != nil {
		f.t.Error(err.Error())
	}
	ap, err := findBlueskidAssertion(a, 6)
	if err != nil {
		f.t.Error(err.Error())
	}
	_, err = checkGrantAssertionPair(gp, "twitter.com@tim", ap, "reddit.com@tim")
	if err == nil {
		f.t.Error("Should fail: " + msg)
	}
}
