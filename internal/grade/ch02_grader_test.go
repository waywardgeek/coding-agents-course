package grade

// Mutation testing for the Chapter 2 grader.
//
// A grader that passes the reference solution has demonstrated nothing. The
// question is whether it FAILS when it should, on the specific check that
// names the defect — a grader that fails everything whenever anything is
// wrong is as useless as one that fails nothing.
//
// So each mutation applies one surgical patch to a copy of the reference
// solution and asserts the EXACT SET of failing check ids. If a patch no
// longer matches the solution's source, the test fails loudly rather than
// silently grading an unmutated program.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type patch struct {
	file string
	old  string
	new  string
}

type mutation struct {
	name    string
	why     string
	patches []patch
	expect  []string // check ids expected to FAIL
}

var mutations = []mutation{
	{
		name:   "control",
		why:    "the unmodified reference solution",
		expect: nil,
	},
	{
		name: "hint-in-seq-order",
		why:  "delivers the hint at its arrival position instead of holding it pending, so it lands before the tool results",
		patches: []patch{{
			file: "context.go",
			old:  "c.PendingHints = append(c.PendingHints, entry)",
			new:  "c.Dialogue = append(c.Dialogue, entry)",
		}},
		expect: []string{"hint"},
	},
	{
		name: "hint-not-retained",
		why:  "carries the hint once and then forgets it, treating a thing a human said as if it were ephemera",
		patches: []patch{{
			file: "context.go",
			old:  "c.Dialogue = append(c.Dialogue, c.PendingHints...)\n\t\tc.PendingHints = nil",
			new:  "c.PendingHints = nil",
		}},
		expect: []string{"hint"},
	},
	{
		name: "hint-resent-every-round",
		why:  "never consumes the pending hint, so it is re-delivered on every subsequent request",
		patches: []patch{{
			file: "context.go",
			old:  "c.PendingHints = nil\n\t\tc.Ephemera = nil",
			new:  "c.Ephemera = nil",
		}},
		expect: []string{"hint"},
	},
	{
		name: "hint-dropped-mid-turn",
		why:  "receives the mid-turn message and throws it away, which is what a request builder does",
		patches: []patch{{
			file: "engine.go",
			old:  "\t\te.record(Event{Type: EvMessageReceived, Actor: &You, Parts: TextParts(m.Text)})",
			new:  "\t\tif before != InFlight && before != ToolsPending {\n\t\t\te.record(Event{Type: EvMessageReceived, Actor: &You, Parts: TextParts(m.Text)})\n\t\t}",
		}},
		expect: []string{"hint"},
	},
	{
		name: "ephemera-never-consumed",
		why:  "leaves ephemera attached after the request that carried them, so volatile data becomes permanent",
		patches: []patch{{
			file: "context.go",
			old:  "c.PendingHints = nil\n\t\tc.Ephemera = nil",
			new:  "c.PendingHints = nil",
		}},
		expect: []string{"ephemera"},
	},
	{
		name: "redaction-deletes-the-log-too",
		why:  "removes the event from the audit record as well as from the context — amnesia, not redaction",
		patches: []patch{{
			file: "engine.go",
			old:  "\t\te.record(Event{Type: EvRedacted, Actor: &You, TargetSeq: m.Seq})",
			new: "\t\tfor i, ev := range e.log.Events {\n" +
				"\t\t\tif ev.Seq == m.Seq {\n\t\t\t\te.log.Events = append(e.log.Events[:i], e.log.Events[i+1:]...)\n\t\t\t\tbreak\n\t\t\t}\n\t\t}\n" +
				"\t\te.record(Event{Type: EvRedacted, Actor: &You, TargetSeq: m.Seq})",
		}},
		expect: []string{"redaction"},
	},
	{
		name: "redaction-recorded-but-not-applied",
		why:  "writes the Redacted event and never acts on it, so the payload stays in the context",
		patches: []patch{{
			file: "context.go",
			old:  "c.Dialogue[i].Redacted = true",
			new:  "_ = i",
		}},
		expect: []string{"redaction"},
	},
	{
		name: "nondeterministic-render",
		why:  "leaks the clock into the rendered request, the bug class that later destroys prefix caching",
		patches: []patch{
			{file: "render.go", old: "\t\"crypto/sha256\"", new: "\t\"time\"\n\t\"crypto/sha256\""},
			{
				file: "render.go",
				old:  "req := wireRequest{Model: model, MaxTokens: maxTokens, Tools: toolSchema()}",
				new:  "req := wireRequest{Model: model, MaxTokens: maxTokens, Tools: toolSchema()}\n\treq.System = \"rendered at \" + time.Now().Format(time.RFC3339Nano)",
			},
		},
		expect: []string{"replay"},
	},
	{
		name: "interrupt-is-a-noop",
		why:  "records the interrupt but does not make it a state, so the late tool call is executed",
		patches: []patch{{
			file: "context.go",
			old:  "\t\tc.Turn = Interrupted\n\t\tc.PendingTools = nil",
			new:  "\t\tbreak",
		}},
		expect: []string{"interrupt"},
	},
	{
		name: "junk-on-stdout",
		why:  "prints a diagnostic to stdout, where the protocol lives",
		patches: []patch{{
			file: "engine.go",
			old:  "\tgo e.readStdin()",
			new:  "\tgo e.readStdin()\n\tfmt.Println(\"engine: starting up\")",
		}},
		expect: []string{"session", "ch1parity"},
	},
	{
		name: "reads-stdin-only-between-turns",
		why:  "keeps the channel but refuses to look at it while a request is outstanding — the Chapter 1 shape, which cannot receive a hint at any price",
		patches: []patch{{
			file: "engine.go",
			old:  "\t\tselect {\n\t\tcase m, ok := <-inbox:",
			new: "\t\tinboxNow := inbox\n" +
				"\t\tif e.ctx.Turn == InFlight || e.ctx.Turn == ToolsPending {\n\t\t\tinboxNow = nil\n\t\t}\n" +
				"\t\tselect {\n\t\tcase m, ok := <-inboxNow:",
		}},
		expect: []string{"hint", "interrupt", "session"},
		// Three checks, one root cause, and that is correct rather than a
		// cascade bug. A program with no mailbox cannot receive the interrupt
		// directive while blocked either, so the late tool call runs and the
		// dead turn answers late. The prediction when this mutation was
		// written was {hint} alone; the grader was right and the prediction
		// was wrong. What matters is that the FIRST failure a student reads
		// names the mailbox.
	},
	{
		name: "no-render-subcommand",
		why:  "cannot expose log -> context -> request as a pure function, which means transport and context are entangled",
		patches: []patch{{
			file: "main.go",
			old:  "case args[0] == \"render\":",
			new:  "case args[0] == \"render\" && false:",
		}},
		expect: []string{"logdump", "replay"},
	},
}

func TestCh2GraderDetectsMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("each mutation builds and grades a full submission")
	}
	for _, m := range mutations {
		m := m
		t.Run(m.name, func(t *testing.T) {
			dir := mutantDir(t, m)
			bin, cleanup, err := Build(dir)
			if err != nil {
				t.Fatalf("building mutant %q: %v", m.name, err)
			}
			defer cleanup()

			res, err := Ch2Run(bin)
			if err != nil {
				t.Fatalf("grading mutant %q: %v", m.name, err)
			}
			var failed []string
			for _, c := range Ch2Evaluate(res) {
				if !c.Passed {
					failed = append(failed, c.ID)
				}
			}
			sort.Strings(failed)
			want := append([]string(nil), m.expect...)
			sort.Strings(want)

			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Errorf("mutation %q (%s)\n  failing checks: [%s]\n  expected:       [%s]",
					m.name, m.why, strings.Join(failed, ", "), strings.Join(want, ", "))
				for _, c := range Ch2Evaluate(res) {
					if !c.Passed {
						t.Logf("  %s: %s", c.ID, strings.Join(c.Details, " | "))
					}
				}
			}
		})
	}
}

// mutantDir copies the reference solution into a scratch directory inside the
// module and applies the mutation's patches.
func mutantDir(t *testing.T, m mutation) string {
	t.Helper()
	src := filepath.Join("..", "..", "solutions", "ch02")
	dst := filepath.Join("..", "..", "testdata", "mutants", m.name)
	if err := os.RemoveAll(dst); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dst) })

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	contents := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		contents[e.Name()] = string(b)
	}
	for _, p := range m.patches {
		body, ok := contents[p.file]
		if !ok {
			t.Fatalf("mutation %q patches %s, which is not part of the solution", m.name, p.file)
		}
		if n := strings.Count(body, p.old); n != 1 {
			t.Fatalf("mutation %q: pattern appears %d times in %s (want exactly 1):\n%s",
				m.name, n, p.file, p.old)
		}
		contents[p.file] = strings.Replace(body, p.old, p.new, 1)
	}
	for name, body := range contents {
		if err := os.WriteFile(filepath.Join(dst, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}
