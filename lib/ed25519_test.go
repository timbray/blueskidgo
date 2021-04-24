package blueskidgo

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestInterchangeOneThousandKeys(t *testing.T) {
	for i := 0; i < 1000; i++ {
		keyIn, _, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Error("Generate: " + err.Error())
		}

		s, err := KeyToString(keyIn)
		if err != nil {
			t.Error("k2s: " + err.Error())
		}
		keyOut, err := StringToKey(s)
		if err != nil {
			t.Error("s2k: " + err.Error())
		}
		if bytes.Compare(keyIn, keyOut) != 0 {
			t.Error("Keys not equal")
		}
	}
}

// Now prove we can unpack and use a key and sig from a Java program
func TestExternallyGeneratedKey(t *testing.T) {
	nonce := "f571bbdc-bfed-4edd-a9bc-e375783da846"
	keyText := "MCowBQYDK2VwAyEAZwCr+eQWBM0P8LuQr0l+TM4ZyEKD2IJ7vfN1I6qK1bY="
	sigText := "RYjSsXJts00FnI9bKIya979tTzbKNNUvD+DoWTwSTn50j5hYV6kYKbaMo9r52TGGnnZkw0II4Jage7WJ90enBQ=="

	key, err := StringToKey(keyText)
	if err != nil {
		t.Error("s2K " + err.Error())
	}
	sig, err := base64.StdEncoding.DecodeString(sigText)
	if err != nil {
		t.Error("Decode " + err.Error())
	}
	if !ed25519.Verify(key, []byte(nonce), sig) {
		t.Error("Not verified")
	}
}
