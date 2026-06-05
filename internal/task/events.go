package task

import (
	"encoding/json"
	"strings"
)

// SSETransportEventName is the fixed SSE event name used on the wire.
const SSETransportEventName = "progress"

// EventType is the business event type carried inside the SSE JSON payload.
type EventType string

const (
	EventTypeNovelStatus     EventType = "novel_status"
	EventTypeProgressSummary EventType = "progress_summary"
	EventTypeChapterStatus   EventType = "chapter_status"
	EventTypeChapterStream   EventType = "chapter_stream"
	EventTypeOutlineStream   EventType = "outline_stream"
	EventTypeComicStatus     EventType = "comic_status"
	EventTypeLog             EventType = "log"
	EventTypeError           EventType = "error"
)

// Event is the common SSE payload envelope.
//
// Type is the business dispatch key. Payload carries event-specific content.
type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload,omitempty"`
}

func NewEvent(eventType EventType, payload any) Event {
	return Event{Type: eventType, Payload: payload}
}

// NormalizeEventType maps both canonical and legacy event type names to a single source of truth.
func NormalizeEventType(typeName string) (EventType, bool) {
	switch EventType(typeName) {
	case EventTypeNovelStatus,
		EventTypeProgressSummary,
		EventTypeChapterStatus,
		EventTypeChapterStream,
		EventTypeOutlineStream,
		EventTypeComicStatus,
		EventTypeLog,
		EventTypeError:
		return EventType(typeName), true
	}

	switch typeName {
	case "progress":
		return EventTypeProgressSummary, true
	case "chapter":
		return EventTypeChapterStatus, true
	case "comic":
		return EventTypeComicStatus, true
	case "error":
		return EventTypeError, true
	default:
		return "", false
	}
}

// EventFromLegacyMessage wraps historical Push(string) payloads into the structured event envelope.
//
// It prefers preserving a stable business type even when the original JSON is malformed,
// because upstream frequently hand-builds JSON strings.
func EventFromLegacyMessage(message string) Event {
	if canonicalType, ok := legacyTypeFromMessage(message); ok {
		var raw map[string]any
		if err := json.Unmarshal([]byte(message), &raw); err == nil {
			if payload, hasPayload := raw["payload"]; hasPayload && len(raw) == 2 {
				return NewEvent(canonicalType, payload)
			}

			delete(raw, "type")
			if len(raw) == 0 {
				return NewEvent(canonicalType, nil)
			}

			return NewEvent(canonicalType, raw)
		}

		return NewEvent(canonicalType, map[string]any{"raw": message})
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(message), &raw); err == nil {
		return NewEvent(EventTypeLog, raw)
	}

	return NewEvent(EventTypeLog, map[string]any{"message": message})
}

func legacyTypeFromMessage(message string) (EventType, bool) {
	const marker = `"type":"`
	start := strings.Index(message, marker)
	if start < 0 {
		return "", false
	}

	start += len(marker)
	end := strings.Index(message[start:], `"`)
	if end < 0 {
		return "", false
	}

	return NormalizeEventType(message[start : start+end])
}
