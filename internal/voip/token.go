package voip

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const tokenTTL = time.Hour

type VertoCredentials struct {
	Token     string
	Username  string
	Password  string
	ExpiresAt time.Time
}

// IssueVertoCredentials generates short-lived FreeSWITCH Verto credentials.
func IssueVertoCredentials(userID, secret string) (*VertoCredentials, error) {
	nonce, err := generateHex(8)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(tokenTTL)
	expStr := fmt.Sprintf("%d", expiresAt.Unix())

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID + ":" + nonce + ":" + expStr))
	password := hex.EncodeToString(mac.Sum(nil))

	return &VertoCredentials{
		Token:     nonce,
		Username:  userID + "_" + nonce,
		Password:  password,
		ExpiresAt: expiresAt,
	}, nil
}

func generateHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
