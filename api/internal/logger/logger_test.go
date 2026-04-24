package logger

import "testing"

func TestInfo(t *testing.T) {
	Info("test %s", "info")
}

func TestWarn(t *testing.T) {
	Warn("test %s", "warn")
}

func TestError(t *testing.T) {
	Error("test %s", "error")
}
