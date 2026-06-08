package handler

import (
	"errors"
	"testing"

	"github.com/gin-contrib/sessions"
)

type stubSession struct {
	err error
}

func (s *stubSession) ID() string                                       { return "" }
func (s *stubSession) Get(key interface{}) interface{}                 { return nil }
func (s *stubSession) Set(key interface{}, val interface{})            {}
func (s *stubSession) Delete(key interface{})                          {}
func (s *stubSession) Clear()                                          {}
func (s *stubSession) AddFlash(value interface{}, vars ...string)      {}
func (s *stubSession) Flashes(vars ...string) []interface{}            { return nil }
func (s *stubSession) Options(options sessions.Options)                {}
func (s *stubSession) Save() error                                     { return s.err }

func TestSaveSessionReturnsWrappedError(t *testing.T) {
	err := saveSession(&stubSession{err: errors.New("boom")})
	if err == nil {
		t.Fatal("expected saveSession to return error")
	}
	if err.Error() != "save session: boom" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveSessionSucceeds(t *testing.T) {
	if err := saveSession(&stubSession{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
