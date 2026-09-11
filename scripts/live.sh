#!/usr/bin/env bash
#
# Run the Chapter 1 reference solution against the REAL Anthropic API.
#
#   scripts/live.sh              three scripted rounds, then the token bill
#   scripts/live.sh chat         interactive REPL — talk to it yourself
#   scripts/live.sh models       list the model IDs your key can actually use
#
# The API key is taken from, in order:
#   1. $ANTHROPIC_API_KEY
#   2. the file named by $ANTHROPIC_API_KEY_FILE
#   3. ~/.cr/settings.json (CodeRhapsody's config, if you happen to use it)
#
# The key is never echoed.

set -euo pipefail
cd "$(dirname "$0")/.."

if [ -z "${ANTHROPIC_API_KEY:-}" ] && [ -n "${ANTHROPIC_API_KEY_FILE:-}" ]; then
	ANTHROPIC_API_KEY="$(tr -d '[:space:]' < "$ANTHROPIC_API_KEY_FILE")"
fi

if [ -z "${ANTHROPIC_API_KEY:-}" ] && [ -f "$HOME/.cr/settings.json" ]; then
	ANTHROPIC_API_KEY="$(python3 -c "
import json, os
p = os.path.expanduser('~/.cr/settings.json')
print(json.load(open(p)).get('directClaudeAPIKey', ''))
")"
fi

if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
	echo "No API key found. Try:" >&2
	echo "    export ANTHROPIC_API_KEY=sk-ant-..." >&2
	exit 1
fi
export ANTHROPIC_API_KEY

# Leave ANTHROPIC_BASE_URL unset for api.anthropic.com, or point it at the
# course proxy. The program does not care which.
export ANTHROPIC_MODEL="${ANTHROPIC_MODEL:-claude-sonnet-5}"

case "${1:-rounds}" in
models)
	curl -s "https://${ANTHROPIC_BASE_URL:-api.anthropic.com}/v1/models?limit=40" \
		-H "x-api-key: $ANTHROPIC_API_KEY" \
		-H "anthropic-version: 2023-06-01" |
		python3 -c "
import json, sys
for m in json.load(sys.stdin).get('data', []):
    print(f\"{m['id']:<40} {m.get('display_name','')}\")
"
	;;

chat)
	echo "model: $ANTHROPIC_MODEL"
	exec go run ./solutions/ch01 chat
	;;

rounds)
	echo "model: $ANTHROPIC_MODEL"
	echo "three live rounds — round 3 asks it to recall what it said in round 1,"
	echo "which only works if the whole history is being resent."
	echo
	printf '%s\n' \
		'{"user":"Hello! I am starting a new project. Give it a codename and remember it."}' \
		'{"user":"What is the capital of France?"}' \
		'{"user":"What codename did you give my project?"}' |
		go run ./solutions/ch01
	;;

*)
	echo "usage: scripts/live.sh [rounds|chat|models]" >&2
	exit 2
	;;
esac
