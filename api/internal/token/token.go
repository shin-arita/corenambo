package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
)

var randReader io.Reader = rand.Reader

type DefaultGenerator struct{}

func (g DefaultGenerator) Generate() (string, error) {
	b := make([]byte, 32)
	if _, err := randReader.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

type SHA256Hasher struct{}

func (h SHA256Hasher) Hash(value string) (string, error) {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:]), nil
}
