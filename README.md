# Building Advanced AI Coding Agents — course code

Exercise graders and reference solutions for the book/course.

Book text lives separately; this repo is the machinery: an auto-grader per
chapter, plus the canonical solution the chapter's code listings are drawn
from.

## Layout

```
cmd/grade/              the Chapter 1 auto-grader CLI
internal/fakeanthropic/ deterministic stand-in for the Messages API
internal/grade/         script, process harness, checks, report
solutions/ch01/         reference solution (the chapter's own code)
testdata/students/      deliberately defective submissions (grader self-test)
```

## Quick start

```bash
go test ./...                      # grader self-test: does it catch real defects?
go run ./cmd/grade ./solutions/ch01 # grade the reference solution
go run ./cmd/grade -json ./mysubmission
```

Grading needs **no API key and no network**. The submission is pointed at a
fake Anthropic server on localhost through `ANTHROPIC_BASE_URL`.

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
