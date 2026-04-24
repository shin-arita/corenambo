package clock

import "testing"

func TestSystemClockNow(t *testing.T) {
	c := SystemClock{}

	if c.Now().IsZero() {
		t.Fatal("now is zero")
	}
}
