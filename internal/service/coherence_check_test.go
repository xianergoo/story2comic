package service

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckCoherenceReturnsErrorWhenProviderMissing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	svc := NewOutlineService(db)

	result, err := svc.CheckCoherence("prev", "characters", "world", "chapter")
	if err == nil {
		t.Fatal("expected error when provider is missing")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected provider error, got %v", err)
	}
}
