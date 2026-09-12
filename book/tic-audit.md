# Tic Audit — Chapters 1 and 2

Audited 2026-09-12 against:

- `book/voice.md`
- `book/chapter-01-outline.md` (383 lines, 53 em-dashes)
- `book/chapter-02-outline.md` (1262 lines, 146 em-dashes)

**Nothing in either outline was edited.** This file is verdicts only.

---

## Scope ruling applied

Judged as **draft prose**: flowing sentences meant to reach the reader,
including bulleted reader-facing exercise text.

Skipped as **authorial guidance**, per instruction:

- Chapter 2's entire `## Draft 4 — what changed and why` section (lines 9–58) —
  a draft changelog addressed to the author, never printed.
- All headings and sidebar titles.
- Go code listings, code comments, and JSON listings.
- All tables (grader tables, point tables, and content tables alike) — cells are
  compressed notation, not prose.
- Parenthetical author notes: `*(Publication note…)*`, `*(Author note…)*`,
  `*(For the coder…)*`, `> *Authorial note…*`, `*(Ratified…)*`.
- Chapter 2 §2.9 `Open questions for Bill` (1188–1247) and the Appendix
  (1250–1262).
- Chapter 2 lines 980–991 and 1170 — second-person instructions to the drafter
  ("Tell the reader…", "Do not let them…", "state the intended reading in the
  prose anyway").

---

## Summary

### TIC 1 — "X, not Y"

| chapter | instances in draft prose | KEEP | CUT |
|---|---|---|---|
| Ch 1 | 16 | 15 | 1 |
| Ch 2 | 45 | 39 | 6 |
| **Total** | **61** | **54** | **7** |

The construction is overwhelmingly **earned** in this book, and that is a real
finding rather than a soft grade. The whole subject matter is reflexes that
fail — store the system prompt, use a flag not a state, return a message from
`Parse`, treat usage as bookkeeping — so the Y half is nearly always something
the reader was about to do. The seven CUTs are the cases where the author is
answering a charge nobody made, or restating a correction he already landed.

### TIC 2 — em-dashes

| chapter | total | guidance (skipped) | prose | KEEP (SETUP) | CUT (APPOSITIVE) | RANGE/COMPOUND |
|---|---|---|---|---|---|---|
| Ch 1 | 53 | 11 | 42 | 11 | 31 | 0 |
| Ch 2 | 146 | 52 | 94 | 27 | 67 | 0 |

**RANGE/COMPOUND count is zero in both chapters.** Every true range in these
files (`R1–R11`, `$20–$100`, `0%→98%`) is an *en*-dash or an arrow and was never
in the count. There is no legitimate typographic em-dash use here at all: every
one of the 199 is doing rhetorical work, which is itself the diagnosis.

### Does the ch2 CUT list hit the target?

**No. It undershoots by 29.**

- Chapter 2 starts at **146**.
- My CUT list removes **67**.
- Chapter 2 lands at **79**. Target was ~50.

Why the gap, stated precisely so you can decide how much further to go:

1. **52 of the 146 are in guidance** — the Draft-4 changelog, §2.9's open
   questions, authorial notes, code comments, tables. I was told to skip them,
   so they survive untouched. They alone are more than the 50-dash target.
2. **Only 94 em-dashes are in draft prose.** Cutting 67 of 94 is an 71% cut
   rate, and the 27 survivors are genuine punchline/reversal beats — the "It
   costs you a tail — and the tail is paid by whoever is using the product while
   you migrate" class. Cutting those would flatten the voice, and `voice.md`
   wants those landings.

**To actually reach ~50 you must open one of two doors:**

- **Door A (recommended): let me cut guidance too.** The authorial notes and the
  Draft-4 section are as dash-infested as the prose — §2.9 alone holds 11. Cut
  ~29 there and you reach 50 without touching a single beat the reader sees.
  Guidance is where nobody is watching, which is exactly why the tic
  concentrated there.
- **Door B: sacrifice ~29 SETUP dashes.** That means cutting almost every
  landing beat in the chapter. I do not recommend it; it buys a number and
  spends the voice.

Chapter 1 goes **53 → 22** under its CUT list. No target was set for ch1, but
note the prose there is already tighter: 42 prose dashes in 383 lines versus
94 in 1262 — comparable density, so ch2's problem is length, not a worse habit.

### Low-priority CUTs, flagged

A subset of my CUTs are **label dashes** — `**Bold term** — gloss` or
`` `term` — definition ``. These are conventional definition-list typography, not
the machine tell. I classify them APPOSITIVE and cut them because a colon or
comma serves identically and you asked for the count down, but they are the
first ones to restore if a pass reads as over-corrected:

- Ch 1: lines 23, 25, 107, 108, 229, 284, 309, 319
- Ch 2: lines 247, 858, 866, 872, 966, 967, 970, 1154

That is 16 of the 98 total cuts.

---

## Section 1 — Chapter 1, "X, not Y" verdicts

Ordered by line number ascending.

| # | line | quote | verdict |
|---|---|---|---|
| 1 | 36 | "the value was never the code, it was what I learned building it" | **KEEP** |

Reason: the reader has just been told he wrote a PoC in two weeks to win a bet;
the artifact is obviously the deliverable. Correcting that is the book's thesis.

| # | line | quote | verdict |
|---|---|---|---|
| 2 | 48–49 | "buying the **execution layer** … not chat wrappers" | **KEEP** |

Reason: the named acquisitions are editors and chat products, and "thin wrapper
on Claude" is the standard dismissal of the whole category. The reader really
does hold Y.

| # | line | quote | verdict |
|---|---|---|---|
| 3 | 74 | "**The mechanism, not just the slogan.**" | **CUT** |

Reason: nobody expected to be handed "just the slogan" — the preceding text gave
a decision rule, not sloganeering. Both halves are the author's invention, and
the phrase mostly congratulates the section it introduces.
**Replacement:** `**The mechanism.**`

| # | line | quote | verdict |
|---|---|---|---|
| 4 | 84 | "as a product toggle, not an API their frameworks let you reach" | **KEEP** |

Reason: the previous clause says frontier products ship real-time steering. A
reader told a product has a feature does assume it is reachable — that is the
entire complaint of §1.1.

| # | line | quote | verdict |
|---|---|---|---|
| 5 | 115 | "The response never does — response `content` is always a list of typed blocks" | **KEEP** |

Reason: maximally earned. Same field name, different shape; the outline itself
calls it "the classic day-one stumble."

| # | line | quote | verdict |
|---|---|---|---|
| 6 | 133 | "Ask; don't remember." | **KEEP** |

Reason: the author has just confessed to doing exactly Y one paragraph earlier.
The correction is receipted.

| # | line | quote | verdict |
|---|---|---|---|
| 7 | 169 | "(No caching, no remedies — just the habit and the observed cost curve.)" | **KEEP** |

Reason: a reader shown a rising cost curve genuinely expects the fix in the same
section. This is scope-limiting, which is legitimate work.

| # | line | quote | verdict |
|---|---|---|---|
| 8 | 209–211 | "It does **not** bite in this course … It bites the day you point a finished agent at real work." | **KEEP** |

Reason: the bullet opens by saying you cannot cheaply get a fast key, which
reads as an immediate blocker on the course itself. Strong live expectation.

| # | line | quote | verdict |
|---|---|---|---|
| 9 | 212–213 | "removes is *having a provider account at all*, not the throughput ceiling" | **KEEP** |

Reason: same paragraph spent its length on the throughput gate, so the reader
absolutely assumes the proxy solves it. Correcting it prevents a refund request.

| # | line | quote | verdict |
|---|---|---|---|
| 10 | 227 | "The proxy exists to remove the toll booth, not to become one." | **KEEP** |

Reason: a Stripe-funded proxy run by the author invites exactly this suspicion.
Y is the reader's live objection, not an invention.

| # | line | quote | verdict |
|---|---|---|---|
| 11 | 238 | "a legitimate way to read this book, not a consolation prize" | **KEEP** |

Reason: the preceding sentence says a non-runner gets "most of the value,"
which reads as a downgrade. The correction repairs a real implication.

| # | line | quote | verdict |
|---|---|---|---|
| 12 | 243 | "budget **$20–$100 for the entire book**, not per chapter" | **KEEP** |

Reason: cost figures in a chapterized book are genuinely ambiguous, and this
bullet exists because readers fuse numbers. Prevents a concrete misreading.

| # | line | quote | verdict |
|---|---|---|---|
| 13 | 243–245 | "The real cost here is not money, it is that the exercises are *sized for an engineer working with a coding assistant*" | **KEEP** |

Reason: the bullet is a cost warning denominated in dollars, so "money" is
precisely what the reader has in hand. A real pivot.

| # | line | quote | verdict |
|---|---|---|---|
| 14 | 281 | "they execute your binary and read what it emits, never your source or its history" | **KEEP** |

Reason: the paragraph has just invited the reader to start from the reference
solution. Plagiarism anxiety is the reader's live worry; Y is held.

| # | line | quote | verdict |
|---|---|---|---|
| 15 | 315–317 | "It does not reject on the first mistake. One run therefore tells you about *all* your bugs rather than one bug per run." | **KEEP** |

Reason: fail-fast is how every validator the reader has used behaves. Y is the
default assumption and the correction carries the design lesson.

| # | line | quote | verdict |
|---|---|---|---|
| 16 | 350–351 | "which is a *grading device*, **not** a real tokenizer" | **KEEP** |

Reason: the best-earned instance in the chapter. "One token per four characters"
is a famous rule of thumb and a student will absolutely bank it as fact.

**Chapter 1 TIC 1: 16 instances — 15 KEEP, 1 CUT (line 74).**

---

## Section 2 — Chapter 1, em-dash verdicts

31 CUTs, ordered by line number ascending. KEEPs listed after.

| line | quote | class | verdict | replacement |
|---|---|---|---|---|
| 23 | "**SpaceX acquires Cursor … June 2026** — an AI code editor valued above…" | APPOSITIVE (label) | **CUT** | comma: `…June 2026**, an AI code editor…` |
| 25 | "**The Windsurf drama** — one product, three suitors, three deal shapes:" | APPOSITIVE (label) | **CUT** | comma: `**The Windsurf drama**, one product…` |
| 38 ×2 | "writing the initial version of CodeRhapsody properly — production-worthy — and that is…" | APPOSITIVE | **CUT** | comma pair: `properly, production-worthy, and that is…` |
| 48, 49 | "the **execution layer** — runtimes, orchestration, secure execution — not chat wrappers" | APPOSITIVE | **CUT** | parentheses: `**execution layer** (runtimes, orchestration, secure execution), not chat wrappers` |
| 75 | "they **hard-code delivery** — what goes into the request payload…" | APPOSITIVE | **CUT** | colon: `**hard-code delivery**: what goes into…` |
| 81 | "*in the same message* — which is how a human redirects an agent…" | APPOSITIVE (trailing *which*) | **CUT** | comma: `…same message*, which is how…` |
| 86 | "never in history — because delivery *position* is the feature" | APPOSITIVE | **CUT** | comma: `…never in history, because delivery…` |
| 106 | "(required — and why the API refuses to guess)" | APPOSITIVE | **CUT** | comma: `(required, and why the API refuses to guess)` |
| 107 | "`system` — a string outside the messages array" | APPOSITIVE (label) | **CUT** | colon: `` `system`: a string outside… `` |
| 108 | "`messages` — the array of `{role, content}`" | APPOSITIVE (label) | **CUT** | colon: `` `messages`: the array of… `` |
| 115 | "The response never does — response `content` is always a list…" | APPOSITIVE | **CUT** | colon: `The response never does: response `content` is…` |
| 190, 191 | "an interactive mode — plain terminal REPL, live via the course proxy (§1.7) — where the student…" | APPOSITIVE | **CUT** | comma pair: `an interactive mode, a plain terminal REPL live via the course proxy (§1.7), where…` |
| 217 ×2 | "own the high-value tools on top — Claude Code, Claude Cowork, Codex — and to *discount tokens*…" | APPOSITIVE | **CUT** | parentheses: `on top (Claude Code, Claude Cowork, Codex) and to…` |
| 220, 221 | "(Same strategy as §1.0's acquisitions — buying the execution layer — pointed downmarket…)" | APPOSITIVE | **CUT** | comma pair: `(…acquisitions, buying the execution layer, pointed downmarket…)` |
| 229 | "**The honest cost warning — three meters, and everyone confuses them.**" | APPOSITIVE (label) | **CUT** | colon: `**The honest cost warning: three meters…**` |
| 242 | "and they are small — budget **$20–$100 for the entire book**" | APPOSITIVE | **CUT** | colon: `they are small: budget **$20–$100…**` |
| 251 | "pay **Claude Code or Codex** to help you write it — the assistant spend it takes…" | APPOSITIVE | **CUT** | comma: `…write it, the assistant spend it takes…` |
| 271 | "**Go required** — the book's code is Go, and later chapters build on this program." | APPOSITIVE (label) | **CUT** | colon: `**Go required**: the book's code is Go…` |
| 284 | "**stdout carries the protocol and nothing else** — one JSON object per line." | APPOSITIVE (label) | **CUT** | colon: `…nothing else**: one JSON object per line.` |
| 305, 306 | "passes the fake — which accepts any non-empty model — and then breaks…" | APPOSITIVE | **CUT** | parentheses: `passes the fake (which accepts any non-empty model) and then breaks…` |
| 309 | "**Grading rig — a fake Anthropic server.**" | APPOSITIVE (label) | **CUT** | colon: `**Grading rig: a fake Anthropic server.**` |
| 319 | "**The seven checks — 100 points, and all of them must pass:**" | APPOSITIVE (label) | **CUT** | parentheses: `**The seven checks (100 points, and all of them must pass):**` |
| 341 | "catches the laziest possible cheat — a program that never parses the response at all" | APPOSITIVE | **CUT** | comma: `…possible cheat, a program that never parses…` |
| 350, 351 | "The fake's counter is deterministic — one token per four characters, rounded up — which is…" | APPOSITIVE | **CUT** | parentheses: `deterministic (one token per four characters, rounded up), which is…` |

**Chapter 1 em-dash KEEPs (SETUP, 11):** lines 35, 46, 53, 100, 110, 132, 169,
180, 205, 222, 274.

Each lands a beat rather than interrupting one — 35 ("a vibe-coded pile of shit
I had to throw away"), 132 ("It worked — and it was a year old"), 180 ("the
fifth question costs nearly three times the first"), 205 ("fast enough —
eventually"). These are the voice.

**Chapter 1 totals: 53 em-dashes → 11 guidance, 11 KEEP, 31 CUT → lands at 22.**

---

## Section 3 — Chapter 2, "X, not Y" verdicts

45 instances in draft prose. Ordered by line number ascending. To keep this
readable I list the 6 CUTs in full and the 39 KEEPs in a compact table, since
every KEEP verdict has the same shape: the reader plausibly holds Y.

### The 6 CUTs

**Line 155 — "This is also why `Provenance` records a **surface** and not merely a vendor (§2.4a)."**
**CUT.** The correction is real, but the *very next sentence* — "The vendor is
not the unit of compatibility" — makes the same correction better and keeps it.
Landing it twice in two sentences is the filler, not the idea.
**Replacement:** `This is also why `Provenance` records a **surface** (§2.4a).`
(Keep line 156 exactly as it stands; it carries the load.)

**Line 159 — "`SurfaceInteractions` is in the enum because this was foreseeable, not because it was foreseen."**
**CUT.** Nobody accused the author of prescience. The antithesis is a pun on two
forms of one word, and the humility survives without it.
**Replacement:** `` `SurfaceInteractions` is in the enum because this was foreseeable. ``

**Line 510 — "Signature validity is scoped to *(vendor, model, surface)* — not to vendor:"**
**CUT.** Third statement of the same correction inside thirty lines (480, 498,
510). The two bullets underneath it prove the point; the disclaimer is dead
weight, and cutting it removes an em-dash too.
**Replacement:** `Signature validity is scoped to *(vendor, model, surface)*:`

**Line 885–886 — "That is the entire argument for the seam, and it is not an analogy — the student will watch…"**
**CUT.** The preceding sentence is a plain technical claim about three wire
formats. No reader was going to file it as an analogy; the author is rebutting
a charge nobody made.
**Replacement:** `That is the entire argument for the seam: the student will watch one `Actor: Tool` event become three different claims about authorship, and none of the three is worth storing.`

**Line 950–951 — "Vendor is not a fine enough grain — see `Provenance` in §2.4a."**
**CUT.** Fourth restatement. By §2.6 the reader has been told three times and
has seen the measurement table.
**Replacement:** `Match the grain to the model (see `Provenance` in §2.4a).`

**Line 1030 — "you need to already know, not hope, that your context assembly is correct"**
**CUT.** "Know, not hope" is a stock intensifier. Nobody advocates hoping; both
halves are invented and the positive claim is stronger alone.
**Replacement:** `you need to already know that your context assembly is correct`

### The 39 KEEPs

| line | quote | why it earns it |
|---|---|---|
| 133 | "that is our bug, not yours" | students blame themselves by default |
| 137–138 | "not mainly for people running three models. It is for people running one" | the chapter title itself plants Y |
| 146–147 | "not a hypothetical migration used to motivate a design. It is a live one" | invented motivating migrations are an architecture-book cliché |
| 156 | "The vendor is not the unit of compatibility." | every framework the reader has used is vendor-keyed |
| 198 | "a request is not a description of the current moment" | the sidebar documents the author making exactly this error |
| 209–210 | "a new event that supersedes, not a mutation of an old one" | mutate-in-place is the reflex implementation |
| 230 | "It is *not* a demand for value semantics." | the functional notation genuinely implies Y to a Go reader |
| 236 | "Classification is the reducer's job, not the capture site's" | capture time is the obvious place |
| 286 | "it must be a *state*, not a flag" | `interrupted bool` is the first thing anyone writes |
| 298 | "That sentence, not the length of the table, is what makes the reducer total" | readers credit the enumeration; that is the misreading |
| 312–313 | "not tracked separately — it is replaced in place" | an earlier draft did track it separately |
| 315 | "**Content is Parts, not a string.**" | the reader shipped a string in Chapter 1 |
| 325 | "not a blob of text living in the log, and … not a field on the context" | storing it is universal practice |
| 356 | "Shapes, not implementations." | it is a Go listing; readers copy Go listings |
| 480 | "bound to the model, not the vendor" | the surrounding tables are vendor-keyed; strongest of the three |
| 498 | "about **integrity, not authorship**" | the paragraph just proved the reader's authorship prior false |
| 559–560 | "Not a default, not a skip" | lenient parsing is the reflex |
| 571 | "**derived from `Seq`**, not generated randomly" | a UUID is the obvious move |
| 603 | "**`Entry.Seq` is not decoration.**" | a seq on a derived entry does look redundant |
| 615 | "The context is not a request buffer." | the reader built exactly that in Chapter 1 |
| 631–632 | "not metadata about content — **it is content**" | the draft's own `map[Seq]bool` made this mistake |
| 680 | "**Stubs are synthesized, not stored.**" | storing the replacement is the obvious design, and `RedactSummary` does |
| 701 | "`Usage` looks like bookkeeping and is not." | Chapter 1 treated it as bookkeeping |
| 741 | "**Record counts, never money.**" | cost is what people want to see; frameworks log it |
| 762 | "it looks like a violation and is not" | the no-vendor-vocabulary rule was just stated absolutely |
| 788 | "`Parse` returns **events**, not a message or a context" | every SDK's parse returns a message |
| 835 | "a property of the room, not of the message" | a `To` field is the natural first design |
| 909 | "not a fact about the conversation; it is a fact about a wire format" | storing the merged shape is what Ch1's structure forces |
| 927 | "That 'apart from' is not a loophole" | an exception clause in a byte-identity check does look like a fudge |
| 937 | "an HTTP 429 is an `ErrorOccurred`, not a response" | it arrives on the response path from the same call |
| 948 | "carried, never interpreted" | decoding a thinking signature is a real temptation |
| 976 | "true because the third was *easy*, not because the seam was right" | names a specific false positive in the chapter's own test |
| 1011 | "not because anything about agent architecture changed, but because a vendor deprecated a surface" | "second edition" implies the architecture moved |
| 1022 | "You are not building it to add vendors." | the chapter is titled *Three Vendors*; Y is maximally planted |
| 1028 | "not a difficulty ramp, it is establishing a control" | easy-to-hard is the default pedagogical read |
| 1050 | "Bent, then, not broken" | the table just showed Gemini forcing a field; the reader is scoring it a failure |
| 1062 | "Replay with **current code**, not with historical code." | event-sourcing intuition says the opposite |
| 1072 | "**a refusal to load, loudly.** Not a skip." | lenient parsing reflex again, with consequence named |
| 1075 | "Retention is **policy**, not architecture" | baking rotation into the log type is a common conflation |
| 1133 | "not second-class citizens" | a three-vendor exercise does imply three accounts |

**Chapter 2 TIC 1: 45 instances — 39 KEEP, 6 CUT (155, 159, 510, 885, 950, 1030).**

---

## Section 4 — Chapter 2, em-dash verdicts

67 CUT occurrences across 58 lines, ordered by line number ascending.

| line | quote | class | verdict | replacement |
|---|---|---|---|---|
| 71 ×2 | "Extract an interface — `AIClientInterface` — with all the methods needed" | APPOSITIVE | **CUT** | comma pair: `an interface, `AIClientInterface`, with all…` |
| 82, 83 | "eventually designed — one context, one renderer per vendor, one parser per vendor — and delivered" | APPOSITIVE | **CUT** | parentheses: `designed (one context, one renderer per vendor, one parser per vendor) and delivered` |
| 143 ×2 | "its replacement — the Interactions API — is not yet available on Vertex AI" | APPOSITIVE | **CUT** | comma pair: `its replacement, the Interactions API, is not yet available` |
| 151, 152 | "Without one — with a data structure shaped like a particular vendor's request body … — it is a rewrite" | APPOSITIVE | **CUT** | parentheses: `Without one (with a data structure shaped like a particular vendor's request body, which is §2.0's mistake) it is a rewrite` |
| 183 | "you buy the ability to edit history — which is the core capability of a coding agent" | APPOSITIVE (trailing *which*) | **CUT** | comma: `edit history, which is the core capability…` |
| 211 | "greppable with ordinary tools — a property that becomes load-bearing in Chapter 3" | APPOSITIVE | **CUT** | comma: `ordinary tools, a property that becomes…` |
| 221 ×2 | "just the next event — nothing else — we can correctly compute" | APPOSITIVE | **CUT** | parentheses: `just the next event (nothing else) we can…` |
| 232 | "that is the one to write — a `Context` full of slices copied by value gives you two contexts" | APPOSITIVE | **CUT** | colon: `the one to write: a `Context` full of slices…` |
| 247 | "**Additive, always** — this is the first place the write-once discipline is visible" | APPOSITIVE (label) | **CUT** | colon: `**Additive, always**: this is the first place…` |
| 256, 257 | "opaque replay material — a signature, a redacted block, an id — tagged with the exact model" | APPOSITIVE | **CUT** | parentheses: `replay material (a signature, a redacted block, an id) tagged with…` |
| 263 | "holds everything the assistant produced — text and `ToolCallPart`s together" | APPOSITIVE | **CUT** | colon: `everything the assistant produced: text and…` |
| 267, 268 | "The alternative — a `ToolCalled` per call, with `ResponseEnded` carrying only text — throws away" | APPOSITIVE | **CUT** | parentheses: `The alternative (a `ToolCalled` per call, with `ResponseEnded` carrying only text) throws away` |
| 286 | "(`Interrupted` arrives in Chapter 4 — and it must be a *state*, not a flag…)" | APPOSITIVE | **CUT** | semicolon: `Chapter 4; and it must be a *state*…` |
| 313 | "not tracked separately — it is replaced in place by the reducer" | APPOSITIVE | **CUT** | semicolon: `separately; it is replaced in place…` |
| 479 ×2 | "**Thinking signatures — the encrypted reasoning material — are bound to the model**" | APPOSITIVE | **CUT** | comma pair: `Thinking signatures, the encrypted reasoning material, are bound…` |
| 490 | "silently discards a *newer* model's — while returning 400 for one that has been modified" | APPOSITIVE | **CUT** | comma: `model's, while returning 400…` |
| 510 | "scoped to *(vendor, model, surface)* — not to vendor:" | APPOSITIVE | **CUT** | delete whole clause (see TIC 1, line 510): `scoped to *(vendor, model, surface)*:` |
| 527 | "Inference is impossible in principle here — by the time you are rendering…" | APPOSITIVE | **CUT** | colon: `impossible in principle here: by the time…` |
| 539 | "closed sets the renderer *switches on* — there is exactly one renderer and parser per vendor" | APPOSITIVE | **CUT** | colon: `*switches on*: there is exactly one…` |
| 551 | "genuine Anthropic/messages one — and since provenance must be captured at write time…" | APPOSITIVE | **CUT** | period: `…messages one. And since provenance…` |
| 566 | "that legitimately enters the context — you cannot answer a call without quoting the id" | APPOSITIVE | **CUT** | colon: `enters the context: you cannot answer…` |
| 606 | "a `map[Seq]bool` kept off to the side — which is precisely the unbounded field" | APPOSITIVE (trailing *which*) | **CUT** | comma: `off to the side, which is precisely…` |
| 616 | "an actor that may run for **years** — memory, identity, recent conversation, everything" | APPOSITIVE | **CUT** | colon: `may run for **years**: memory, identity…` |
| 636, 637 | "bounded by a *policy* — compaction and retention, a later chapter — rather than by its shape" | APPOSITIVE | **CUT** | parentheses: `by a *policy* (compaction and retention, a later chapter) rather than by its shape` |
| 644 ×2 | "The shape above — a **span**, a **level**, and an optional replacement — looks like over-modeling" | APPOSITIVE | **CUT** | parentheses: `The shape above (a **span**, a **level**, and an optional replacement) looks like…` |
| 655 ×2 | "The common framework approach — Google's ADK does this — is to replace the oldest *portion*" | APPOSITIVE | **CUT** | parentheses: `approach (Google's ADK does this) is to replace…` |
| 681, 682 | "from the event it supersedes — the tool's name, the size, the path the output still lives at — which makes it" | APPOSITIVE | **CUT** | parentheses: `it supersedes (the tool's name, the size, the path the output still lives at), which makes it` |
| 695, 696 | "The policy — what thresholds trigger which level, and where the boundaries fall — is context engineering" | APPOSITIVE | **CUT** | parentheses: `The policy (what thresholds trigger which level, and where the boundaries fall) is context engineering` |
| 724, 725 | "Normalize naively — sum everything you are given — and you double-count" | APPOSITIVE | **CUT** | parentheses: `Normalize naively (sum everything you are given) and you double-count` |
| 749 | "bills cache *storage* by duration — a real cost with no token count attached" | APPOSITIVE | **CUT** | comma: `by duration, a real cost with no token count attached` |
| 764 | "The first is history — it happened, it is not re-derivable, and throwing it away is lossy" | APPOSITIVE | **CUT** | colon: `The first is history: it happened…` |
| 789 ×2 | "exactly one path into the context — append events, run the reducer — so a vendor response" | APPOSITIVE | **CUT** | parentheses: `one path into the context (append events, run the reducer) so a vendor response` |
| 835 | "**no `To` field** — addressing is a property of the room" | APPOSITIVE | **CUT** | colon: `**no `To` field**: addressing is a property…` |
| 858 | "**Anthropic** — a `tool_result` block inside a **user** message:" | APPOSITIVE (label) | **CUT** | comma: `**Anthropic**, a `tool_result` block…` |
| 866 | "**OpenAI** — its own message with a **tool** role:" | APPOSITIVE (label) | **CUT** | comma: `**OpenAI**, its own message…` |
| 872 | "**Gemini** — a `functionResponse` part in a **user** turn:" | APPOSITIVE (label) | **CUT** | comma: `**Gemini**, a `functionResponse` part…` |
| 879 | "Anthropic says the human did — which is false, and is the tidiest available lie" | APPOSITIVE (trailing *which*) | **CUT** | comma: `the human did, which is false…` |
| 886 | "it is not an analogy — the student will watch one `Actor: Tool` event" | APPOSITIVE | **CUT** | colon, with clause deleted per TIC 1: `argument for the seam: the student will watch…` |
| 936 | "and errors — an HTTP 429 is an `ErrorOccurred`, not a response" | APPOSITIVE | **CUT** | colon: `and errors: an HTTP 429 is…` |
| 951 | "not a fine enough grain — see `Provenance` in §2.4a" | APPOSITIVE | **CUT** | parentheses, per TIC 1 rewrite: `(see `Provenance` in §2.4a)` |
| 966 | "**Anthropic** — the baseline. Everything the student already has." | APPOSITIVE (label) | **CUT** | colon: `**Anthropic**: the baseline.` |
| 967 | "**OpenAI** — a moderate difference: a `tool` role of its own" | APPOSITIVE (label) | **CUT** | comma: `**OpenAI**, a moderate difference: …` |
| 970 | "**Gemini** — the genuinely alien one." | APPOSITIVE (label) | **CUT** | colon: `**Gemini**: the genuinely alien one.` |
| 1007 | "not available on Vertex AI — the platform Google sells to exactly the enterprises" | APPOSITIVE | **CUT** | comma: `on Vertex AI, the platform Google sells…` |
| 1017 ×2 | "polling a vendor's roadmap — weekly? monthly? — is left as an exercise" | APPOSITIVE | **CUT** | parentheses: `roadmap (weekly? monthly?) is left as an exercise` |
| 1029, 1030 | "When a vendor's failure is ambiguous — and one of them always is — you need to already know" | APPOSITIVE | **CUT** | parentheses: `is ambiguous (and one of them always is) you need to already know` |
| 1066 | "Then be **lenient about it** — a log with no header is assumed current" | APPOSITIVE | **CUT** | colon: `**lenient about it**: a log with no header…` |
| 1101 | "not actually separate from your transport — which is the finding the exercise exists to surface" | APPOSITIVE (trailing *which*) | **CUT** | comma: `from your transport, which is the finding…` |
| 1111 | "The reason is in §2.6 — the order is what makes the chapter's prediction a real test" | APPOSITIVE | **CUT** | colon: `The reason is in §2.6: the order is what…` |
| 1154 | "**The parse side still outweighs the render side, 25 to 15** — it is just itemized now" | APPOSITIVE (label) | **CUT** | colon: `…25 to 15**: it is just itemized now.` |

### Chapter 2 em-dash KEEPs (SETUP, 27 occurrences across 26 lines)

Lines **91, 105, 111, 119, 127, 176, 201, 339, 472, 481, 498, 543, 560, 632,
665, 757, 828, 924, 932, 959, 975, 1050, 1082, 1132, 1175, 1183 (×2)**.

These are reversals and landing beats, not parentheticals. Representative:

- 91 — "It costs you a tail — and the tail is paid by whoever is using the product while you migrate."
- 176 — "it is a lie by omission — necessarily."
- 201 — "must be asked of the *end* of the request, or of the log — never of the whole."
- 472 — "cannot produce a valid Gemini 3.x request after a tool call at all — the API answers 400."
- 828 — "the chapter's answer to 'am I over-engineering?' — no, and here is the receipt."
- 1082 — "the map serializes differently on some future run, on some future machine, and never on the one where you tested."
- 1132 — "graded with a single API key — or none."

**Chapter 2 totals: 146 em-dashes → 52 guidance, 27 KEEP, 67 CUT → lands at 79.
Target ~50: undershoots by 29.** See the Summary for the two routes to 50.

---

## Application note

Apply each section top-to-bottom; line numbers are ascending and were verified
against the files on 2026-09-12 after the reads, not transcribed from memory.
Because several fixes delete text, **line numbers will drift as you go** — work
one section at a time from the bottom up if you prefer stable offsets, or
re-grep after each section.

Lines carrying **both** a TIC 1 and a TIC 2 verdict, where the two fixes
interact and should be applied together:

- Ch 1: 48–49, 115, 350–351
- Ch 2: 286, 313, 510, 632, 835, 886, 936, 951, 1030, 1050
