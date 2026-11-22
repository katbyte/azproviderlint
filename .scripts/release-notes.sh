#!/usr/bin/env bash
#
# Prints the changelog entries for one release, which the release workflow
# publishes as that release's notes. Fails when the changelog has no entries
# for the tag, so a release never goes out with empty notes.
#
# Usage: ./.scripts/release-notes.sh v0.14.0
set -euo pipefail

tag="${1:?give the release tag, e.g. v0.14.0}"

# everything between the tag's "## vX.Y.Z (date)" heading and the next heading
notes="$(awk -v tag="$tag" '$1 == "##" { on = $2 == tag; next } on && (NF || seen) { seen = 1; print }' CHANGELOG.md)"

if [ -z "$notes" ]; then
  echo "CHANGELOG.md has no entries under a \"## $tag\" heading" >&2
  exit 1
fi

printf '%s\n' "$notes"
