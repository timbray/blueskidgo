package blueskidgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type bidRequest struct {
	Post string
}

type grantRequest struct {
	GrantPost  string
	AcceptPost string
}

// no response bodies to the BID-update calls

func ClaimBIDHandler(w http.ResponseWriter, httpRequest *http.Request) {
	body := openPost(w, httpRequest)
	if body == nil {
		return
	}
	var req bidRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Can't parse JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	fields, pid, err := fetchAssertionFromPost(req.Post, 2)
	if err != nil {
		http.Error(w, "Failed to find assertion: "+err.Error(), http.StatusBadRequest)
		return
	}

	if fields[0] != "C" {
		http.Error(w, "not a BID Claim assertion", http.StatusBadRequest)
		return
	}

	bid, err := strconv.ParseUint(fields[1], 16, 64)
	if err != nil {
		http.Error(w, "BID in Claim assertion is not a hex 64-bit quantity", http.StatusBadRequest)
		return
	}
	err = appendToLedger(&LedgerRecord{
		RecType:  ClaimBID,
		BID:      fmt.Sprintf("%016X", bid),
		PIDs:     []string{pid},
		PostURLs: []string{req.Post},
	})

	if err != nil {
		http.Error(w, "Database update rejected: "+err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func GrantBIDHandler(w http.ResponseWriter, httpRequest *http.Request) {
	body := openPost(w, httpRequest)
	if body == nil {
		return
	}
	var req grantRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Can't parse JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}
	gFields, gPID, err := fetchAssertionFromPost(req.GrantPost, 6)
	if err != nil {
		http.Error(w, "Failed to fetch assertion: "+err.Error(), http.StatusBadRequest)
		return
	}
	aFields, aPID, err := fetchAssertionFromPost(req.AcceptPost, 6)
	if err != nil {
		http.Error(w, "Failed to fetch assertion: "+err.Error(), http.StatusBadRequest)
		return
	}

	bid, err := checkGrantAssertionPair(gFields, gPID, aFields, aPID)
	if err != nil {
		http.Error(w, "grant and accept assertions invalid: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = appendToLedger(&LedgerRecord{
		RecType:  GrantBID,
		BID:      fmt.Sprintf("%016X", bid),
		PIDs:     []string{gPID, aPID},
		PostURLs: []string{req.GrantPost, req.AcceptPost},
	})
	if err != nil {
		http.Error(w, "Database update rejected: "+err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func UnclaimBIDHandler(w http.ResponseWriter, httpRequest *http.Request) {
	body := openPost(w, httpRequest)
	if body == nil {
		return
	}
	var req bidRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Can't parse JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	fields, pid, err := fetchAssertionFromPost(req.Post, 2)
	if err != nil {
		http.Error(w, "Failed to find assertion: "+err.Error(), http.StatusBadRequest)
		return
	}

	if fields[0] != "U" {
		http.Error(w, "not a BID Unclaim assertion", http.StatusBadRequest)
		return
	}

	bid, err := strconv.ParseUint(fields[1], 16, 64)
	if err != nil {
		http.Error(w, "BID in Unclaim assertion is not a hex 64-bit quantity", http.StatusBadRequest)
		return
	}
	err = appendToLedger(&LedgerRecord{
		RecType:  UnclaimBID,
		BID:      fmt.Sprintf("%016X", bid),
		PIDs:     []string{pid},
		PostURLs: []string{req.Post},
	})
	if err != nil {
		http.Error(w, "Database update rejected: "+err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func openPost(w http.ResponseWriter, req *http.Request) []byte {
	if req.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusBadRequest)
		return nil
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Can't read request body: "+err.Error(), http.StatusInternalServerError)
		return nil
	}
	return body
}
