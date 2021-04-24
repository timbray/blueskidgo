package blueskidgo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type grantAssertionsRequest struct {
	BID      string
	Granter  string
	Accepter string
}
type grantAssertionsResponse struct {
	GrantAssertion  string
	AcceptAssertion string
}
type bidAssertionRequest struct {
	BID string
}
type bidAssertionResponse struct {
	Assertion string
}

func ClaimAssertionsHandler(w http.ResponseWriter, httpRequest *http.Request) {
	bidAssertionHandler(w, httpRequest, "C")
}
func UnclaimAssertionsHandler(w http.ResponseWriter, httpRequest *http.Request) {
	bidAssertionHandler(w, httpRequest, "U")
}

func bidAssertionHandler(w http.ResponseWriter, httpRequest *http.Request, opcode string) {
	if httpRequest.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(httpRequest.Body)
	if err != nil {
		http.Error(w, "Can't read request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var req bidAssertionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Can't parse JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := bidAssertionResponse{Assertion: assertionFromFields(opcode, req.BID)}
	respJSON, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		http.Error(w, "Can't generate JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respJSON)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
	}
	return
}

func GrantAssertionsHandler(w http.ResponseWriter, httpRequest *http.Request) {
	if httpRequest.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(httpRequest.Body)
	if err != nil {
		http.Error(w, "Can't read request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, problem, myFault := newGrantAssertionsResponse(body)

	if problem != "" {
		if myFault {
			http.Error(w, problem, http.StatusInternalServerError)
		} else {
			http.Error(w, problem, http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
	}
	return
}

// this is broken out so it can be tested
func newGrantAssertionsResponse(reqBody []byte) (resp []byte, msg string, myProblem bool) {
	var req grantAssertionsRequest
	err := json.Unmarshal(reqBody, &req)
	if err != nil {
		msg = "Can't parse JSON body: " + err.Error()
		return
	}

	if req.BID == "" || req.Accepter == "" || req.Granter == "" {
		msg = "Missing fields in JSON body"
		return
	}

	bid, err := strconv.ParseUint(req.BID, 16, 64)
	if err != nil {
		msg = "BID isn't a 64-bit quantity: " + err.Error()
		return
	}

	g, a, err := generateGrantAssertions(bid, req.Granter, req.Accepter)
	if err != nil {
		myProblem = true
		msg = "Assertion generation error: " + err.Error()
		return
	}

	response := grantAssertionsResponse{GrantAssertion: g, AcceptAssertion: a}
	respJSON, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		myProblem = true
		msg = "JSON encoding error: " + err.Error()
		return
	}

	resp = respJSON
	return
}
