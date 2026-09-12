# Course policy — standing rulings that span chapters

*Voice and tone are ruled separately, in `voice.md`. This file is about what the
course does; that one is about how it sounds.*

Chapter outlines state these to the *reader*. This file states them to *us*, in
one place, because a cross-chapter ruling that lives only inside one chapter's
outline is invisible when you sit down to write a later chapter's grader.

Every entry is a ruling by Bill. Do not quietly reverse one; amend it here with
a date and a reason, the way D5 in `review.md` was amended.

---

## P1. Write-once from Chapter 2

Chapter 1 is the single sacrificial chapter. From Chapter 2 on, every chapter is
strictly **additive** — new events, new tools, new seams; never "delete what you
built."

**Consequence:** chapter order is a dependency graph, not editorial taste.
Reordering chapters is a design change, not a formatting one.

## P2. A student may start any chapter from our reference solution

*Ruled by Bill, 2026-09-12: "they may use our solution as the starting point for
the next chapter, failing one chapter should not stop their progress."*

After each chapter's grading closes, the reference solution is published, and
the student may begin the next chapter from ours instead of their own.

**This is what makes P1 safe to promise.** Additive means the book never
demolishes code the student wrote. Rebasing means one bad structural choice in
Chapter 2 never strands them in Chapter 4. Without P2, "strictly additive"
stops being a promise to the reader and becomes a trap laid for them.

**Consequence — the expensive one.** The reference solution is no longer an
answer key, it is a **maintained baseline**. If it is a legal starting point
for every later chapter, then it must actually work as one:

- Changing a Chapter 2 data structure (say, shipping `ToolCallPart.Opaque`)
  obliges every later chapter's solution to move with it.
- `solutions/chNN` must compile and pass chapter NN's grader *and* be a viable
  input to chapter NN+1.
- Cheap to honour now. Expensive to discover at Chapter 6.

## P3. Graders grade behaviour, never lineage

No check may inspect the student's source, their naming, or their git history to
establish that the code "is theirs" or descends from their own earlier work. A
grader runs the binary and reads what it emits.

This falls out of P2 — a rebased student must be indistinguishable from one who
carried their own code forward — but it is worth stating separately because it
is easy to violate by accident, and because it is the thing that makes the
student's choice under P2 free of penalty.

Already-settled corollaries: normalize the *student's* own names (event types,
field names) case- and punctuation-insensitively; grade the **vendor's** wire
names exactly. `tool_use_id` is Anthropic's spelling, not a naming preference.

## P4. Three cost meters, never fused

Stated in full at `review.md` D5 and in Chapter 1 §1.7. Summary:

| meter | amount |
|---|---|
| reading the book | $0 |
| the graded exercises | $0 on the fake server; $20–$100 live for the **whole book** |
| building your own agent afterwards | $1,000–$10,000 of **Claude Code / Codex assistant spend**, optional, begins after the last chapter |

Do not delete the big numbers, do not attach them to the course or the toy
agent, and do not collapse the three back into one.

## P5. The exercises are sized for a directed assistant, not a typist

Chapter 2's reference solution is ~1,600 lines. Hand-typed that is a semester
project; directed, it is a week. State this as a *time* cost, not a money cost,
and keep the read-only path explicitly legitimate: an experienced engineer who
runs no exercises still gets most of the value.

*Rejected 2026-09-12:* making this the book's subtitle. It describes the
reader's workflow rather than the book's content, "AI-accelerated" on a cover
in 2026 reads as a confession about how the book was produced, and it would
have implied a money cost where the real one is time.
