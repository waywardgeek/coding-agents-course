# Building Advanced AI Coding Agents — course code

Exercise graders and reference solutions for the book/course.

Book text lives separately; this repo is the machinery: an auto-grader per
chapter, plus the canonical solution the chapter's code listings are drawn
from.

## Layout

```
book/                   chapter outlines (the text these graders serve)
cmd/grade/              the auto-grader CLI (-ch selects the chapter)
internal/fakeanthropic/ deterministic stand-ins for the Messages API
                        (Server: chapter 1; ToolServer: chapter 2, drives a
                        tool loop and can hold its reply open)
internal/grade/         scripts, process harnesses, checks, report
solutions/ch01/         reference solution (the chapter's own code)
solutions/ch02/         reference solution: event log, reducer, renderer, actor
scripts/live.sh         run a solution against the real API
testdata/students/      deliberately defective submissions (grader self-test)
```

## Try it yourself

**Grade the reference solution** — no API key, no network, no cost:

```bash
make grade                          # or: go run ./cmd/grade ./solutions/ch01
```

**Grade your own submission** — point it at any package directory or built
binary:

```bash
make grade-dir DIR=~/my-agent       # or: go run ./cmd/grade ~/my-agent
go run ./cmd/grade -json ~/my-agent # machine-readable report
```

Exit status is 0 on a pass, 1 on a fail, 2 if the grader itself could not run.

**Talk to it live**, against the real Anthropic API:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
scripts/live.sh models   # which model IDs your key can actually use
scripts/live.sh          # three scripted rounds + the token bill
scripts/live.sh chat     # interactive REPL
```

`scripts/live.sh` never echoes your key. It reads `$ANTHROPIC_API_KEY`, else
the file named by `$ANTHROPIC_API_KEY_FILE`, else `~/.cr/settings.json` if you
happen to run CodeRhapsody. Override the model with `ANTHROPIC_MODEL=...`.

**Check the grader itself** — does it catch real defects?

```bash
make test
```

## Chapter 1 — the exercise contract

Ship a Go program that speaks JSON lines on stdio:

| direction | line |
|---|---|
| grader → program | `{"user": "..."}` |
| program → grader | `{"assistant": "..."}` |
| grader → program | *(closes stdin)* |
| program → grader | `{"usage": {"input": N, "output": M}}`, then exit 0 |

Environment supplied to the submission:

- `ANTHROPIC_BASE_URL` — where to POST (`$BASE/v1/messages`)
- `ANTHROPIC_API_KEY` — send it as the `x-api-key` header
- `ANTHROPIC_MODEL` — the model string to put in the request

**stdout carries the protocol only.** Diagnostics go to stderr.

### What is checked, and why

| check | pts | property |
|---|---|---|
| `protocol` | 15 | the stdio contract: one answer per round, clean exit, no junk on stdout |
| `wire` | 15 | well-formed Messages API calls: headers, `max_tokens`, alternating non-empty roles |
| `calls` | 10 | exactly one API call per round |
| `replies` | 10 | answers are the text the server returned — i.e. the response was parsed |
| `memory` | 25 | **the conversation exists** |
| `growth` | 15 | each request extends the previous one byte-for-byte (append-only) |
| `usage` | 10 | cumulative token totals reported and correct |

Every check must pass. There is no partial credit for a conversation that
does not exist.

**How the memory check works.** The fake plants a fixed string in its
*round-1 assistant reply*. At round 4 it looks for that string in an assistant
message of the incoming request. The student's program never types that string
— the server did — so it can only be present if the program appended the
model's reply to the history and resent the whole thing. A program that makes
a fresh single-message call per question cannot fake it. This is the proof
that the API is stateless and the conversation lives in your process.

The probe locates its request by *content* (round 4's user text as the final
message) rather than by arrival order, so it stays correct even when a
submission makes stray extra calls.

## Chapter 2 — the exercise contract

Rebuild the Chapter 1 chatbot on an append-only **event log** and a derived
**context**. For everything Chapter 1 could do it is observably identical.
Then it does two things Chapter 1 could not have expressed at any price: it
accepts a **hint** mid-turn, and it survives an **interrupt**.

Three surfaces:

| invocation | behavior |
|---|---|
| `./ch02` | grader mode — the Chapter 1 stdio contract, plus directives |
| `./ch02 chat` | the REPL (type while it is working to steer it) |
| `./ch02 render LOG` | play LOG → context → render; print the request; **no network** |

Directives extend the Chapter 1 protocol. `{"user": ...}` still expects exactly
one `{"assistant": ...}`; every other object expects exactly one `{"ok": true}`:
`{"hint": "..."}`, `{"interrupt": true}`, `{"redact": <seq>}`,
`{"ephemera": {"instruction": "...", "text": "..."}}`, `{"dump": "<path>"}`.

The log is JSON-lines, one event per line, ascending `seq`.

| check | pts | property |
|---|---|---|
| `session` | 0 | diagnostic: the graded session ran; request census |
| `ch1parity` | 25 | all seven Chapter 1 checks still pass, unchanged |
| `logdump` | 5 | log serializes and round-trips; `seq` monotonic, never reused |
| `replay` | 15 | two `render` invocations on one log are byte-identical |
| `redaction` | 15 | payload gone from the request, still in the log |
| `ephemera` | 10 | carried in exactly one request, never a dialogue event |
| `hint` | 15 | received while blocked, classified, positioned, once, retained |
| `interrupt` | 10 | late tool call recorded and **not** executed; next request legal |
| `usage` | 5 | cumulative totals from `ResponseEnded` events |

The grader holds its reply open at two chosen moments and types at your program
while it is blocked on an HTTP request it has already sent. A round-synchronous
program cannot acknowledge anything in that window, which is how "received
while blocked" is measured rather than assumed.

## Grader design rules

1. **Record, then judge.** The fake captures every request verbatim; checks run
   afterwards against evidence, never against live state.
2. **A malformed request still gets a 200.** Violations are recorded, not
   enforced at the transport layer. Returning 400 on the first mistake teaches
   one bug per run; this teaches all of them in one run.
3. **Deterministic.** Replies are scripted by request index and token counts are
   a pure function of the payload (one token per four characters). No model, no
   network, no cost, identical output everywhere. That is also what lets the
   grader catch invented usage numbers.
4. **One defect, one diagnosis.** Junk on stdout does not desynchronize the
   stream, and request *count* problems are reported by `calls`, not `wire`.

## The grader's own test suite

`internal/grade/grader_test.go` is a sensitivity proof, not a smoke test. A
single mutant submission (`testdata/students/mutant`) takes a defect from
`COURSE_MUTATION` — `amnesiac`, `useronly`, `fabricate`, `twocalls`,
`noversion`, `nomaxtokens`, `nousage`, `fakeusage`, `chatty` — and the test
asserts the **exact set** of checks that fails for each, plus a no-defect
control that must pass.

Exact-set assertions matter: a grader that fails every check on any defect is
as useless as one that fails none, because it cannot tell the student what is
wrong.

## Running a solution live

The same binary talks to the real API (or the course proxy) — one env var, and
the program never knows the difference:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
go run ./solutions/ch01 chat        # interactive REPL
```

Leave `ANTHROPIC_BASE_URL` unset for `api.anthropic.com`, or point it at the
course proxy.

## License

Apache License 2.0 — see [LICENSE](LICENSE). The solutions are teaching code:
copy them, ship them, build on them.
