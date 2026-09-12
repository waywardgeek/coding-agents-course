# Review — Chapter 1 outline

**Reviewer:** CodeRhapsody · **Date:** 2026-09-11
**Target:** `book/chapter-01-outline.md` (LGTM'd draft 1)
**Status:** **APPLIED 2026-09-11** to `chapter-01-outline.md` (now draft 2).
Kept in the repo as the rationale record — the outline states the rules, this
states why each one is there.

## Disposition

All of R1–R11 applied; D1–D5 checked on a final pass and held. Three
deviations, all additive:

- **Every ground-truth claim was re-verified against the running code before
  being written into the book**, not taken on the review's word. The seven
  check IDs, the point split, the 49 → 78 → 95 → 116 → 138 curve (cumulative
  476 in / 65 out), the three env vars, and grader-mode-as-default all
  confirmed. `countTokens` is `ceil(len/4)` with a floor of 1 — the review's
  "one token per four characters" is right, and the chapter now says
  "rounded up".
- **R6 survived a challenge.** A first grep found alternation but no
  first/last role enforcement, which looked like the review overstating its
  case. Reading `fake.go:246–254` directly showed both rules are enforced.
  The review was right; the grep was too narrow.
- **R5 extended.** Auditing the fake for *other* graded-but-unstated rules
  turned up three more in the same defect class: `content-type:
  application/json`, rejection of `stream: true`, and non-empty content. All
  are now documented, since a rule that is enforced but unstated is the exact
  failure mode R2 and R6 exist to fix.

Two additions beyond the review: a `make grade-dir` invocation so students
have a local feedback loop (R10 says grading is free but never said how), and
the record-then-judge design note promoted into the chapter, because students
build graders themselves later in the course.

The review's method — rebuild the spec from the artifact that actually runs —
was then applied forward to `chapter-02-outline.md`, whose exercise section
was under-specified in the same way but more severely: it referenced
interrupt directives, event-log serialization and replay without defining
any of them. See that file's open questions 6–9.

---

## Why this review exists

These comments are **empirical, not editorial**. They come from building the
Chapter 1 auto-grader and the reference solution in this repo and then running
both — offline against the fake server and live against the real Anthropic
API. Every gap below is one I hit, or one a student will demonstrably hit.

Nothing here is a style opinion. Where I had a taste preference I left it out.

## Ground truth: the contract as actually implemented

The outline is the spec, but the rig is now the *operational* truth, and they
have drifted. Where they disagree, this is what exists and passes:

| artifact | path |
|---|---|
| grader CLI | `cmd/grade` |
| fake Messages API | `internal/fakeanthropic/fake.go` |
| script, harness, checks | `internal/grade/` |
| reference solution | `solutions/ch01/main.go` |
| grader self-test | `internal/grade/grader_test.go` |
| live runner | `scripts/live.sh` |

**Script:** 5 rounds. The fake plants `TANAGER-4417` in its **round-1
assistant reply**; the **round-4** request is probed for it.

**Environment the grader supplies to the submission:**
`ANTHROPIC_BASE_URL`, `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`.

**The seven checks (100 points, all required):**

| id | points | property |
|---|---|---|
| `protocol` | 15 | one answer per round, clean exit 0, nothing but protocol on stdout |
| `wire` | 15 | headers, `max_tokens`, valid JSON, alternating non-empty roles |
| `calls` | 10 | exactly one API call per round |
| `replies` | 10 | the answer equals the text the server returned |
| `memory` | 25 | the round-4 request still carries the whole history |
| `growth` | 15 | each request extends the previous one byte-for-byte |
| `usage` | 10 | cumulative totals reported and **exactly** correct |

---

# Part 1 — Must fix

These are cases where a competent student fails for the **wrong reason**: the
grader is right, but the chapter never told them the rule. Fix before anyone
attempts the exercise.

## R1. State how the two modes are selected (silent-failure bug)

**Anchor** — Exercise section:

> **Two modes, one binary:** grader mode (the stdio contract above) and
> interactive mode (§1.6's REPL, live via proxy or the student's own API key).
> Same conversation code underneath.

**Problem.** It never says which mode is the **default**. The grader executes
the submission with **no arguments**. A student who defaults to the REPL
produces a program that waits for a human, the grader times out, and the
report says *"round 1: no reply on stdout (timeout)"* — a baffling diagnosis
for a trivial mistake.

**Proposed addition:**

> Grader mode is the default: run with **no arguments**, the program speaks
> the stdio protocol. The REPL is opt-in — `./ch01 chat`. The grader always
> invokes your binary with no arguments, so if your program greets a human on
> startup, it will hang the grader.

**Evidence.** The reference solution does exactly this (`solutions/ch01/main.go`,
`main()`); it is load-bearing and currently undocumented.

## R2. "stdout carries the protocol only; diagnostics go to stderr"

**Anchor** — Exercise section, **Contract** paragraph.

**Problem.** Not stated anywhere, but it **is** graded (`protocol`). This is
the single most likely innocent failure in the exercise, because every
programmer debugs with `fmt.Println`.

**Proposed addition to the contract:**

> stdout carries the protocol and nothing else — one JSON object per line.
> Send logs, progress and diagnostics to **stderr**. A stray `fmt.Println`
> is a protocol violation and will be reported as one.

**Evidence.** The `chatty` mutation in `grader_test.go` exists solely for
this, and fixing how the grader reports it was one of two real bugs the
mutation suite caught.

## R3. Name all three environment variables

**Anchor** — Exercise section:

> **Grading rig — a fake Anthropic server.** The grader sets
> `ANTHROPIC_BASE_URL` to a local fake…

**Problem.** The grader supplies **three** variables. A student who hardcodes
the model passes the fake (which accepts any non-empty model) but breaks on
the live path and the proxy — a failure that shows up later, far from its
cause.

**Proposed addition:**

> The grader supplies three variables, and your program should read all three:
> `ANTHROPIC_BASE_URL` (POST to `$BASE/v1/messages`), `ANTHROPIC_API_KEY`
> (send as the `x-api-key` header), and `ANTHROPIC_MODEL` (put in the `model`
> field). Hardcode none of them.

## R4. Usage is graded as an **exact** match

**Anchor** — Exercise section, check 4:

> **Usage accounting** — final report present, nonzero, consistent with the
> fake's returned usage fields.

**Problem.** "Consistent" is unfalsifiable, and it is weaker than what the
grader does. The fake counts tokens deterministically, so the expected totals
are known exactly — and exact-match is precisely what catches a student who
*invents* plausible numbers instead of summing responses.

**Proposed replacement:**

> **Usage accounting** — after stdin closes, the program reports cumulative
> input and output token totals that match the sum of the `usage` fields of
> every response, **exactly**. The fake's counter is deterministic (one token
> per four characters); that is a grading device, **not** a real tokenizer —
> do not infer anything about real token math from it.

The parenthetical matters: without it, an attentive student will "learn" a
tokenization rule that is fiction.

## R5. Reconcile the check list: four listed, seven graded

**Anchor** — Exercise section, the numbered list of what the fake checks.

**Problem.** The outline lists 4 checks; 7 are graded. Two of the missing
three are properties a student can't be expected to infer:

- **Exactly one API call per round.** Plausible-looking designs (a "warm-up"
  call, a retry, a second call to summarize) fail this. It is cheap to state.
- **Answers must equal the text the server returned.** This catches the
  laziest possible cheat — a program that never parses the response at all —
  and it is worth stating precisely *because* it is unbeatable: if you make up
  the answer, you fail even if every other check passes.

The third, `protocol`, is covered by R2.

**Proposed:** expand the list to the seven-row table in *Ground truth* above,
keeping the outline's existing prose for checks 1–4.

## R6. State both role rules, not just alternation

**Anchor** — §1.2:

> - `messages` — the array of `{role, content}`; roles alternate
>   `user`/`assistant`.

**Problem.** Two further rules are graded: the **first** message must be
`user`, and the **last** message must be `user` (otherwise there is nothing to
answer). Both are enforced by the fake; neither is stated.

**Proposed:** extend the bullet — roles strictly alternate, the conversation
begins with a `user` message, and the final message is always the `user`'s.

---

# Part 2 — Enrichment

Not correctness bugs. Each replaces a gesture with something measured, or
closes a gap I watched cause a real error.

## R7. §1.5 can print the actual cost curve

**Anchor** — §1.5:

> (No caching, no remedies — just the habit and the observed cost curve.)

**Problem.** The chapter promises a curve and never shows one. It can now.

**Measured, reproducible:**

- Offline, 5 rounds against the fake, input tokens per round:
  **49 → 78 → 95 → 116 → 138**
- Live, 3 rounds, `claude-sonnet-5`: **631 input / 388 output** total

The offline numbers regenerate identically on any machine
(`make grade` prints them), so they are safe to print in a book.

## R8. §1.2 deserves an "ask the API which models exist" sidebar

**Anchor** — §1.2:

> - `model`, `max_tokens` (required — and why the API refuses to guess).

**Problem.** Model IDs date, and — the part that bites — **aliases resolve
silently**. This recommendation is earned: while building this rig I selected
a model ID from memory because it looked familiar. It worked. It was also a
year old, having quietly resolved to a dated alias, and I would not have
noticed had Bill not asked.

**Proposed sidebar:**

> Don't take a model ID from a blog post, a tutorial, or your own memory — ask:
>
> ```bash
> curl -s https://api.anthropic.com/v1/models \
>   -H "x-api-key: $ANTHROPIC_API_KEY" \
>   -H "anthropic-version: 2023-06-01"
> ```
>
> An ID that looks current may be an alias for something much older, and
> nothing in the response will tell you so.

Secondary benefit: it inoculates the book's own listings against ageing, which
matters because they *will* age.

## R9. §1.2 should call out the request/response asymmetry

**Anchor** — §1.2:

> - Response: `content`, `stop_reason`, `usage`.

**Problem.** Request `content` may be a plain **string**, but response
`content` is **always an array of typed blocks**. That asymmetry is the
classic day-one stumble, and the outline lists the field name without
mentioning it.

**Proposed:**

> The response's `content` is not a string — it is a list of typed blocks.
> Walk it and concatenate the `text` blocks. (The request lets you send
> `content` as a bare string; the response never does. This asymmetry catches
> everyone once.)

## R10. §1.7 should say grading is free

**STATUS: APPLIED 2026-09-12 — and superseded in scope.** The "grading costs
$0" fact is now in §1.7. While applying it, Bill corrected a factual error the
original bullet contained (see D5): the $1,000–$10,000 figures were
misattributed to the course's own token burn. §1.7 now states **three
separate meters** instead of one fused figure. The anchor text quoted below no
longer exists.

**Anchor** — §1.7, the honest cost warning bullet ("$1,000, maybe more" /
"$10K in tokens").

**Problem.** The warning is right and should stay exactly as blunt as it is.
But it lands precisely where a reader decides whether to continue, and the
reassuring fact is missing: **grading costs nothing.**

**Proposed addition to that bullet:**

> To be clear about what is *not* costly: every graded exercise in this book
> runs against a local fake server. Grading needs no API key, reaches no
> network, and costs **$0**, as many times as you like. Only the optional live
> smoke tests spend real money, and those cost pennies.

## R11. §1.4 should name the 429 death in the body

**Anchor** — §1.4. Currently the no-retry ruling lives only in the resolved
*Open questions* list at the bottom, which is outline scaffolding and will not
survive into the chapter.

**Proposed addition:**

> There is no retry here, and no backoff. If the API returns 429, this program
> dies. That is the honest state of a naive client, and we are not going to
> paper over it.

---

# Part 3 — Do NOT add

Guard rails. Each of these is a plausible "improvement" that would damage a
standing ruling.

**D1. Do not caveat "the API is stateless."** §1.4's claim is flatly absolute
and must stay that way. The accurate footnote — that vendors now ship stateful
conversation APIs — telegraphs Chapter 2's demolition and violates the
**ambush ruling**. The sentence gets complicated later, by Chapter 2, cold.

**D2. No forward references to Chapter 2 anywhere.** Same ruling. Chapter 1
ends clean and confident. (The existing parenthetical in §1.6 about "what
Chapter 2's opening is aimed at" is *outline* scaffolding addressed to the
author — it must not survive into chapter prose.)

**D3. Do not add retry, backoff, or error handling.** Resolved: cut, on
minimalism grounds. See R11 — name the consequence, don't fix it.

**D4. Do not add streaming, tools, thinking, images, multi-provider support,
or system-prompt assembly.** The outline's existing "out of scope, by design"
list is correct and complete.

**D5. Do not soften the cost warning — and do not re-fuse it.** R10 adds a
true fact beside it; it does not dilute it. The $1,000–$10,000 figure **stays**.

**Amended 2026-09-12 by Bill.** The original bullet was factually wrong about
*whose* meter those dollars are on, and the error is easy to reintroduce
because the fused version sounds scarier and therefore more honest. It isn't.
§1.7 now states three meters and they must remain separate:

- reading the book — **$0**;
- the graded exercises — **$0** against the fake server, plus **$20–$100 for
  the whole book** if you run the optional live tests;
- building your own agent afterwards — **$1,000–$10,000**, which is
  **Claude Code / Codex assistant spend**, not the agent's own token burn,
  and which begins after the last chapter.

So: do not delete the big numbers (they are the honest tuition for building
the real thing), do not attach them to the course or to the toy agent, and do
not collapse the three meters back into one. The course is cheap; the thing
you build afterwards is not.

---

# Part 4 — Facts available to cite

**Verified this session** (reproducible, safe to print):

- `claude-sonnet-5` exists, created 2026-06-29, and is what
  `solutions/ch01` now defaults to.
- `claude-sonnet-4-5` silently resolves to `claude-sonnet-4-5-20250929`.
- Offline token curve: 49 → 78 → 95 → 116 → 138 (R7).
- Live 3-round cost on Sonnet 5: 631 input / 388 output (R7).
- The grader catches 9 distinct defect classes, each with an exact expected
  failure set (`grader_test.go`).

**NOT verified — keep the existing flag:** every figure in §1.0 (the SpaceX /
Cursor ~$60B, the Windsurf deal shapes, the Ona and Bun acquisitions). The
outline's publication note already says to check these against primary sources
before print. That note must survive; I did not verify any of them, and a
book-opening number that turns out to be wrong is expensive.

---

# Suggested order of work

1. R1, R2, R3 — the contract gaps. Nothing else matters until a student can
   attempt the exercise without failing for a reason the chapter never
   mentioned.
2. R4, R5, R6 — align the spec with the seven checks that exist.
3. R7–R11 — enrichment, in any order.
4. Re-read against Part 3 before declaring done.
