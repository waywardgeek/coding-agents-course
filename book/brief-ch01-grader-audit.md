# Brief — Chapter 1 grader audit by deletion (P9)

You are the coder. I am the author. Read `book/course-policy.md` P9 first; this
job is the reason it exists. Read `book/review-ch02-code-v2.md` too, because it
is your own report from last round and this job generalizes its headline.

---

## Why this job exists

In Chapter 2 you proved that `ToolCallPart.Opaque` was graded by nothing:
deleting its only use scored **100/100**. That field is the one §2.6 prints as
the entire price of the seam bet. The book kept its promise in prose and broke
it in the grader.

Two mechanisms produced it, and neither was a mismatch between the code and the
spec:

1. **The fixture could not exercise the property** — the exhibit's tool call
   carried no opaque material, so there was nothing to fail to replay.
2. **The assertion would have passed vacuously** — the exhibit is
   Anthropic-authored, so a correct renderer withholds the material by design.

Chapter 1 is now the oldest ungraded-by-deletion artifact in the repo. It is
also, under P2, a maintained baseline that students rebase onto. If it has a
hole, every later chapter inherits a student who was told they were done.

---

## What Chapter 1 already has. Do not redo this

I checked before writing the brief, so you do not have to rediscover it:

- `internal/grade/checks.go` defines **7 checks**: `calls`, `growth`, `memory`,
  `protocol`, `replies`, `usage`, `wire`.
- `internal/grade/grader_test.go` has `TestMutationsAreCaught`: **10 mutants**,
  driven by a `COURSE_MUTATION` env var against a purpose-built mutant student,
  each asserting the **exact set** of failing check ids, plus a `none` control
  that must produce a clean pass.
- Every one of the 7 check ids is killed by at least one mutant.

That is already better than most graders ever get. It is not what P9 asks for.

---

## The gap this job actually fills

Chapter 1's suite proves the grader **detects ten specific breakages**. It does
not prove that **every behavior the reference solution implements is required**.
Those are different claims, and Chapter 2 is the proof that they come apart:
the hole there sat *inside a check that already had mutants*. Check-level
coverage is not property-level coverage.

The difference is in what gets mutated. Chapter 1 breaks a purpose-built mutant
student and asks which checks notice. The pass that found the Chapter 2 hole
deletes a behavior **from the reference solution** and asks whether the score
still says 100.

Run the second one.

---

## The job

**Step 1 — inventory the promises, from the chapter, not from the code.**
Work through `book/chapter-01-outline.md` and list every behavior the chapter
tells the student is required. For each, name the check id and the specific
assertion that is supposed to collect on it. Promises with no assertion are
findings before you run anything.

**Step 2 — delete each promise from the reference solution and score it.**
One property at a time, in a *copy* of `solutions/ch01`, neutralize exactly
that behavior and run the grader. The expected result is a score drop whose
failing set is exactly the checks that claim the property.

- Score unchanged: **that is the finding.** Report it the way you reported the
  Chapter 2 headline, with the patch and both scores.
- Score drops but the wrong check fires: a diagnosis bug. Report it.

**Step 3 — ask the two vacuity questions of every check.**
Can the fixture exercise this property at all? Does the assertion pass whether
or not the student did the work? Pay special attention to anything asserting
**absence**; absence is the easiest property in the world to satisfy by
accident, and P9 requires a negative control for each one.

**Step 4 — close what you found**, by strengthening existing checks, and add a
mutant for each newly graded property so the new grading is itself audited.

---

## Constraints. This is where I want you to hold the line

- **Do not re-divide the 100 points.** Adding a check or moving points is the
  author's call. Fold new assertions into the check whose *skill* they belong
  to, and keep failure messages distinct so a student still learns which
  property broke. If you believe a property genuinely deserves its own line,
  escalate the argument instead of acting on it. You did exactly this in
  Chapter 2 with `seam-render` and I ratified it; that was the right move and
  the reasoning is now written into §2.8.
- **Only require what the chapter already promises.** Strengthening the grader
  can fail a student who previously scored 100, which is acceptable only when
  the chapter asked for the thing all along. If closing a hole would require
  something Chapter 1 never told the student to build, stop: that is an outline
  change, and it is mine.
- **A property the chapter deliberately does not exercise is not a hole.** P9
  says so explicitly. Chapter 1 is the naive chapter. It has no seam, no event
  log, no provenance, no redaction. Do not grade forward into Chapter 2. If you
  cannot tell whether an omission is deliberate, that is an escalation, not a
  judgment call.
- **P3 still binds: grade behavior, never lineage.** Nothing you add may
  inspect source, infer authorship, or reward how the student wrote it.
- **Chapter 2 is out of scope.** Separate job, and it was just touched.
- **Keep the existing gates green**: reference 100/100, all 10 existing mutants
  detected with unchanged expectations. If an expectation now looks wrong, ask
  first whether the grader is right. Across two Chapter 2 rounds that question
  had six occasions to matter and the grader was correct on four of them.

---

## Method notes that earned their place last round

- The question is not "does the code implement the spec?" That answered yes in
  Chapter 2 and was wrong four times. The question is **"what would still pass
  if I deleted this?"**
- **Comments rot like graders do.** Two of your three Chapter 2 defects were
  comments: one asserting a gap the spec had closed, one asserting the exact
  opposite of §2.7. Read them as claims and check them.
- Patch a **copy**, never the reference in place.
- `cmd 2>&1 | head && echo OK` lies, because the pipeline's status is `head`'s.
  Use `; echo $?` or `${PIPESTATUS[0]}`.
- `ch02.log` and `internal/grade/ch02.log` are tracked grader artifacts that
  churn on every run. Check whether Chapter 1 has the same, and `git checkout
  --` them before committing.
- Never `git add -A`. Stage your own files by name; other sessions have WIP in
  this repo.

---

## Deliverables

1. **`book/review-ch01-grader-audit.md`**, in the shape of your last report:
   headline first if there is one, defects fixed, items escalated to the
   author, what changed, verification table.
2. **The per-property table**, which *is* the audit evidence: property, check
   id, score before deletion, score after, failing set. A row reading
   `100 → 100` is a finding. A reader should be able to scan that table and see
   what the grader actually requires.
3. **Code changes**: strengthened checks plus a mutant for each property whose
   grading is new.
4. **Gates**: `go build ./...`, `gofmt -l .`, `go vet ./...`, reference
   100/100, full mutation suite detected with exact-set assertions intact,
   points sum still exactly 100.

**If you find nothing, say so plainly and show the table that proves it.** A
clean audit with evidence is a real result and I will take it as one. Do not
manufacture a finding to justify the job — that failure mode is the same green
dashboard as the one we are hunting, just painted the other way round.
