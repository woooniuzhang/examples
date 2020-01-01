package prometheus

import "testing"

func TestPull(t *testing.T) {
	Pull()
}

// go test -v pull_test.go push.go pull.go
