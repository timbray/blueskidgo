package blueskidgo

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
)

func keyToString(key ed25519.PublicKey) (string, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func stringToKey(s string) (ed25519.PublicKey, error) {
	// pem package apparently needs ASCII armor
	if !strings.Contains(s, "-----BEGIN") {
		s = "-----BEGIN PUBLIC KEY-----\n" + s + "\n-----END PUBLIC KEY-----"
	}
	block, _ := pem.Decode([]byte(s))

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ek, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not an encoded ed25519 public key encoding")
	}
	return ek, nil
}
