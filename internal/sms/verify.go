package sms

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
)

func verifyTelnyxSignature(pubKey ed25519.PublicKey, sigB64, timestamp string, body []byte) bool {
	if sigB64 == "" || timestamp == "" {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false
	}
	message := []byte(timestamp + "|")
	message = append(message, body...)
	return ed25519.Verify(pubKey, message, sig)
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
