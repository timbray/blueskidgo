package blueskidgo

import (
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"strings"
)

func keyToString(key ed25519.PublicKey) (string, error) {
	type pkInfo struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	info := pkInfo{
		Algorithm: pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 3, 101, 112}},
		PublicKey: asn1.BitString{BitLength: len(key), Bytes: key},
	}
	asnBytes, err := asn1.Marshal(info)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(asnBytes), nil
}

func stringToKey(s string) (ed25519.PublicKey, error) {
	// pem package apparently needs ASCII armor
	if !strings.Contains(s, "-----BEGIN") {
		s = "-----BEGIN PUBLIC KEY-----\n" + s + "\n-----END PUBLIC KEY-----"
	}
	block, _ := pem.Decode([]byte(s))

	// TODO: Figure out why the struct used for marshaling above is different than this one used for unmarshalling
	type pubKey struct {
		OBjectIdentifier struct {
			ObjectIdentifier asn1.ObjectIdentifier
		}
		PublicKey asn1.BitString
	}
	var pk pubKey
	_, err := asn1.Unmarshal(block.Bytes, &pk)
	if err != nil {
		return nil, err
	}

	return pk.PublicKey.Bytes, nil
}
