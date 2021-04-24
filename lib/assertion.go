package blueskidgo

// generates and verifies BID-grant assertions for the @bluesky Identity protocol

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Guitar            = "🎸"
	Drum              = "🥁"
	Opcode            = 0
	BID               = 1
	ClaimNonce        = 2
	ClaimKey          = 3
	ClaimSig          = 4
	ClaimCounterparty = 5
)

// An grantAssertion in general has the syntax
// Drum data Guitar data … Guitar data Drum
// To use it, strip the Drum begin/end markers, and split the remainder in Guitar to isolate the fields
func findBlueskidAssertion(text string, fieldCount int) ([]string, error) {
	if strings.Count(text, Drum) != 2 {
		return nil, errors.New("text does not contain a Blueskid grantAssertion")
	}
	text = text[strings.Index(text, Drum)+len(Drum):]
	text = text[:strings.Index(text, Drum)]
	fields := strings.SplitN(text, Guitar, fieldCount)
	if len(fields) != fieldCount {
		return nil, errors.New(fmt.Sprintf("wrong number of fields (%d requested, %d found)", fieldCount, len(fields)))
	}
	return fields, nil
}

// generateGrantAssertions Generates two strings that represent, respectively, the holder of a BID granting it to
//  another PID, and the PID accepting the grant. Let's call the two strings grant and Accept
//  Each post has the syntax ga/BID/nonce/key/sig/counterparty, where
//  - ga is either G for grant or A for Accept
//  - BID is the Bluesky ID, base64 of an unsigned 64-bit identifier
//  - Nonce takes the form of an RFC3339 date stamp followed by .G in a grant, .A in an accept
//  - The key is the conventional representation of an ed25119 public key
//  - The sig is the base64 encoding in the bits of the signature generated by the private key
//  - counterparty is the receiving PID in a grant, the granting party in an Accept. This is provided
//    last in case to allow parsing with strings.split(), the separator character could appear in the PID
//
func generateGrantAssertions(bid uint64, granter string, accepter string) (grant string, accept string, err error) {

	bidString := fmt.Sprintf("%X", bid)

	// get a timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// a keypair
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return // I like naked returns and I cannot lie
	}
	pubString, err := KeyToString(public)
	if err != nil {
		return
	}

	nonce := timestamp + ".G"
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(private, []byte(nonce)))
	grant = assertionFromFields("G", bidString, nonce, pubString, sig, accepter)

	nonce = timestamp + ".A"
	sig = base64.StdEncoding.EncodeToString(ed25519.Sign(private, []byte(nonce)))
	accept = assertionFromFields("A", bidString, nonce, pubString, sig, granter)

	// TODO: Figure out how to overwrite this so it doesn't linger in memory
	private = nil

	return
}

func assertionFromFields(fields ...string) string {

	s := Drum + fields[0]
	for i := 1; i < len(fields); i++ {
		s += Guitar + fields[i]
	}
	return s + Drum
}

func checkGrantAssertionPair(gFields []string, gPID string, aFields []string, aPID string) (uint64, error) {

	granter, err := checkGrantAssertion(gFields)
	if err != nil {
		return 0, errors.New("invalid granter assertion: " + err.Error())
	}

	accepter, err := checkGrantAssertion(aFields)
	if err != nil {
		return 0, errors.New("invalid accepter assertion: " + err.Error())
	}

	if !granter.pubKey.Equal(accepter.pubKey) {
		return 0, errors.New("granter and accepter not signed with same key")
	}
	if granter.nonce == accepter.nonce {
		return 0, errors.New("granter and accepter used same nonce")
	}
	if granter.bid != accepter.bid {
		return 0, errors.New("granter and accepter BIDs differ")
	}

	if accepter.counterparty != gPID {
		return 0, errors.New("a  ccepter assertion does not identify granter")
	}
	if granter.counterparty != aPID {
		return 0, errors.New("granter assertion does not identify accepter")
	}

	return granter.bid, nil
}

type grantAssertion struct {
	ga           string
	bid          uint64
	counterparty string
	nonce        string
	pubKey       ed25519.PublicKey
}

func checkGrantAssertion(parts []string) (*grantAssertion, error) {

	var a grantAssertion

	ga := parts[Opcode]
	if !(ga == "G" || ga == "A") {
		return nil, errors.New("grant/Accept must begin with either 'A' or 'G'")
	}
	a.ga = ga

	bid, err := strconv.ParseUint(parts[BID], 16, 64)
	if err != nil {
		return nil, err
	}
	a.bid = bid

	nonce := parts[ClaimNonce]
	suffix := nonce[len(nonce)-2:]
	if !(suffix == ".A" || suffix == ".G") {
		return nil, errors.New("malformed nonce, should end with .G or .A")
	}
	if (ga == "G" && suffix != ".G") || (ga == "A" && suffix != ".A") {
		return nil, errors.New("nonce suffix should match ga")
	}
	a.nonce = nonce

	key, err := StringToKey(parts[ClaimKey])
	if err != nil {
		return nil, errors.New("can't parse public key in grantAssertion: " + err.Error())
	}
	a.pubKey = key

	sig, err := base64.StdEncoding.DecodeString(parts[ClaimSig])
	if err != nil {
		return nil, errors.New("malformed signature in assertino: " + parts[ClaimSig] + " - " + err.Error())
	}

	if !ed25519.Verify(key, []byte(nonce), sig) {
		return nil, errors.New("grantAssertion signature vaildation failed")
	}

	a.counterparty = parts[ClaimCounterparty]

	return &a, nil
}

func fetchAssertionFromPost(url string, fieldCount int) (assertionFields []string, pid string, err error) {
	if strings.HasPrefix(url, "https://twitter.com/") {
		assertionFields, pid, err = assertionFromTwitter(url, fieldCount)
	} else {
		err = errors.New("no handler for social-media URL " + url)
	}
	return
}
