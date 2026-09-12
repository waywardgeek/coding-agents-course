package grade

// The Chapter 2 harness.
//
// Three phases, in order:
//
//	1. Chapter 1 parity — the Chapter 1 fake, script and checks, unchanged,
//	   pointed at the Chapter 2 binary. The rewrite has to keep what it had.
//	2. The Chapter 2 session — a tool-driving fake that HOLDS ITS REPLY OPEN
//	   at two chosen moments. That is the whole trick: the grader types at a
//	   program which is provably blocked on an HTTP request it already sent.
//	   A round-synchronous program cannot acknowledge anything in that window,
//	   so "received while blocked" is not a matter of opinion.
//	3. Replay — `render LOG` twice, with no server involvement, byte-compared.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakeanthropic"
)

const (
	Ch2RoundTimeout = 40 * time.Second
	Ch2AckTimeout   = 10 * time.Second
	// GateWindow is how long the fake holds its reply open waiting for the
	// submission to acknowledge a directive typed mid-request.
	GateWindow = 10 * time.Second
	// DrainWindow is how long to wait after an interrupt for any output the
	// submission chooses to emit for the dead turn.
	DrainWindow = 3 * time.Second
)

// Ch2Result is the evidence from one Chapter 2 grading run. Nothing here is a
// judgement; ch02_checks.go turns it into verdicts.
type Ch2Result struct {
	Parity []Check // the seven Chapter 1 checks, re-run against this binary

	Answers   map[string]string // step label -> assistant text
	Acked     map[string]bool   // step label -> directive acknowledged
	Protocol  []string
	Extra     []string
	Stderr    string
	ExitCode  int
	ExitError string
	UsageSeen bool
	UsageIn   int
	UsageOut  int

	Records   []fakeanthropic.ToolRecord
	FakeUsage fakeanthropic.Usage
	Capped    map[string]bool
	Continued map[string]bool

	HintAcked      bool
	HintAckLatency time.Duration
	HintGateFired  bool
	IntAcked       bool
	IntGateFired   bool
	AnswerAfterInt bool

	EarlyLog    []LogEvent
	FinalLog    []LogEvent
	LogErr      string
	RedactSeq   int
	RedactFound bool

	RenderOut1          []byte
	RenderOut2          []byte
	RenderErr           string
	RenderCalledNetwork bool
}

// ch2Line is every shape the program may write to stdout.
type ch2Line struct {
	Assistant *string `json:"assistant,omitempty"`
	OK        *bool   `json:"ok,omitempty"`
	Usage     *struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"usage,omitempty"`
}

type reader struct {
	answers chan string
	acks    chan bool
	usage   chan ch2Line
	junk    chan string
	done    chan struct{}
}

// Ch2Run executes all three phases and returns the evidence.
func Ch2Run(bin string) (*Ch2Result, error) {
	res := &Ch2Result{
		Answers:   map[string]string{},
		Acked:     map[string]bool{},
		Capped:    map[string]bool{},
		Continued: map[string]bool{},
	}

	// --- phase 1: Chapter 1 parity -------------------------------------
	parity, err := Run(bin)
	if err != nil {
		return nil, fmt.Errorf("chapter 1 parity run: %w", err)
	}
	res.Parity = Evaluate(parity)

	// --- phase 2: the Chapter 2 session --------------------------------
	tmp, err := os.MkdirTemp("", "course-ch02-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	earlyPath := filepath.Join(tmp, "log-early.jsonl")
	finalPath := filepath.Join(tmp, "log-final.jsonl")

	fake := &fakeanthropic.ToolServer{Turns: Ch2Turns}
	baseURL, err := fake.Start()
	if err != nil {
		return nil, fmt.Errorf("starting fake server: %w", err)
	}
	defer fake.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_BASE_URL="+baseURL,
		"ANTHROPIC_API_KEY=sk-ant-course-grader-fake",
		"ANTHROPIC_MODEL=claude-fake-course-2",
		"ANTHROPIC_API_URL="+baseURL,
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting submission: %w", err)
	}

	rd := &reader{
		answers: make(chan string, 64),
		acks:    make(chan bool, 64),
		usage:   make(chan ch2Line, 8),
		junk:    make(chan string, 64),
		done:    make(chan struct{}),
	}
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			var pl ch2Line
			if err := json.Unmarshal([]byte(line), &pl); err != nil {
				rd.junk <- line
				continue
			}
			switch {
			case pl.Assistant != nil:
				rd.answers <- *pl.Assistant
			case pl.OK != nil:
				rd.acks <- *pl.OK
			case pl.Usage != nil:
				rd.usage <- pl
			default:
				rd.junk <- line
			}
		}
		close(rd.done)
	}()

	var writeMu sync.Mutex
	write := func(v any) error {
		b, _ := json.Marshal(v)
		writeMu.Lock()
		defer writeMu.Unlock()
		_, err := stdin.Write(append(b, '\n'))
		return err
	}

	// The gate. Called on the fake's HTTP goroutine, with the submission's
	// request still unanswered.
	fake.Hook = func(info fakeanthropic.HookInfo) {
		var directive any
		switch info.Turn {
		case "hint-loop":
			res.HintGateFired = true
			directive = map[string]any{"hint": HintText}
		case "interrupt-loop":
			res.IntGateFired = true
			directive = map[string]any{"interrupt": true}
		default:
			return
		}
		start := time.Now()
		if err := write(directive); err != nil {
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("could not write the %s directive: %v", info.Turn, err))
			return
		}
		select {
		case <-rd.acks:
			if info.Turn == "hint-loop" {
				res.HintAcked = true
				res.HintAckLatency = time.Since(start)
				res.Acked["hint"] = true
			} else {
				res.IntAcked = true
				res.Acked["interrupt"] = true
			}
		case <-time.After(GateWindow):
			// No acknowledgement while blocked. Release anyway: the point is
			// to diagnose, not to hang.
		case <-rd.done:
		}
	}

	drainJunk := func() {
		for {
			select {
			case l := <-rd.junk:
				res.Extra = append(res.Extra, l)
			default:
				return
			}
		}
	}

	awaitAnswer := func(label string) {
		select {
		case a := <-rd.answers:
			res.Answers[label] = a
		case <-time.After(Ch2RoundTimeout):
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("step %q: no {\"assistant\": ...} line within %s", label, Ch2RoundTimeout))
		case <-rd.done:
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("step %q: the program's stdout closed before it answered", label))
		}
		drainJunk()
	}

	awaitAck := func(label string) {
		select {
		case <-rd.acks:
			res.Acked[label] = true
		case <-time.After(Ch2AckTimeout):
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("step %q: no {\"ok\": true} acknowledgement within %s", label, Ch2AckTimeout))
		case <-rd.done:
			res.Protocol = append(res.Protocol,
				fmt.Sprintf("step %q: the program's stdout closed before it acknowledged", label))
		}
		drainJunk()
	}

	for _, step := range Ch2Steps {
		switch step.Kind {
		case StepUser:
			if err := write(map[string]any{"user": step.User}); err != nil {
				res.Protocol = append(res.Protocol,
					fmt.Sprintf("step %q: could not write to stdin (%v) — did the program exit?", step.Label, err))
				goto finished
			}
			if step.Interrupted {
				// A killed turn owes no answer. Accept one if it comes (some
				// designs emit the partial text) but do not require it.
				select {
				case a := <-rd.answers:
					res.AnswerAfterInt = true
					res.Answers[step.Label] = a
				case <-time.After(DrainWindow):
				case <-rd.done:
				}
				drainJunk()
				continue
			}
			awaitAnswer(step.Label)

		case StepDirective:
			d := map[string]any{}
			for k, v := range step.Directive {
				d[k] = v
			}
			switch step.Label {
			case "dump-early":
				d["dump"] = earlyPath
			case "dump-final":
				d["dump"] = finalPath
			case "redact":
				if !res.RedactFound {
					// Nothing to redact; still send the directive so the
					// stream stays synchronous, but aim it at a seq that
					// cannot exist so the failure is legible.
					d["redact"] = -1
				} else {
					d["redact"] = res.RedactSeq
				}
			}
			if err := write(d); err != nil {
				res.Protocol = append(res.Protocol,
					fmt.Sprintf("step %q: could not write to stdin (%v)", step.Label, err))
				goto finished
			}
			awaitAck(step.Label)
			if step.Label == "dump-early" && res.Acked["dump-early"] {
				if log, err := ParseLog(earlyPath); err != nil {
					res.LogErr = fmt.Sprintf("early dump: %v", err)
				} else {
					res.EarlyLog = log
					if ev, ok := FindPayload(log, RedactCanary); ok && ev.SeqOK {
						res.RedactSeq, res.RedactFound = ev.Seq, true
					}
				}
			}
		}
	}

finished:
	_ = stdin.Close()

	// Everything after EOF: the usage line, then exit.
	deadline := time.After(Ch2AckTimeout)
collect:
	for {
		select {
		case u := <-rd.usage:
			res.UsageSeen = true
			res.UsageIn = u.Usage.Input
			res.UsageOut = u.Usage.Output
		case l := <-rd.junk:
			res.Extra = append(res.Extra, l)
		case a := <-rd.answers:
			res.Extra = append(res.Extra, fmt.Sprintf("{\"assistant\": %q} after stdin closed", truncate(a, 80)))
		case <-rd.done:
			break collect
		case <-deadline:
			break collect
		}
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				res.ExitCode = ee.ExitCode()
			} else {
				res.ExitCode = -1
			}
			res.ExitError = err.Error()
		}
	case <-time.After(ExitTimeout):
		_ = cmd.Process.Kill()
		res.ExitCode = -1
		res.ExitError = "program did not exit after stdin was closed"
	}

	// Drain anything the reader still has buffered.
	for {
		select {
		case u := <-rd.usage:
			res.UsageSeen = true
			res.UsageIn, res.UsageOut = u.Usage.Input, u.Usage.Output
			continue
		case l := <-rd.junk:
			res.Extra = append(res.Extra, l)
			continue
		default:
		}
		break
	}

	res.Stderr = stderr.String()
	res.Records = fake.Records()
	res.FakeUsage = fake.TotalUsage()
	for _, t := range Ch2Turns {
		res.Capped[t.Name] = fake.Capped(t.Name)
		res.Continued[t.Name] = fake.ContinuedAfterStop(t.Name)
	}

	if log, err := ParseLog(finalPath); err != nil {
		if res.LogErr == "" {
			res.LogErr = fmt.Sprintf("final dump: %v", err)
		}
	} else {
		res.FinalLog = log
	}

	// --- phase 3: replay ------------------------------------------------
	if len(res.FinalLog) > 0 {
		before := len(fake.Records())
		out1, err1 := runRender(bin, finalPath, baseURL)
		out2, err2 := runRender(bin, finalPath, baseURL)
		after := len(fake.Records())
		res.RenderCalledNetwork = after > before
		res.RenderOut1, res.RenderOut2 = out1, out2
		switch {
		case err1 != nil:
			res.RenderErr = err1.Error()
		case err2 != nil:
			res.RenderErr = err2.Error()
		}
		res.Records = fake.Records()
	} else if res.RenderErr == "" {
		res.RenderErr = "no event log was dumped, so render could not be exercised"
	}

	return res, nil
}

// runRender executes `bin render LOG` and returns its stdout. The fake's base
// URL is passed deliberately: if the submission renders by calling the API,
// the request lands on the fake and the grader sees it.
func runRender(bin, logPath, baseURL string) ([]byte, error) {
	cmd := exec.Command(bin, "render", logPath)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_BASE_URL="+baseURL,
		"ANTHROPIC_API_KEY=sk-ant-course-grader-fake",
		"ANTHROPIC_MODEL=claude-fake-course-2",
		"ANTHROPIC_API_URL="+baseURL,
	)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	fin := make(chan error, 1)
	go func() { fin <- cmd.Wait() }()
	select {
	case err = <-fin:
	case <-time.After(Ch2AckTimeout):
		_ = cmd.Process.Kill()
		return nil, errors.New("render did not exit")
	}
	if err != nil {
		return out.Bytes(), fmt.Errorf("`render %s` failed: %v; stderr: %s",
			filepath.Base(logPath), err, truncate(errb.String(), 300))
	}
	if len(bytes.TrimSpace(out.Bytes())) == 0 {
		return out.Bytes(), errors.New("`render` printed nothing to stdout")
	}
	return out.Bytes(), nil
}
