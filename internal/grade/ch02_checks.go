package grade

// The Chapter 2 checks.
//
// Chapter 1 graded a wire format. Chapter 2 grades a reducer and a renderer,
// so almost nothing carries over except the discipline: every failure names
// the property that failed and says what to do about it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakeanthropic"
)

// Ch2Evaluate turns Chapter 2 run evidence into the graded checklist.
func Ch2Evaluate(res *Ch2Result) []Check {
	return []Check{
		checkSession(res),
		checkCh1Parity(res),
		checkLogDump(res),
		checkReplay(res),
		checkRedaction(res),
		checkEphemera(res),
		checkHint(res),
		checkInterrupt(res),
		checkCh2Usage(res),
	}
}

// --- 0. the session itself --------------------------------------------------
//
// Worth zero points, and it can still sink a submission. The Chapter 2 session
// adds directives and acknowledgements to Chapter 1's protocol, and if that
// stream desynchronizes then every check downstream is measuring noise. This
// check exists so the student sees the cause instead of five confusing
// symptoms — and so does whoever is maintaining the grader.

func checkSession(res *Ch2Result) Check {
	c := Check{ID: "session", Title: "The graded session ran to completion", Points: 0, Passed: true}
	for _, p := range res.Protocol {
		c.failf("%s", p)
	}
	for _, l := range res.Extra {
		c.failf("unexpected stdout line: %q — stdout carries the protocol only", truncate(l, 160))
	}
	if res.ExitError != "" {
		c.failf("program did not exit cleanly: %s (exit code %d)", res.ExitError, res.ExitCode)
	}
	var census []string
	for _, r := range res.Records {
		tag := r.Turn
		if tag == "" {
			tag = "(unscripted)"
		}
		if r.Continuation {
			tag += fmt.Sprintf("/cont%d", r.Iteration)
		}
		census = append(census, fmt.Sprintf("%d:%s", r.Seq, tag))
	}
	c.notef("%d requests — %s", len(res.Records), strings.Join(census, " "))
	return c
}

// --- 1. the rewrite kept what it had ---------------------------------------

func checkCh1Parity(res *Ch2Result) Check {
	c := Check{ID: "ch1parity", Title: "Every Chapter 1 check still passes", Points: 25, Passed: true, Earned: 25}
	var failed []string
	for _, p := range res.Parity {
		if !p.Passed {
			failed = append(failed, p.ID)
			c.failf("chapter 1 check %q now fails: %s", p.ID, strings.Join(p.Details, "; "))
		}
	}
	if c.Passed {
		c.notef("all %d Chapter 1 checks pass against the Chapter 2 binary — the rewrite bought properties, not features", len(res.Parity))
	} else {
		c.notef("failing Chapter 1 checks: %s. Run the Chapter 1 grader against this binary for the full diagnosis.",
			strings.Join(failed, ", "))
	}
	return c
}

// --- 2. the log is a log ----------------------------------------------------

func checkLogDump(res *Ch2Result) Check {
	c := Check{ID: "logdump", Title: "The event log serializes and round-trips", Points: 5, Passed: true, Earned: 5}
	if res.LogErr != "" {
		c.failf("the dumped log could not be read as JSON-lines: %s", res.LogErr)
	}
	if !res.Acked["dump-early"] || !res.Acked["dump-final"] {
		c.failf("a {\"dump\": \"<path>\"} directive was not acknowledged with {\"ok\": true}")
	}
	if len(res.FinalLog) == 0 {
		c.failf("the final dump produced no events")
		return c
	}
	prev := -1
	for i, e := range res.FinalLog {
		if !e.SeqOK {
			c.failf("log line %d has no numeric %q field", i+1, "seq")
			break
		}
		if e.Seq <= prev {
			c.failf("log line %d has seq %d, which does not exceed the previous %d — Seq is monotonic and never reused",
				i+1, e.Seq, prev)
			break
		}
		if e.Norm == "" {
			c.failf("log line %d has no %q field", i+1, "type")
			break
		}
		prev = e.Seq
	}
	if res.RenderErr != "" {
		c.failf("`render` could not replay the dumped log in a fresh process: %s", res.RenderErr)
	}
	if c.Passed {
		c.notef("%d events, seq %d..%d, %s",
			len(res.FinalLog), res.FinalLog[0].Seq, res.FinalLog[len(res.FinalLog)-1].Seq,
			typeCensus(res.FinalLog))
	}
	return c
}

func typeCensus(log []LogEvent) string {
	counts := map[string]int{}
	var order []string
	for _, e := range log {
		if counts[e.Norm] == 0 {
			order = append(order, e.Norm)
		}
		counts[e.Norm]++
	}
	var parts []string
	for _, k := range order {
		parts = append(parts, fmt.Sprintf("%s×%d", k, counts[k]))
	}
	return strings.Join(parts, " ")
}

// --- 3. replay is deterministic ---------------------------------------------

func checkReplay(res *Ch2Result) Check {
	c := Check{ID: "replay", Title: "Two renders of one log are byte-identical", Points: 15, Passed: true, Earned: 15}
	if res.RenderErr != "" {
		c.failf("%s", res.RenderErr)
		return c
	}
	if res.RenderCalledNetwork {
		c.failf("`render` called the API. It must be a pure function: log → context → request bytes, printed to stdout. " +
			"If rendering needs the network, the context is not separable from the transport.")
	}
	if !bytes.Equal(res.RenderOut1, res.RenderOut2) {
		c.failf("the two renders differ: %s", firstDiff(res.RenderOut1, res.RenderOut2))
		c.notef("a log is a fixed input. Two runs of a pure function over a fixed input must agree. " +
			"The usual culprits are a timestamp, a random id, or Go's randomized map iteration order " +
			"leaking into the rendered JSON — the same bug class that destroys prefix caching later.")
		return c
	}
	if _, err := parseRendered(res.RenderOut1); err != nil {
		c.failf("`render` did not print a parseable Messages API request: %v", err)
		return c
	}
	c.notef("two independent `render` invocations produced %d identical bytes, with no network access",
		len(res.RenderOut1))
	return c
}

func firstDiff(a, b []byte) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			lo := i - 40
			if lo < 0 {
				lo = 0
			}
			hiA, hiB := i+40, i+40
			if hiA > len(a) {
				hiA = len(a)
			}
			if hiB > len(b) {
				hiB = len(b)
			}
			return fmt.Sprintf("first difference at byte %d\n  run 1: …%s…\n  run 2: …%s…",
				i, string(a[lo:hiA]), string(b[lo:hiB]))
		}
	}
	return fmt.Sprintf("identical for %d bytes, then lengths differ (%d vs %d)", n, len(a), len(b))
}

// parseRendered decodes render's stdout as a Messages API request.
func parseRendered(out []byte) (fakeanthropic.ToolRequest, error) {
	var req fakeanthropic.ToolRequest
	dec := json.NewDecoder(bytes.NewReader(out))
	if err := dec.Decode(&req); err != nil {
		return req, err
	}
	if len(req.Messages) == 0 {
		return req, fmt.Errorf("the rendered request has no messages")
	}
	return req, nil
}

// --- 4. redaction: gone from one artifact, present in the other -------------

func checkRedaction(res *Ch2Result) Check {
	c := Check{ID: "redaction", Title: "Redacted content leaves the context and stays in the log", Points: 15, Passed: true, Earned: 15}
	if !res.RedactFound {
		c.failf("could not find %q in the early dump, so there was nothing to redact. "+
			"The fake planted that string in its first assistant reply; it must appear in the event log.",
			RedactCanary)
		return c
	}
	if !res.Acked["redact"] {
		c.failf("the {\"redact\": %d} directive was not acknowledged", res.RedactSeq)
	}

	// Gone from the context: the live request after the redaction...
	after := requestsAfterTurn(res.Records, "post-redaction")
	if len(after) == 0 {
		c.failf("no request was sent after the redaction, so its effect could not be observed")
	}
	for _, r := range after {
		if r.Count(RedactCanary) > 0 {
			c.failf("request %d, sent after Redacted(seq=%d), still contains %q — "+
				"redaction must remove the payload from what the model sees",
				r.Seq, res.RedactSeq, RedactCanary)
			break
		}
	}
	// ...and from the replayed render, which proves the decision is an event
	// and not a mutation someone forgot to record.
	if len(res.RenderOut1) > 0 && bytes.Contains(res.RenderOut1, []byte(RedactCanary)) {
		c.failf("`render` of the dumped log still emits %q. The Redacted event must be replayed, not just applied "+
			"to the live context at the time it happened.", RedactCanary)
	}

	// Present in the log.
	if ev, ok := FindPayload(res.FinalLog, RedactCanary); !ok {
		c.failf("%q is gone from the dumped log too. Redaction edits the CONTEXT; the log is the audit record "+
			"and must still show what was said. Deleting from both is not redaction, it is amnesia.", RedactCanary)
	} else if ev.Seq != res.RedactSeq {
		c.notef("note: the payload now lives at seq %d, not the seq %d that was redacted", ev.Seq, res.RedactSeq)
	}

	// The decision itself is an auditable event.
	reds := FindEvents(res.FinalLog, "redacted")
	switch {
	case len(reds) == 0:
		c.failf("the log contains no Redacted event. The decision to redact is itself a fact about the conversation, " +
			"and replay cannot reproduce an effect that was never recorded.")
	default:
		found := false
		for _, r := range reds {
			if t, ok := intField(r.Fields, "target_seq", "targetseq", "target", "seq_target"); ok && t == res.RedactSeq {
				found = true
			}
		}
		if !found {
			c.failf("no Redacted event names seq %d as its target (looked for a target_seq field)", res.RedactSeq)
		}
	}
	if c.Passed {
		c.notef("%q: absent from every request after Redacted(seq=%d) and from the replayed render, still present in the log. "+
			"That pair of assertions is History != Context.", RedactCanary, res.RedactSeq)
	}
	return c
}

// --- 5. ephemera are delivered once and never remembered --------------------

func checkEphemera(res *Ch2Result) Check {
	c := Check{ID: "ephemera", Title: "Ephemeral data appears once and is never dialogue", Points: 10, Passed: true, Earned: 10}
	if !res.Acked["set-ephemera"] {
		c.failf("the {\"ephemera\": ...} directive was not acknowledged")
	}
	var carrying []int
	total := 0
	for _, r := range res.Records {
		if n := r.Count(EphemeraCanary); n > 0 {
			carrying = append(carrying, r.Seq)
			total += n
		}
	}
	switch {
	case len(carrying) == 0:
		c.failf("%q never reached the model. Ephemera are attached to the next request; if they are never rendered, "+
			"the EphemeraSet event did nothing.", EphemeraCanary)
	case len(carrying) > 1:
		c.failf("%q appears in %d requests (%v). Ephemera are delivered once and then consumed — "+
			"that is what makes them safe to hold volatile data. A stale timestamp resent forever is a lie.",
			EphemeraCanary, len(carrying), carrying)
	case total > 1:
		c.failf("%q appears %d times inside request %d; it should be carried exactly once",
			EphemeraCanary, total, carrying[0])
	}

	for _, e := range res.FinalLog {
		if e.Has(EphemeraCanary) && dialogueTypes[e.Norm] {
			c.failf("the %s event at seq %d contains %q. Ephemera are not something anyone said: "+
				"they belong in an EphemeraSet event, never in the dialogue.", e.Type, e.Seq, EphemeraCanary)
			break
		}
	}
	if len(res.RenderOut1) > 0 && bytes.Contains(res.RenderOut1, []byte(EphemeraCanary)) {
		c.failf("the replayed render still carries %q. The ephemera were consumed by a RequestSent during the session; "+
			"replaying the log must consume them again, leaving none pending at the end.", EphemeraCanary)
	}
	if c.Passed {
		c.notef("%q: carried in exactly one request (#%d), absent from the dialogue and from the replay",
			EphemeraCanary, carrying[0])
	}
	return c
}

// --- 6. the hint: five separable properties ---------------------------------

func checkHint(res *Ch2Result) Check {
	c := Check{ID: "hint", Title: "A mid-turn message is received, classified, positioned, delivered once, and kept", Points: 15, Passed: true, Earned: 15}

	if !res.HintGateFired {
		c.failf("the fake never reached the point where it holds its reply open — " +
			"the agent_status loop did not run, so there was no turn to steer")
		return c
	}

	// (a) received while blocked.
	if !res.HintAcked {
		c.failf("PROPERTY 'received while blocked': the hint was written to stdin while your program was waiting on an "+
			"HTTP response it had already sent, and no {\"ok\": true} came back within %s. A program that reads stdin "+
			"only between rounds cannot even receive this line. An actor has a mailbox (§2.5): read stdin on its own "+
			"goroutine and deliver into a channel.", GateWindow)
	} else {
		c.notef("received while blocked: acknowledged %s into a request that was still unanswered",
			res.HintAckLatency.Round(time.Microsecond))
	}

	if res.Capped["hint-loop"] {
		c.failf("PROPERTY 'delivered': the fake called agent_status until it hit its iteration cap and never saw %q "+
			"come back. The loop only ends when the steer arrives — a hint that is dropped, or saved for the next "+
			"round, leaves the agent running.", HintText)
		return c
	}

	first := firstRequestContaining(res.Records, HintText)
	if first == nil {
		c.failf("PROPERTY 'delivered': %q never appeared in any request", HintText)
		return c
	}

	// (b) classified as a hint, not a prompt.
	if !first.Continuation {
		c.failf("PROPERTY 'classified': %q first arrived in request %d, which carries no tool_result blocks — "+
			"it started a turn of its own. The identical bytes are a prompt when the turn is Idle and a hint when it "+
			"is InFlight; only the reducer knows which, because only the reducer holds the state.", HintText, first.Seq)
	}
	if res.Answers["hint-loop"] == "" {
		c.failf("PROPERTY 'classified': the hinted turn never produced its {\"assistant\": ...} line. " +
			"A hint steers a turn; it does not end one and it does not earn a reply of its own.")
	}

	// (c) positioned after the tool results.
	hintIdx, lastResult := -1, -1
	for i, b := range first.Blocks {
		if strings.Contains(b.Text, HintText) && b.Type != "tool_result" && hintIdx < 0 {
			hintIdx = i
		}
		if b.Type == "tool_result" {
			lastResult = i
		}
	}
	switch {
	case hintIdx < 0 && strings.Contains(first.Body.SystemText(), HintText):
		c.notef("positioned: carried in the request's system field — the first-class mid-turn path. " +
			"The ordering rule is specific to the hack and does not apply here.")
	case hintIdx < 0:
		c.failf("PROPERTY 'positioned': %q is in request %d somewhere the grader cannot locate as a content block",
			HintText, first.Seq)
	case lastResult >= 0 && hintIdx < lastResult:
		c.failf("PROPERTY 'positioned': the hint is block %d and the last tool_result is block %d — the hint comes "+
			"BEFORE the results it is meant to redirect, so the model reads it as a comment on nothing. "+
			"In the hack, the text block goes after the tool_result blocks, always.", hintIdx, lastResult)
	default:
		c.notef("positioned: block %d, after the last tool_result at block %d", hintIdx, lastResult)
	}

	// (d) delivered once.
	for _, r := range res.Records {
		if n := r.Count(HintText); n > 1 {
			c.failf("PROPERTY 'delivered once': request %d carries %q %d times. It is one thing the human said once; "+
				"re-attaching it to every round is the ephemera mistake applied to the wrong kind of event.",
				r.Seq, HintText, n)
			break
		}
	}

	// (e) retained.
	last := lastRequestOfTurn(res.Records, "after-interrupt")
	if last == nil {
		c.notef("note: could not locate a late request in which to check retention")
	} else if last.Count(HintText) == 0 {
		c.failf("PROPERTY 'retained': %q is gone from request %d, several turns later. Unlike ephemera, a hint really "+
			"was said: permanent in the log, permanent in the dialogue. Delivered once; remembered always.",
			HintText, last.Seq)
	} else {
		c.notef("retained: still present %d requests later", last.Seq-first.Seq)
	}

	// And it must be in the log as a message, not as some special thing.
	found := false
	for _, e := range res.FinalLog {
		if e.Has(HintText) && e.Norm == "messagereceived" {
			found = true
			break
		}
	}
	if !found {
		c.failf("PROPERTY 'retained': no MessageReceived event in the log contains %q. A hint is not a new kind of "+
			"event — it is the same event as a prompt, distinguished only by the state it arrived in.", HintText)
	}
	return c
}

// --- 7. interrupt: recorded, not executed -----------------------------------

func checkInterrupt(res *Ch2Result) Check {
	c := Check{ID: "interrupt", Title: "An interrupted turn records tool calls without executing them", Points: 10, Passed: true, Earned: 10}
	if !res.IntGateFired {
		c.failf("the fake never reached the interrupt gate — the second agent_status loop did not run")
		return c
	}
	if !res.IntAcked {
		c.failf("the {\"interrupt\": true} directive was written while your program was blocked on an HTTP request "+
			"and was not acknowledged within %s", GateWindow)
	}

	// Executed-after-interrupt shows up as the loop continuing.
	var maxIter int
	for _, r := range res.Records {
		if r.Turn == "interrupt-loop" && r.Iteration > maxIter {
			maxIter = r.Iteration
		}
	}
	if maxIter > 2 {
		c.failf("after the interrupt, your program answered the tool call anyway: the fake saw %d continuations of the "+
			"interrupt-loop turn. A tool call that arrives on an Interrupted turn is RECORDED and NOT EXECUTED — "+
			"truth without action. That is only expressible if Interrupted is a state, not a boolean someone remembers "+
			"to check.", maxIter)
	}

	// The log must show the call, though.
	ints := FindEvents(res.FinalLog, "interrupted")
	if len(ints) == 0 {
		c.failf("the log contains no Interrupted event")
	} else {
		lastInt := ints[len(ints)-1]
		var calledAfter, returnedAfter int
		for _, e := range res.FinalLog {
			if e.Seq <= lastInt.Seq {
				continue
			}
			switch e.Norm {
			case "toolcalled":
				calledAfter++
			case "toolreturned":
				returnedAfter++
			}
		}
		if calledAfter == 0 {
			c.failf("no ToolCalled event follows the Interrupted event at seq %d. The model's already-issued call still "+
				"arrived; the log is the record of what happened, so it belongs there even though nothing ran.", lastInt.Seq)
		}
		if returnedAfter > 0 {
			c.failf("%d ToolReturned event(s) follow the Interrupted event at seq %d — something executed",
				returnedAfter, lastInt.Seq)
		}
	}

	// The next request must be legal despite the dangling call, and the next
	// message must be a prompt.
	next := lastRequestOfTurn(res.Records, "after-interrupt")
	if next == nil {
		c.failf("no request was sent for the prompt that followed the interrupt — the next message after a dead turn " +
			"is a prompt, not a hint, and it starts a new turn")
	} else {
		for _, v := range next.Violations {
			c.failf("request %d (the first after the interrupt) is not legal Anthropic JSON: %s", next.Seq, v)
		}
		if want := Ch2ExpectedAnswer("after-interrupt"); strings.TrimSpace(res.Answers["after-interrupt"]) != want {
			c.failf("the prompt after the interrupt answered %q, expected %q",
				truncate(res.Answers["after-interrupt"], 80), truncate(want, 80))
		}
	}
	if c.Passed {
		c.notef("interrupt acknowledged mid-request; the post-interrupt tool call was recorded and never executed; " +
			"the next request rendered legally despite a dangling tool_use")
		if res.AnswerAfterInt {
			c.notef("note: the killed turn still emitted an {\"assistant\": ...} line. That is allowed — " +
				"a partial answer is a defensible design — but it is not required.")
		}
	}
	return c
}

// --- 8. usage ---------------------------------------------------------------

func checkCh2Usage(res *Ch2Result) Check {
	c := Check{ID: "usage", Title: "Cumulative usage, derived from ResponseEnded events", Points: 5, Passed: true, Earned: 5}
	if !res.UsageSeen {
		c.failf("no {\"usage\": {\"input\": .., \"output\": ..}} line after stdin closed")
		return c
	}
	want := res.FakeUsage
	if res.UsageIn != want.InputTokens || res.UsageOut != want.OutputTokens {
		c.failf("reported input=%d output=%d, but the server's totals for this session are input=%d output=%d. "+
			"Every response ends in a ResponseEnded event carrying usage; summing the log is the whole job.",
			res.UsageIn, res.UsageOut, want.InputTokens, want.OutputTokens)
		return c
	}
	ended := len(FindEvents(res.FinalLog, "responseended"))
	if ended == 0 {
		c.failf("the log contains no ResponseEnded events, so the totals cannot have come from the log")
		return c
	}
	c.notef("input=%d output=%d across %d ResponseEnded events", res.UsageIn, res.UsageOut, ended)
	return c
}

// --- helpers ----------------------------------------------------------------

func firstRequestContaining(recs []fakeanthropic.ToolRecord, needle string) *fakeanthropic.ToolRecord {
	for i := range recs {
		if recs[i].Count(needle) > 0 {
			return &recs[i]
		}
	}
	return nil
}

func requestsAfterTurn(recs []fakeanthropic.ToolRecord, turn string) []fakeanthropic.ToolRecord {
	var out []fakeanthropic.ToolRecord
	for i := range recs {
		if recs[i].Turn == turn {
			out = append(out, recs[i:]...)
			return out
		}
	}
	return nil
}

func lastRequestOfTurn(recs []fakeanthropic.ToolRecord, turn string) *fakeanthropic.ToolRecord {
	for i := len(recs) - 1; i >= 0; i-- {
		if recs[i].Turn == turn {
			return &recs[i]
		}
	}
	return nil
}
