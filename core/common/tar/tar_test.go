package tar

import (
	"testing"
)

func TestTarDir(t *testing.T) {

	exc := NewExclusions(".pfx", ".json")

	if len(exc) != 2 {
		t.Fatalf("Invalid length")
	}

	if !exc.Contains(".pfx") {
		t.Fatalf("Should exclude .pfx")
	}

	if exc.Contains(".txt") {
		t.Fatalf("Should not exclude .txt")
	}

	if exc.Contains("") {
		t.Fatalf("Should not exclude empty extension")
	}
}
