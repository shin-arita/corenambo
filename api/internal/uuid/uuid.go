package uuid

import gouuid "github.com/google/uuid"

var newV7 = gouuid.NewV7

type UUIDv7Generator struct{}

func (g UUIDv7Generator) NewV7() (string, error) {
	id, err := newV7()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
