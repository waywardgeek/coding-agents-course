# Ch2 code rework against amended §2.4a — coder's report

Round 2. Brief: `book/brief-ch02-code-v2.md`. Baseline at start of this job:
build clean, gofmt clean, reference 100/100, 16/16 mutants detected.

---

## Part 0 — The brief's premise was partly stale (say it plainly)

The brief says "your code currently implements the *previous* §2.4a." That is
**not** what I found. Three of the four structural changes were already in the
solution before this job started:

| §2.4a change | state at job start |
|---|---|
| `ToolCallPart.Opaque` | **already present** (`part.go`) |
| `Entry.Seq` | **already present** (`context.go`) |
| `Surface` has ChatCompletions / GenerateContent | **already present** (`provenance.go`), but in a different order and with a now-false comment |
| `{"log_version":1}` header line | **already present** (`log.go`) |

This is not a coincidence and it is not the author being wrong. Those changes
*originated* in the last round's build: the code grew the fields first, the
findings were reported, and the outline then adopted them. The brief was
written from the outline's point of view, where they look new.

**The consequence worth noting under P2:** the drift that remained was not in
the struct shapes at all. It was in *comments that still described the old
spec*, and in one piece of behavior nothing in Chapter 2 exercises. Field-level
diffing would have reported "no drift" and been wrong.

---

## Part 0.5 — THE HEADLINE: §2.4a's newest field was graded by nothing

**`ToolCallPart.Opaque` was exercised by no check, in either direction. I proved
it by deleting its only use and scoring the reference: 100/100.**

```
$ perl -0pi -e 's/^\s*p\.ThoughtSignature = call\.Opaque.*$//m' gemini.go
$ go run ./cmd/grade -ch 2 /tmp/ch02_noopaque
score: 100/100 — PASS
```

So a student could read §2.4a, decide the field looked like over-modeling,
omit it, and be told by our grader that their Chapter 2 is perfect — and then
meet the 400 on a Tuesday in Chapter 3, which is the exact fate the chapter
added the field to prevent. The book would have kept its promise in prose and
broken it in the grader.

Why it slipped through, precisely — two independent reasons, each sufficient:

1. **The fixture's tool call carries no opaque material.** `ExhibitLog`'s
   `tool_call` part has `call_id`, `from`, `name`, `args` and no `opaque`.
   There was nothing to replay, so nothing could notice it was not replayed.
2. **The fixture's provenance is Anthropic.** Even had the field been
   populated, the renderer correctly gates replay on `SameModel`, so rendering
   that log to Gemini would omit it *by design*. A test would have asserted
   absence and passed either way.

The standalone `OpaquePart` is thoroughly graded — replayed to Anthropic,
withheld from OpenAI. That is what made the gap invisible: greping for
"opaque" in the grader returns hits, and they are all the *other* kind of
opaque. The distinction between a thinking block that belongs to the turn and a
signature bound to one call — the distinction the whole field exists for — is
precisely the distinction the tests did not draw.

**Consequence for the brief's requested mutant.** Fable asked for a mutant that
drops `ToolCallPart.Opaque` on a Gemini render, "since that is the 400 the
change exists to prevent." As specified it would not have been detected — it
would have been a mutant that no check kills, which is a silently passing test.
The grading had to be built first. It now is (Part 3).

---

## Part 1 — Defects found and fixed

### D1. `RedactSummary` duplicates the summary across the span (correctness)

`applyRedaction` is a single loop over entries in the span, with one `case` per
level. That is correct for three of the four levels and wrong for the fourth.

`RedactResult`, `RedactTool` and `RedactDialogue` are **per-entry filters**:
apply to each entry independently, N entries in, N entries out.
`RedactSummary` is a **fold**: §2.4a says "the span is replaced by compressed
prose", and §2.4a's own defense of `Dialogue` says "compaction replaces entries
with a summary entry" — singular.

Implemented uniformly, a summary over a span of N entries writes the *same*
`Replacement` into all N, so the compacted context repeats the summary N times.
Compaction then *grows* the context it was called to shrink.

Nothing in Chapter 2 catches this: the chapter exercises only `RedactResult`.
It would be inherited, under write-once, by the context-engineering chapter —
i.e. by the one chapter whose entire job is this operation.

The tidy uniform loop is what causes it. Three of four cases genuinely are
per-entry, so the fourth gets written as though it were.

### D2. The `Surface` enum carried a comment that the amended outline falsifies

The constants were ordered Messages, Interactions, Responses, ChatCompletions,
GenerateContent, under a comment reading "two surfaces the chapter's own
exhibits use but §2.4a's enum omitted."

§2.4a no longer omits them — that was the point of change (c). A student
rebasing onto this solution (P2) would read a comment telling them the book has
a gap the book does not have. Reordered to match §2.4a exactly and rewrote the
comment. Numeric values are not on the wire (marshaled as readable strings), so
this is a documentation fix, not a behavioral one.

### D3. The reducer's `default` arm explained itself with the wrong rule

The comment on the identity `default` said, in substance, that an agent which
halts on an event it does not recognize is worse than one that ignores it.

That is the exact opposite of §2.7: "An unknown event type is a refusal to
load, loudly. Not a skip."

Both rules are right; they govern different things, and the comment collapsed
them:

- **§2.3** — every *(state, event)* pair not in the table is **identity**. A
  known event in a state that does not transition. `default` arm, not a panic.
- **§2.7** — an unknown *event type string* is a **loud refusal at load time**.

An unknown type can never reach the reducer, because the loader rejects it
first. The comment made the reducer look like the place that tolerates unknown
events, which would license exactly the silent-skip §2.7 forbids.

---

## Part 2 — Reported to the author, not fixed by me

### A1. §2.4a does not say what a `RedactSummary` replacement *is*

Fixing D1 forced two decisions the spec does not make, and I had to make them
to write the code:

1. **Seq of the collapsed entry.** Used `From` — the span's own start, already
   in the event, deterministic under replay.
2. **Actor of the collapsed entry.** A span can cross Human, Agent and Tool. A
   summary of a multi-actor span is not any of their speech; it is compaction
   output. Used `System`, which §2.5 already models.

Both are defensible and neither is in the outline. They are the kind of thing
that, left unstated, two chapters will answer differently. Recommend §2.4a say
so in one sentence, since the context-engineering chapter depends on it.

### A2. §2.8's description of `seam-render` is now too narrow

§2.8 describes `seam-render` as "one log renders correctly to all three vendor
request shapes." Closing the `Opaque` gap required a **second** fixture, for the
reason in Part 0.5: the property needs a log whose provenance matches the
render target, and `ExhibitLog` is Anthropic-authored by design.

I folded the assertion into `seam-render` rather than adding a check, because
adding one would mean re-dividing 100 points and that is the author's call, not
the coder's. The point value is unchanged at 15 and the sum is still exactly
100. But the prose now under-describes the check. Suggested amendment: "one log
renders correctly to all three vendor request shapes, and per-call replay
material survives a round trip back to the model that issued it."

The failure messages are kept distinct so a student is still told *which*
property broke — the same granularity argument §2.8 makes when it explains why
`usage` was split out of `seam-parse`.

### A3. Not a defect — recording why `RedactSummary` is ungraded and `Opaque` was not

These two look like the same omission and are not, which is worth stating so a
later session does not "fix" the wrong one.

`RedactSummary` is ungraded **deliberately**: §2.4a's own table lists redaction
levels above `RedactResult` as "not exercised" in Chapter 2. The chapter says it
is building shape ahead of capability, so an ungraded level is the plan working.
D1 was still a real bug — the shape was wrong, not just untested — which is why
I verified it by hand (below) rather than by a check.

`ToolCallPart.Opaque` was different: §2.4a lists it as shipped, needed, and
load-bearing *now*, and the chapter's §2.6 table prints it as the one field the
seam bet cost. A field the book advertises as the thing standing between you and
a 400 should not be omissible for full marks.

---

## Part 3 — What changed

### Grader

- **`GeminiReplayLog`** (`internal/grade/ch02_log.go`) — new fixture. Gemini
  provenance, a tool call carrying `opaque`, model matching
  `ch2RequestedModel("gemini")` in all three fields, since `SameModel` compares
  vendor, model **and** surface.
- **`Ch2Result.ThoughtReplay`** + phase 4b (`ch02_harness.go`) — one extra
  `render`, no network.
- **`checkGeminiThoughtReplay`** (`ch02_checks.go`) — asserts the replayed
  `functionCall` carries `thoughtSignature` as a **sibling key**, with the
  recorded value, and that the assistant turn is present at all. Folded into
  `seam-render`; points unchanged.
- **Mutant `gemini-call-opaque-dropped`** — 17 mutants now, all detected, still
  asserting the exact set of failing ids (`{seam-render}`).

### Solution

- **D1** `applyRedaction` lifted `RedactSummary` out of the per-entry loop into
  `summarizeSpan`.
- **D2** `Surface` constants reordered to §2.4a's order; false comment replaced.
- **D3** reducer `default` comment rewritten to distinguish §2.3 identity from
  §2.7 loud refusal.
- **P7 egg** planted in `solutions/ch02/part.go`, in the `OpaquePart` doc
  comment, quoted as a decoded replay block. Ledger row appended.

The egg's placement is the joke rather than a hiding place: it sits inside the
one type whose stated contract is *carried, never interpreted*, where any model
reading the file to answer a question about the file will interpret it. Chapter
7 gets a reveal that indicts the mechanism and not just the reader. It is an
ordinary Go comment — no white text, no zero-width characters — and no check
reads source, so it cannot move a score.

---

## Part 4 — Verification

| gate | result |
|---|---|
| `go build ./...` | exit 0 |
| `gofmt -l .` | clean |
| `go vet ./...` | exit 0 |
| reference score | **100/100** |
| mutation suite | **17/17 detected**, exact-set assertions intact |
| §2.8 points sum (awk on the points column) | **100** |

**D1 verified in both directions**, because no check covers it. A log
summarizing a three-entry span, rendered to Anthropic:

| build | copies of the summary | messages |
|---|---|---|
| before the fix | **3** | 3 |
| after the fix | **1** | 1 |

Compaction was multiplying the thing it was called to remove.

**The headline verified in both directions too**: deleting the render-side use
of `ToolCallPart.Opaque` scored **100/100** before this job and **85/100**
after, failing exactly `seam-render` with a message naming the 400.

---

## Part 5 — Method note, because it generalizes

The three defects and the headline have one shape in common: **every one of
them was invisible to a field-by-field diff of the code against §2.4a.** The
structs already matched. What did not match were a comment asserting a gap the
spec had closed, a comment asserting the opposite of §2.7, a level of a family
the chapter does not exercise, and a field nothing tested.

A "does the code implement the spec?" pass answers yes here and is wrong four
times. The questions that actually found things were: *what does this comment
claim, and is it still true?* and *what would still pass if I deleted this?*

The second one is just mutation testing pointed at the spec instead of the
code, and it is the one I would run first next time. The grader is an artifact
that can rot exactly like the guard documents did — a check that cannot fail is
a green dashboard with a schema around it.


