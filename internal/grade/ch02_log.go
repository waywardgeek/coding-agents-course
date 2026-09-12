package grade

// Chapter 2 log fixtures and normalization.
//
// NEVER GRADE ON GO IDENTIFIERS. Event type names and field names are compared
// lowercased with punctuation stripped, so ToolCalled, tool_called and
// TOOL-CALLED are the same name, and CallID, call_id and callid are the same
// field. Failing someone for Kind instead of Type is not a lesson.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExhibitLog is the log the seam is graded on. One conversation containing
// every shape the chapter's exhibits need:
//
//   - a human message
//   - an assistant turn with opaque replay material, text, and a tool call
//   - a tool result (Exhibit A: three authorships)
//   - a following human message (Exhibit B: merged by Anthropic, separate
//     everywhere else)
//
// The opaque block is tagged with the SAME model the grader renders for, so a
// correct Anthropic render replays it and a correct OpenAI or Gemini render
// omits it — material goes back only to the exact model that issued it.
const ExhibitLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"read config.json"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"opaque","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"data":{"type":"thinking","thinking":"The user wants the config file.","signature":"sig-exhibit-1"}},{"type":"text","text":"I'll read it."},{"type":"tool_call","call_id":"toolu_exhibit_1","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"name":"read_file","args":{"path":"config.json","limit":40}}],"usage":{"input":100,"cache_write":0,"cache_read":50,"output":20},"from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"toolu_exhibit_1","name":"read_file","args":{"path":"config.json","limit":40}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"toolu_exhibit_1","parts":[{"type":"text","text":"port=8080\nhost=localhost"}]}}
{"seq":5,"type":"message_received","time":"2026-01-01T00:00:04Z","message":{"actor":"human","parts":[{"type":"text","text":"now check the logs instead"}]}}
`

// RedactionLog is ExhibitLog plus a Redacted event superseding the tool
// result. Chapter 2 exercises only RedactResult: the RESULT becomes a stub and
// the CALL survives, so the model can still see what it asked for and why.
const RedactionLog = ExhibitLog + `{"seq":6,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}
`

// RedactedSecret is the content that must NOT survive a redaction.
const RedactedSecret = "port=8080"

// Ch2LogLine is one dumped event, with keys normalized.
type Ch2LogLine struct {
	Seq  int
	Type string
	Data map[string]any
}

func parseLogLines(dump string) ([]Ch2LogLine, string) {
	var out []Ch2LogLine
	for i, line := range strings.Split(dump, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return out, fmt.Sprintf("line %d of dump is not JSON: %v", i+1, err)
		}
		m := normKeys(raw).(map[string]any)
		// A header line carrying only a format version is not an event.
		if m["type"] == nil && m["seq"] == nil {
			continue
		}
		l := Ch2LogLine{Data: m}
		if s, ok := m["type"].(string); ok {
			l.Type = normName(s)
		}
		if f, ok := m["seq"].(float64); ok {
			l.Seq = int(f)
		}
		out = append(out, l)
	}
	return out, ""
}

// normName lowercases and strips everything that is not a letter or digit.
func normName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normKeys rewrites every map key in a decoded JSON tree to its normalized
// form, recursively, so lookups never depend on a student's spelling.
func normKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[normName(k)] = normKeys(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = normKeys(val)
		}
		return out
	default:
		return v
	}
}

// get looks up the first present of several normalized field names.
func get(m map[string]any, names ...string) any {
	for _, n := range names {
		if v, ok := m[normName(n)]; ok {
			return v
		}
	}
	return nil
}

func getMap(m map[string]any, names ...string) map[string]any {
	if v, ok := get(m, names...).(map[string]any); ok {
		return v
	}
	return nil
}

func getSlice(m map[string]any, names ...string) []any {
	if v, ok := get(m, names...).([]any); ok {
		return v
	}
	return nil
}

func getStr(m map[string]any, names ...string) string {
	if v, ok := get(m, names...).(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, names ...string) (int, bool) {
	if v, ok := get(m, names...).(float64); ok {
		return int(v), true
	}
	return 0, false
}

// canonicalJSON re-serializes a decoded value with map keys sorted, so that
// `{"a":1,"b":2}` and `{"b":2,"a":1}` compare equal, and so that OpenAI's
// JSON-encoded argument STRING compares equal to Anthropic's nested object
// once the student has decoded it.
func canonicalJSON(v any) string {
	if s, ok := v.(string); ok {
		var inner any
		if json.Unmarshal([]byte(s), &inner) == nil {
			v = inner
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// SaidProjection reduces a dumped log to "everything the model said", with the
// things that legitimately differ between vendors removed:
//
//   - provenance (checked separately, and required to survive)
//   - tool-call ids, which are issued by the vendor and cannot match
//   - opaque replay material, which is bound to one model by definition
//   - token accounting, which is graded by `usage`
//
// Usage is deliberately NOT compared here. It is parsing work, and it is real
// parsing work — but it has its own check so that a student who gets the
// message shapes right and the accounting wrong is told which half failed,
// instead of losing two large blocks to one cause.
//
// What remains must be identical across all three vendors. If it is not, the
// student has leaked vendor shape past the parser — and leaked vendor shape is
// precisely what makes the second implementation a copy-paste.
func SaidProjection(lines []Ch2LogLine) []string {
	var out []string
	for _, l := range lines {
		switch l.Type {
		case normName("response_ended"):
			resp := getMap(l.Data, "response")
			if resp == nil {
				out = append(out, "response_ended:<no payload>")
				continue
			}
			out = append(out, "response:"+partsProjection(getSlice(resp, "parts")))
		case normName("message_received"):
			msg := getMap(l.Data, "message")
			if msg == nil {
				continue
			}
			out = append(out, "message:"+getStr(msg, "actor")+":"+partsProjection(getSlice(msg, "parts")))
		}
	}
	return out
}

func partsProjection(parts []any) string {
	var bits []string
	for _, p := range parts {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		switch normName(getStr(m, "type", "kind")) {
		case normName("tool_call"):
			// Name and arguments must normalize. The ID must not: it is
			// vendor-issued. Its PRESENCE is checked elsewhere.
			bits = append(bits, "call("+getStr(m, "name")+","+canonicalJSON(get(m, "args", "arguments", "input"))+")")
		case normName("tool_result"):
			bits = append(bits, "result")
		case normName("opaque"):
			// Excluded: bound to one model by definition.
		case normName("redacted"):
			bits = append(bits, "redacted")
		default:
			if t := getStr(m, "text"); t != "" {
				bits = append(bits, "text("+t+")")
			}
		}
	}
	return strings.Join(bits, "|")
}

func usageProjection(u map[string]any) string {
	if u == nil {
		return "<none>"
	}
	f := func(names ...string) string {
		if n, ok := getInt(u, names...); ok {
			return fmt.Sprint(n)
		}
		return "?"
	}
	return strings.Join([]string{
		f("input"), f("cache_write"), f("cache_read"), f("output"),
	}, "/")
}
