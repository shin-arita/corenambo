package uuid

type Generator interface {
	NewV7() (string, error)
}
