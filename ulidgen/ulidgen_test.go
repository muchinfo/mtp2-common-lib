package ulidgen

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenerateULID(t *testing.T) {
	id, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID error: %v", err)
	}
	if len(id) != 26 {
		t.Errorf("ULID length should be 26, got %d", len(id))
	}
}

func TestGenerateShortULID(t *testing.T) {
	id, err := GenerateShortULID(18)
	if err != nil {
		t.Fatalf("GenerateShortULID error: %v", err)
	}
	if len(id) != 18 {
		t.Errorf("Short ULID length should be 18, got %d", len(id))
	}

	id2, err := GenerateShortULID(30)
	if err != nil {
		t.Fatalf("GenerateShortULID error: %v", err)
	}
	if len(id2) != 26 {
		t.Errorf("Short ULID length should be capped at 26, got %d", len(id2))
	}
}

func TestGenerateULIDWithPrefix(t *testing.T) {
	prefix := "JYI"
	maxLen := 20
	id, err := GenerateULIDWithPrefix(prefix, maxLen)
	if err != nil {
		t.Fatalf("GenerateULIDWithPrefix error: %v", err)
	}
	if !strings.HasPrefix(id, prefix) {
		t.Errorf("ULID should have prefix %s", prefix)
	}
	if len(id) != maxLen {
		t.Errorf("ULID with prefix length should be %d, got %d", maxLen, len(id))
	}
}

func TestGenerateULIDWithRandSource(t *testing.T) {
	tm := time.Date(2024, 7, 21, 12, 0, 0, 0, time.UTC)
	r := rand.New(rand.NewSource(42))
	id := GenerateULIDWithRandSource(tm, r)
	if len(id) != 26 {
		t.Errorf("ULID length should be 26, got %d", len(id))
	}
}
