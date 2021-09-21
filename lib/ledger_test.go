package blueskidgo

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"
)

func newPubKey() string {
	public, _, _ := ed25519.GenerateKey(rand.Reader)
	s, _ := KeyToString(public)
	return s
}

func TestDatabase(t *testing.T) {
	BIDnums := []uint64{0xb1b1b1, 0xb2b2b2, 0xb3b3b3}
	var BIDs []string
	for _, b := range BIDnums {
		BIDs = append(BIDs, fmt.Sprintf("%016x", b))
	}

	PIDs := []string{"twitter.com@p1", "reddit.com@p2", "tumblr.com@p3" }

	// pid claims
	var err error
	for i := 0; i < 3; i++ {
		err = appendToLedger(&LedgerRecord{RecType: ClaimBID, BID: BIDs[i], PIDs: []string{PIDs[i]}, Key: newPubKey()})
		if err != nil {
			t.Errorf("PID claim %d: %s", i, err.Error())
		}
	}

	// claim already-claimed BID
	err = appendToLedger(&LedgerRecord{RecType: ClaimBID, BID: BIDs[1], PIDs: []string{PIDs[0]}, Key: newPubKey()})
	if err == nil {
		t.Error("accepted duplicate BID claim")
	}

	// unclaim it all
	for i := 0; i < 3; i++ {
		err = appendToLedger(&LedgerRecord{RecType: UnclaimBID, BID: BIDs[i], PIDs: []string{PIDs[i]}, Key: newPubKey()})
		if err != nil {
			t.Errorf("PID claim %d: %s", i, err.Error())
		}
	}

	// try to unclaim again
	err = appendToLedger(&LedgerRecord{RecType: UnclaimBID, BID: BIDs[2], PIDs: []string{PIDs[2]}, Key: newPubKey()})
	if err == nil {
		t.Error("accepted double unclaim")
	}

	// try to claim an already-used BID
	err = appendToLedger(&LedgerRecord{RecType: ClaimBID, BID: BIDs[2], PIDs: []string{PIDs[2]}, Key: newPubKey()})
	if err == nil {
		t.Error("re-use BID")
	}

	// fresh new BIDs
	BIDnums = []uint64{0xbb1, 0xbb2, 0xbb3}
	for i, b := range BIDnums {
		BIDs[i] = fmt.Sprintf("%016x", b)
	}

	// p1 claims b1, gives it to b2 and b3
	// p2 claims b2, gives it to p3
	// p3 claims b3
	for i := 0; i < 3; i++ {
		err = appendToLedger(&LedgerRecord{RecType: ClaimBID, BID: BIDs[i], PIDs: []string{PIDs[i]}})
		if err != nil {
			t.Errorf("PID claim %d: %s", i, err.Error())
		}
	}

	grant := LedgerRecord{
		RecType:  GrantBID,
		BID:      BIDs[0],
		PIDs:     []string{PIDs[0], PIDs[1]},
		PostURLs: []string{"g12", "a12"},
		Key:      newPubKey(),
	}
	err = appendToLedger(&grant)
	if err != nil {
		t.Error("prob with grant 1->2")
	}

	grant.PIDs = []string{PIDs[0], PIDs[2]}
	grant.PostURLs = []string{"g13", "a13"}
	grant.Key = newPubKey()
	err = appendToLedger(&grant)
	if err != nil {
		t.Error("prob with grant 1->3: " + err.Error())
	}

	grant.BID = BIDs[1]
	grant.PIDs = []string{PIDs[1], PIDs[2]}
	grant.PostURLs = []string{"g23", "a23"}
	grant.Key = newPubKey()
	err = appendToLedger(&grant)
	if err != nil {
		t.Error("prob with grant 2->3: " + err.Error())
	}

	// now, p3 mapped to all 3, p2 mapped to 2 & 3, p1 mapped to only 1
	var group map[string]bool

	for _, subject := range PIDs {
		group = makePIDgroup(subject)
		if len(group) != 3 {
			t.Errorf("Group size is %d should be 3", len(group))
		}
		for _, pid := range(PIDs) {
			if !group[pid] {
				t.Error("PID " + pid + "not in group")
			}
		}
	}

	group = makePIDgroup("friendless")
	if (len(group) != 1) || !group["friendless"] {
		t.Error("not friendless")
	}
}
