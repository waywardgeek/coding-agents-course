package grade

// The Chapter 2 scripted session.
//
// Two rules govern every choice here, both inherited from Chapter 1:
//
//   - Nothing tests the model's intelligence, because there is no model.
//     Every check is structural.
//   - Every payload the grader looks for is planted by the SERVER, never
//     typed by the student. A canary the submission could have produced
//     itself proves nothing.

import "github.com/waywardgeek/coding-agents-course/internal/fakeanthropic"

const (
	// RedactCanary is planted in the fake's round-1 assistant reply. It is
	// the payload the grader later redacts: it must vanish from the rendered
	// request and survive in the dumped log. Those two assertions together
	// are History != Context, expressed in grader code.
	RedactCanary = "GRACKLE-8813"

	// EphemeraCanary is injected by the grader as ephemeral data. It must
	// appear in exactly one request, exactly once, and never as a dialogue
	// event in the log.
	EphemeraCanary = "KESTREL-5150"

	// HintText is the mid-turn steer. The fake keeps calling agent_status
	// until it sees these words come back, so a dropped or deferred hint
	// shows up as a loop that will not end.
	HintText = "please stop"
)

// The prompts. Each one is shared by the stdin script and by the fake, which
// uses it to recognize the start of a new turn — so they cannot drift apart.
const (
	promptGreeting  = "Hello! Give my project a codename and write it into your reply."
	promptEphemera  = "What build am I on?"
	promptHintLoop  = "Check your status over and over until I tell you to stop."
	promptRetention = "Fine. Are you still there?"
	promptPostRedac = "Carry on."
	promptIntLoop   = "Loop on your status again, please."
	promptAfterInt  = "Are you still with me?"
)

// Ch2Turns is what the fake does, keyed by the order in which turns begin.
// The grader's stdin script (Ch2Steps) must keep these in lockstep.
var Ch2Turns = []fakeanthropic.Turn{
	{
		Name:   "greeting",
		Prompt: promptGreeting,
		Kind:   fakeanthropic.TurnText,
		Text: "Hello! I have written your project's codename, " + RedactCanary +
			", into this reply so that it lives in the conversation rather than in your source.",
	},
	{
		Name:   "ephemera",
		Prompt: promptEphemera,
		Kind:   fakeanthropic.TurnText,
		Text:   "Noted. I can see the volatile data you attached to this request.",
	},
	{
		Name:          "hint-loop",
		Prompt:        promptHintLoop,
		Kind:          fakeanthropic.TurnLoop,
		Text:          "Understood — stopping the status loop as you asked.",
		GateAt:        2,
		MaxIterations: 8,
		StopWhen: func(r fakeanthropic.ToolRecord) bool {
			// Stop as soon as the words appear ANYWHERE. Whether they were
			// carried in the right place is the hint check's job, not the
			// server's: a student who is 90% right deserves a named failure,
			// not an infinite loop and a bare timeout.
			return r.Count(HintText) > 0
		},
	},
	{
		Name:   "retention",
		Prompt: promptRetention,
		Kind:   fakeanthropic.TurnText,
		Text:   "Still here. Nothing has been forgotten.",
	},
	{
		Name:   "post-redaction",
		Prompt: promptPostRedac,
		Kind:   fakeanthropic.TurnText,
		Text:   "Acknowledged.",
	},
	{
		Name:          "interrupt-loop",
		Prompt:        promptIntLoop,
		Kind:          fakeanthropic.TurnLoop,
		Text:          "(this reply should never be reached)",
		GateAt:        2,
		MaxIterations: 4,
		// No StopWhen: the only thing that ends this loop is the submission
		// declining to answer a tool call it recorded after an interrupt.
	},
	{
		Name:   "after-interrupt",
		Prompt: promptAfterInt,
		Kind:   fakeanthropic.TurnText,
		Text:   "Yes, I am still here, and I remember everything up to the interruption.",
	},
}

// StepKind distinguishes the two shapes of line the grader writes.
type StepKind int

const (
	StepUser StepKind = iota
	StepDirective
)

// Step is one line written to the submission's stdin.
type Step struct {
	Kind  StepKind
	Label string

	// User is the prompt text for StepUser.
	User string

	// Directive is the raw JSON object for StepDirective, built at run time
	// for the ones that depend on discovered state (a seq to redact, a temp
	// path to dump to).
	Directive map[string]any

	// Interrupted marks a round whose turn is killed mid-flight. Such a round
	// is not expected to produce an {"assistant"} line.
	Interrupted bool
}

// Ch2Steps is the stdin script. Directives whose contents are discovered at
// run time (dump paths, the seq to redact) are filled in by the harness.
var Ch2Steps = []Step{
	{Kind: StepUser, Label: "greeting", User: promptGreeting},
	{Kind: StepDirective, Label: "dump-early", Directive: map[string]any{"dump": ""}},
	{Kind: StepDirective, Label: "set-ephemera", Directive: map[string]any{"ephemera": map[string]any{
		"instruction": "Volatile context for this request only. Do not treat it as something the user said.",
		"text":        "current_build=" + EphemeraCanary,
	}}},
	{Kind: StepUser, Label: "ephemera", User: promptEphemera},
	{Kind: StepUser, Label: "hint-loop", User: promptHintLoop},
	{Kind: StepUser, Label: "retention", User: promptRetention},
	{Kind: StepDirective, Label: "redact", Directive: map[string]any{"redact": 0}},
	{Kind: StepUser, Label: "post-redaction", User: promptPostRedac},
	{Kind: StepUser, Label: "interrupt-loop", User: promptIntLoop, Interrupted: true},
	{Kind: StepUser, Label: "after-interrupt", User: promptAfterInt},
	{Kind: StepDirective, Label: "dump-final", Directive: map[string]any{"dump": ""}},
}

// Ch2ExpectedAnswer returns the assistant text the fake will have produced for
// a given step label, or "" if that step produces no answer.
func Ch2ExpectedAnswer(label string) string {
	for _, t := range Ch2Turns {
		if t.Name == label {
			if t.Kind == fakeanthropic.TurnLoop && label == "interrupt-loop" {
				return ""
			}
			return t.Text
		}
	}
	return ""
}
