package token

import (
	"encoding/base64"
	"errors"
	"testing"
)

type errorReader struct{}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestDefaultGeneratorGenerate(t *testing.T) {
	g := DefaultGenerator{}

	token, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("token is empty")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatal(err)
	}

	if len(decoded) != 32 {
		t.Fatalf("expected decoded token length 32, got %d", len(decoded))
	}
}

func TestDefaultGeneratorGenerateError(t *testing.T) {
	originalRandReader := randReader
	randReader = &errorReader{}
	defer func() {
		randReader = originalRandReader
	}()

	g := DefaultGenerator{}

	token, err := g.Generate()
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestSHA256HasherHash(t *testing.T) {
	h := SHA256Hasher{}

	got, err := h.Hash("test")
	if err != nil {
		t.Fatal(err)
	}

	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
