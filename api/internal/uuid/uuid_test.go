package uuid

import (
	"errors"
	"testing"

	gouuid "github.com/google/uuid"
)

func TestUUIDv7GeneratorNewV7(t *testing.T) {
	g := UUIDv7Generator{}

	id, err := g.NewV7()
	if err != nil {
		t.Fatal(err)
	}

	if id == "" {
		t.Fatal("id is empty")
	}
}

func TestUUIDv7GeneratorNewV7Error(t *testing.T) {
	originalNewV7 := newV7
	newV7 = func() (gouuid.UUID, error) {
		return gouuid.UUID{}, errors.New("uuid failed")
	}
	defer func() {
		newV7 = originalNewV7
	}()

	g := UUIDv7Generator{}

	id, err := g.NewV7()
	if err == nil {
		t.Fatal("expected error")
	}

	if id != "" {
		t.Fatalf("unexpected id: %s", id)
	}
}
