package authutil

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken() string {
	tok := make([]byte, 32)
	rand.Read(tok)
	return base64.URLEncoding.EncodeToString(tok)
}
