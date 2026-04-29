package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func RandHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func SignConnectPayload(appId uint64, secret string, timestamp int64, nonce string) string {
	payload := fmt.Sprintf("%d:%d:%s", appId, timestamp, nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(appId uint64, secret string, timestamp int64, nonce, sig string) error {
	ts := time.Unix(timestamp, 0)
	if diff := time.Since(ts); diff > 5*time.Minute || diff < -5*time.Minute {
		return fmt.Errorf("timestamp out of window: %v", diff)
	}

	expected := SignConnectPayload(appId, secret, timestamp, nonce)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
