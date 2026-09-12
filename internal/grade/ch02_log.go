package grade

// Reading the submission's event log.
//
// The grader has to make assertions about events the student defined, which
// means agreeing on names. Rather than demand an exact spelling, every type
// name is normalized: lowercased with non-alphanumerics removed, so
// "ToolCalled", "tool_called" and "TOOL-CALLED" are the same event. The
// vocabulary itself (the set of names in the outline's taxonomy) still has to
// be honoured, and the chapter has to say so.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LogEvent is one line of the dumped JSON-lines event log.
type LogEvent struct {
	Line   string         // the raw line, for substring searches and diagnostics
	Seq    int            // the event's sequence number
	SeqOK  bool           // whether a seq field was present and numeric
	Type   string         // the type as written
	Norm   string         // normalized type
	Fields map[string]any // the whole object, decoded
}

// Has reports whether the raw line contains the given payload.
func (e LogEvent) Has(needle string) bool { return strings.Contains(e.Line, needle) }

// normType lowercases and strips punctuation from an event type name.
func normType(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// dialogueTypes are the events that carry conversation content. Ephemera must
// never be one of these.
var dialogueTypes = map[string]bool{
	"messagereceived":  true,
	"assistantmessage": true,
	"assistantthought": true,
	"toolcalled":       true,
	"toolreturned":     true,
}

// ParseLog reads a JSON-lines event log.
func ParseLog(path string) ([]LogEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []LogEvent
	for i, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		ev := LogEvent{Line: line}
		if err := json.Unmarshal([]byte(line), &ev.Fields); err != nil {
			return nil, fmt.Errorf("line %d is not JSON: %v", i+1, err)
		}
		if v, ok := ev.Fields[findKey(ev.Fields, "type")]; ok {
			if s, ok := v.(string); ok {
				ev.Type = s
				ev.Norm = normType(s)
			}
		}
		if v, ok := ev.Fields[findKey(ev.Fields, "seq")]; ok {
			if f, ok := v.(float64); ok {
				ev.Seq = int(f)
				ev.SeqOK = true
			}
		}
		out = append(out, ev)
	}
	return out, nil
}

// findKey returns the actual key in m whose normalized form matches want, or
// "" if there is none. Lets the grader accept "target_seq", "targetSeq" and
// "TargetSeq" without argument.
func findKey(m map[string]any, want string) string {
	for k := range m {
		if normType(k) == normType(want) {
			return k
		}
	}
	return ""
}

// intField pulls an integer field by normalized name.
func intField(m map[string]any, names ...string) (int, bool) {
	for _, n := range names {
		if k := findKey(m, n); k != "" {
			if f, ok := m[k].(float64); ok {
				return int(f), true
			}
		}
	}
	return 0, false
}

// FindEvents returns every event whose normalized type matches.
func FindEvents(log []LogEvent, norm string) []LogEvent {
	var out []LogEvent
	for _, e := range log {
		if e.Norm == norm {
			out = append(out, e)
		}
	}
	return out
}

// FindPayload returns the first event whose raw line contains the payload,
// preferring dialogue events (the redaction target is a message, not the
// RequestSent that happened to carry it).
func FindPayload(log []LogEvent, payload string) (LogEvent, bool) {
	for _, e := range log {
		if e.Has(payload) && dialogueTypes[e.Norm] {
			return e, true
		}
	}
	for _, e := range log {
		if e.Has(payload) {
			return e, true
		}
	}
	return LogEvent{}, false
}
