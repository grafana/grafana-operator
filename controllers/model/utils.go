package model

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func RandStringRunes(s int) string {
	b := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)
}

func MergeAnnotations(requested map[string]string, existing map[string]string) map[string]string {
	if existing == nil {
		return requested
	}

	for k, v := range requested {
		existing[k] = v
	}
	return existing
}

func BoolPtr(b bool) *bool { return &b }

func IntPtr(b int64) *int64 { return &b }
