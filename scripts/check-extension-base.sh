#!/usr/bin/env bash
set -euo pipefail

expected="${1:-v0.8.3}"
actual="$(git merge-base HEAD "$expected")"
tag_commit="$(git rev-list -n 1 "$expected")"

printf 'branch: %s\n' "$(git branch --show-current)"
printf 'expected base: %s (%s)\n' "$expected" "$tag_commit"
printf 'merge base: %s\n' "$actual"

if [[ "$actual" != "$tag_commit" ]]; then
  printf 'error: current branch is not based on %s\n' "$expected" >&2
  exit 1
fi

if git merge-base --is-ancestor "$tag_commit" HEAD; then
  printf 'ok: %s is an ancestor of HEAD\n' "$expected"
else
  printf 'error: %s is not an ancestor of HEAD\n' "$expected" >&2
  exit 1
fi
