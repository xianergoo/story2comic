package app

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"novelforge/internal/config"
)

func TestNewBuildsApplicationFromExplicitConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Port:          "18080",
		DBPath:        filepath.Join(tempDir, "novelforge.db"),
		ImageDir:      filepath.Join(tempDir, "images"),
		SessionSecret: "test-session-secret",
		MaxChapters:   5,
		ImageEnabled:  false,
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	if application.Config != cfg {
		t.Fatal("expected app to keep explicit config reference")
	}
	if application.Services == nil {
		t.Fatal("expected services to be initialized")
	}
	if application.Runtime == nil {
		t.Fatal("expected runtime to be initialized")
	}
	if application.Router == nil {
		t.Fatal("expected router to be initialized")
	}
	if application.Services.Outline == nil || application.Services.Chapter == nil {
		t.Fatal("expected core services to be initialized")
	}
	if application.Runtime.TaskQueue == nil || application.Runtime.SSEHandler == nil {
		t.Fatal("expected runtime dependencies to be initialized")
	}

	novelValue := reflect.ValueOf(application.Services.Novel).Elem()
	if novelValue.FieldByName("outlineSvc").IsNil() {
		t.Fatal("expected novel service to have outline service dependency")
	}
	if novelValue.FieldByName("chapterSvc").IsNil() {
		t.Fatal("expected novel service to have chapter service dependency")
	}
	if novelValue.FieldByName("comicSvc").IsNil() {
		t.Fatal("expected novel service to have comic service dependency")
	}
	if novelValue.FieldByName("taskQ").IsNil() {
		t.Fatal("expected novel service to have task queue dependency")
	}
}

func TestApplicationShutdownIsSafeToCallRepeatedly(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Port:          "18080",
		DBPath:        filepath.Join(tempDir, "novelforge.db"),
		ImageDir:      filepath.Join(tempDir, "images"),
		SessionSecret: "test-session-secret",
		MaxChapters:   5,
		ImageEnabled:  false,
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}
	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("second shutdown: %v", err)
	}
}
