package blueskidgo

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// Confession: The ledger is a fake, lives only in memory and is not transactional, not concurrent, and not
//  persisted. Databases are hard and this is just a demo!

type recordType int

const (
	ClaimBID = iota
	GrantBID
	UnclaimBID
)

// LedgerRecord
//  we could have separate types for ClaimBID, GrantBID, and UnclaimBID, but all of these things have a BID,
//  one or more PIDs, and the URLs of one or more posts.  So, this has less casting.
// for ClaimBID: PIDs[0] is the claimer, PostURLs[0] is the claim post.
// for GrantBID: PIDS[0] and [1] are the claimer and accepter, and PostURLs[0] & [1] the grant/accept posts
// for UnclaimbID: PIDS[0] is the unclaimer, PostURLs[0] is the unclaim post
// The Key field is provided only for Grant records, to help ensure no re-use of key-pairs.
type LedgerRecord struct {
	RecType  recordType
	BID      string
	PIDs     []string
	PostURLs []string
	Key      string
}

// we'll build a dumb little database to maintain BID/PID mappings

// PIDsForBID is indexed by BID; values are set-like maps containing the PIDs mapped to that BID
var PIDsForBID = make(map[string]map[string]bool)

// BIDsForPID is indexed by PID; values are set-like maps containing the BIDs mapped to that BID
var BIDsForPID = make(map[string]map[string]bool)

// KeysUsed tracks the public keys that have appeared in assertions, so as to prevent re-use.
var KeysUsed = make(map[string]bool)

/*
// for debugging
func dumpDB(label string) {
	fmt.Println(label)
	fmt.Println("P4B")
	for k, v := range PIDsForBID {
		fmt.Println(k + ": ")
		for kk := range v {
			fmt.Println("  " + kk)
		}
	}
	fmt.Println("\nB4P")
	for k, v := range BIDsForPID {
		fmt.Println(k + ": ")
		for kk := range v {
			fmt.Println("  " + kk)
		}
	}
}
*/

type LedgerScanner interface {
	processRecord(record *LedgerRecord) error
}

func Scan(scanner LedgerScanner) error {
	var err error
	for _, record := range theLedger.Records {
		err = scanner.processRecord(record)
		if err != nil {
			return err
		}
	}
	return nil
}

type ledger struct {
	Records []*LedgerRecord
}

var theLedger  = ledger{Records: make([]*LedgerRecord,0)}
var theLock sync.Mutex

// appendToLedger also performs sanity-checking to make sure the claim/unclaim/grant being requested is legitimate
func appendToLedger(record *LedgerRecord) error {

	// mad skillz at distributed concurrency
	theLock.Lock()
	defer theLock.Unlock()

	switch record.RecType {
	case ClaimBID:
		claimingPID := record.PIDs[0]

		// is this BID available?
		_, ok := PIDsForBID[record.BID]
		if ok {
			return errors.New("BID '" + record.BID + "' has already been claimed by another account")
		}

		// map from BID to PID
		PIDsForBID[record.BID] = map[string]bool{claimingPID: true}

		// map from PID to BID
		bidsForClaimingPID, ok := BIDsForPID[claimingPID]
		if !ok {
			bidsForClaimingPID = make(map[string]bool)
			BIDsForPID[claimingPID] = bidsForClaimingPID
		}
		bidsForClaimingPID[record.BID] = true

	case GrantBID:
		granter := record.PIDs[0]
		accepter := record.PIDs[1]
		pidsForGrantedBID, ok := PIDsForBID[record.BID]

		// granter has to own PID
		if !ok {
			return errors.New("no such BID: " + record.BID)
		}
		_, ok = pidsForGrantedBID[granter]
		if !ok {
			return errors.New("this account is not mapped to BID " + record.BID)
		}

		// has key been used?
		_, ok = KeysUsed[record.Key]
		if ok {
			return errors.New("public key has been used in a previous grant transaction")
		}
		KeysUsed[record.Key] = true

		// map from BID to accepter PID
		pidsForGrantedBID[accepter] = true

		// map from accepter PID to BID
		bidsForAccepter, ok := BIDsForPID[accepter]
		if !ok {
			bidsForAccepter = make(map[string]bool)
			BIDsForPID[accepter] = bidsForAccepter
		}
		bidsForAccepter[record.BID] = true

	case UnclaimBID:
		// can only do this if this BID exists and I'm mapped to it
		currentPIDs, ok := PIDsForBID[record.BID]
		if !ok {
			return errors.New("no such BID: " + record.BID)
		}
		_, ok = currentPIDs[record.PIDs[0]]
		if !ok {
			return errors.New("this account is not mapped to BID " + record.BID)
		}

		// remove the mapping between PID to BID
		// note - the pidsForGrantedBID map may now be empty but we won't free up the BID, because they probably
		//  shouldn't be re-used.
		delete(currentPIDs, record.PIDs[0])

		currentBIDs, _ := BIDsForPID[record.PIDs[0]]
		delete(currentBIDs, record.BID)
	}

	theLedger.Records = append(theLedger.Records, record)
	return nil
}

func LedgerHandler(w http.ResponseWriter, _ *http.Request) {
	bytes, err := json.MarshalIndent(theLedger, "", " ")
	writeJson(w, bytes, err)
}

type getPIDGroupHandlerResult struct {
	PIDGroup []string
}

func makePIDgroup(pid string) map[string]bool {
	var group = map[string]bool{pid: true}
	// a PID=>PID map could be precomputed of course
	for bid := range BIDsForPID[pid] {
		for otherPid := range PIDsForBID[bid] {
			group[otherPid] = true
		}
	}
	return group
}

func GetPIDGroupHandler(w http.ResponseWriter, httpRequest *http.Request) {
	if !openGet(w, httpRequest) {
		return
	}
	pid := httpRequest.Form.Get("pid")

	if pid == "" {
		http.Error(w, "missing parameter 'pid'", http.StatusBadRequest)
		return
	}

	group := makePIDgroup(pid)
	var resp getPIDGroupHandlerResult
	for member := range group {
		resp.PIDGroup = append(resp.PIDGroup, member)
	}

	respJSON, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		http.Error(w, "response creation failure: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, respJSON, err)
	return
}

type getBIDsforPIDResponse struct {
	BIDs []string
}

func GetBIDsforPIDHandler(w http.ResponseWriter, httpRequest *http.Request) {
	if !openGet(w, httpRequest) {
		return
	}
	pid := httpRequest.Form.Get("pid")

	if pid == "" {
		http.Error(w, "missing parameter 'pid'", http.StatusBadRequest)
		return
	}
	var resp getBIDsforPIDResponse
	pidSet, ok := BIDsForPID[pid]
	if ok {
		for bid := range pidSet {
			resp.BIDs = append(resp.BIDs, bid)
		}
	}

	respJSON, err := json.MarshalIndent(resp, "", " ")
	writeJson(w, respJSON, err)
	return
}

type getPIDsForBIDResponse struct {
	PIDs []string
}

func GetPIDsForBIDHandler(w http.ResponseWriter, httpRequest *http.Request) {
	if !openGet(w, httpRequest) {
		return
	}
	bid := httpRequest.Form.Get("bid")

	if bid == "" {
		http.Error(w, "missing parameter 'bid'", http.StatusBadRequest)
		return
	}
	var resp getPIDsForBIDResponse
	bidSet, ok := PIDsForBID[bid]
	if ok {
		for pid := range bidSet {
			resp.PIDs = append(resp.PIDs, pid)
		}
	}

	respJSON, err := json.MarshalIndent(resp, "", " ")
	writeJson(w, respJSON, err)
	return
}

func openGet(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusBadRequest)
		return false
	}
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Malformed URL.", http.StatusBadRequest)
		return false
	}

	return true
}

func writeJson(w http.ResponseWriter, body []byte, err error) {
	if err != nil {
		http.Error(w, "response creation failure: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, "failed sending response: "+err.Error(), http.StatusInternalServerError)
	}
	return
}
