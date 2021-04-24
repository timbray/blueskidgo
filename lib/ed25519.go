package blueskidgo

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

func KeyToString(key ed25519.PublicKey) (string, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func StringToKey(s string) (ed25519.PublicKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}
	ek, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not an encoded ed25519 public key encoding")
	}
	return ek, nil
}
