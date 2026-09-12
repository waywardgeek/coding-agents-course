#!/usr/bin/env python3
"""Verify mid-turn hint delivery against the REAL Anthropic API.

    scripts/live_hint.py --model claude-sonnet-5   --carriage append
    scripts/live_hint.py --model claude-opus-4-8   --carriage system

What it does: asks the model to call agent_status several times in a row,
narrating between calls, so the turn is long enough to have a middle. Partway
through it types a hint at the running program. Then it reads the dumped event
log and reports, in order, every assistant message and which RequestSent
carried the hint.

That last part is the measurement. "The hint worked" is not interesting; what
is interesting is how many model responses happen between the hint being
carried and the model's behavior changing. Zero is the first-class path
working. One is the lag.

The key is read exactly as scripts/live.sh reads it, and is never echoed.
"""

import argparse
import json
import os
import pathlib
import queue
import re
import subprocess
import sys
import threading
import time

PIRATE = re.compile(r"\b(arr+|ahoy|matey|avast|ye|yer|aye|belay|scallywag|hearties|seas|cap'?n)\b", re.I)

PROMPT = (
    "Please call the agent_status tool, then tell me in ONE short sentence what it returned. "
    "Then call it again and report again. Do this {n} times in a row, one tool call and one "
    "sentence at a time. Do not call it more than once per message, and do not stop early."
)

HINT = "Please speak like a pirate for the rest of this turn."


def api_key() -> str:
    key = os.environ.get("ANTHROPIC_API_KEY")
    if key:
        return key.strip()
    settings = pathlib.Path.home() / ".cr" / "settings.json"
    if settings.exists():
        data = json.loads(settings.read_text())

        def find(node):
            if isinstance(node, dict):
                for k, v in node.items():
                    if isinstance(v, str) and v.startswith("sk-ant"):
                        return v
                    got = find(v)
                    if got:
                        return got
            elif isinstance(node, list):
                for v in node:
                    got = find(v)
                    if got:
                        return got
            return None

        got = find(data)
        if got:
            return got.strip()
    sys.exit("no API key found (set ANTHROPIC_API_KEY or ~/.cr/settings.json)")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", required=True)
    ap.add_argument("--calls", type=int, default=14)
    ap.add_argument("--carriage", default="append", choices=["append", "system"])
    ap.add_argument("--hint-delay", type=float, default=5.0,
                    help="seconds to wait before typing the hint (must land mid-turn)")
    ap.add_argument("--timeout", type=float, default=300.0)
    ap.add_argument("--slow-tool", type=float, default=0.0,
                    help="seconds agent_status should sleep, to test whether a slow tool delays the steer")
    args = ap.parse_args()

    root = pathlib.Path(__file__).resolve().parent.parent
    binary = pathlib.Path("/tmp/ch02live")
    srcdir = root / "solutions" / "ch02"
    if args.slow_tool > 0:
        # Patch a COPY. The reference solution is a printed listing and must
        # stay exactly as the book shows it.
        import shutil
        srcdir = root / "testdata" / "slowtool"
        shutil.rmtree(srcdir, ignore_errors=True)
        srcdir.mkdir(parents=True)
        for f in (root / "solutions" / "ch02").glob("*.go"):
            text = f.read_text()
            if f.name == "engine.go":
                text = text.replace('import (\n\t"bufio"', 'import (\n\t"bufio"\n\t"time"', 1)
                marker = "func (e *Engine) agentStatus() string {\n"
                assert marker in text
                text = text.replace(
                    marker, marker + "\ttime.Sleep(%d * time.Second)\n" % int(args.slow_tool), 1)
            (srcdir / f.name).write_text(text)
    subprocess.run(["go", "build", "-o", str(binary), "."], cwd=srcdir, check=True)

    outdir = pathlib.Path("/tmp/live-hint")
    outdir.mkdir(exist_ok=True)
    tag = f"{args.model}-{args.carriage}"
    logpath = outdir / f"{tag}.jsonl"

    env = dict(os.environ)
    env["ANTHROPIC_API_KEY"] = api_key()
    env["ANTHROPIC_MODEL"] = args.model
    env["HINT_CARRIAGE"] = args.carriage
    env.pop("ANTHROPIC_BASE_URL", None)
    env.pop("ANTHROPIC_API_URL", None)

    proc = subprocess.Popen([str(binary)], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, env=env, text=True, bufsize=1)

    lines: "queue.Queue[str]" = queue.Queue()

    def reader():
        for line in proc.stdout:
            lines.put(line.strip())
        lines.put("")

    threading.Thread(target=reader, daemon=True).start()

    def send(obj):
        proc.stdin.write(json.dumps(obj) + "\n")
        proc.stdin.flush()

    t0 = time.time()
    send({"user": PROMPT.format(n=args.calls)})
    time.sleep(args.hint_delay)
    hint_at = time.time() - t0
    send({"hint": HINT})
    print(f"[{tag}] hint typed {hint_at:.1f}s into the turn", flush=True)

    answered = False
    deadline = time.time() + args.timeout
    while time.time() < deadline:
        try:
            line = lines.get(timeout=1.0)
        except queue.Empty:
            continue
        if not line:
            break
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if "assistant" in obj:
            answered = True
            break

    send({"dump": str(logpath)})
    time.sleep(1.5)
    proc.stdin.close()
    try:
        proc.wait(timeout=20)
    except subprocess.TimeoutExpired:
        proc.kill()
    stderr = proc.stderr.read()

    print(f"[{tag}] turn completed: {answered}")
    if stderr.strip():
        print(f"[{tag}] stderr: {stderr.strip()[:600]}")
    if not logpath.exists():
        print(f"[{tag}] NO LOG DUMPED — nothing to analyse")
        return 1

    events = [json.loads(l) for l in logpath.read_text().splitlines() if l.strip()]
    hint_seq = next((e["seq"] for e in events
                     if e["type"] == "MessageReceived"
                     and HINT[:20] in json.dumps(e.get("parts", ""))), None)
    carried_seq = next((e["seq"] for e in events
                        if e["type"] == "RequestSent" and hint_seq and e["seq"] > hint_seq), None)

    print(f"[{tag}] hint recorded at seq {hint_seq}; first RequestSent after it: seq {carried_seq}")
    import datetime
    hev = next((e for e in events if e["seq"] == hint_seq), None)
    if hev and hev.get("time"):
        rec = datetime.datetime.fromisoformat(hev["time"].replace("Z", "+00:00")).timestamp()
        print(f"[{tag}] MAILBOX LATENCY: typed at t+{hint_at:.1f}s, recorded at t+{rec - t0:.1f}s "
              f"=> {rec - t0 - hint_at:+.1f}s. Near zero means the mailbox stayed live "
              f"while the tool ran; a delay near the tool duration means it did not.")
    print(f"[{tag}] {'seq':>5}  {'after?':>6}  {'pirate?':>7}  assistant text")
    n_after = 0
    first_pirate_after = None
    for e in events:
        if e["type"] != "AssistantMessage":
            continue
        text = "".join(p.get("text", "") for p in e.get("parts", []))
        after = carried_seq is not None and e["seq"] > carried_seq
        pirate = bool(PIRATE.search(text))
        if after:
            n_after += 1
            if pirate and first_pirate_after is None:
                first_pirate_after = n_after
        print(f"[{tag}] {e['seq']:>5}  {'yes' if after else 'no':>6}  "
              f"{'YES' if pirate else '-':>7}  {text[:96]}")

    print(f"[{tag}] RESULT: {n_after} assistant message(s) after the hint was carried; "
          f"first piratical one was #{first_pirate_after}"
          if first_pirate_after else
          f"[{tag}] RESULT: {n_after} assistant message(s) after the hint was carried; "
          f"NONE were piratical")
    return 0


if __name__ == "__main__":
    sys.exit(main())
