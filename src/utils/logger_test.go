package utils

import "testing"

func TestNewLogger(t *testing.T) {
	l := NewLogger("[test] ", "./test.log")
	l.Println("a test log")
}
