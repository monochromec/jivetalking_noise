#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACTS_DIR="$ROOT_DIR/artifacts/silence_trim_demo"
mkdir -p "$ARTIFACTS_DIR"

LOGFILE=$(mktemp)
echo "Running integration test; log -> $LOGFILE"

cd "$ROOT_DIR"

# Run the integration test which prints input/output paths and durations
go test ./internal/processor -run TestSilenceTrimIntegration -v 2>&1 | tee "$LOGFILE"


# Prefer the stable marker file if the test wrote it
MARKER_FILE="/tmp/jivetalking_silence_trim_demo_paths.txt"
if [[ -f "$MARKER_FILE" ]]; then
  INPUT_PATH=$(sed -n '1p' "$MARKER_FILE") || true
  OUTPUT_PATH=$(sed -n '2p' "$MARKER_FILE") || true
else
  INPUT_PATH=$(grep -m1 '^input: ' "$LOGFILE" | sed 's/^input: //') || true
  OUTPUT_PATH=$(grep -m1 '^output: ' "$LOGFILE" | sed 's/^output: //') || true
fi

if [[ -z "$INPUT_PATH" || -z "$OUTPUT_PATH" ]]; then
  echo "Failed to locate input/output paths in test output; see $LOGFILE or $MARKER_FILE"
  exit 1
fi

cp -v "$INPUT_PATH" "$ARTIFACTS_DIR/"
cp -v "$OUTPUT_PATH" "$ARTIFACTS_DIR/"

echo "Artifacts copied to $ARTIFACTS_DIR"
echo "Log: $LOGFILE"

echo "Contents of $ARTIFACTS_DIR:"
ls -l "$ARTIFACTS_DIR"

echo "Test summary (grep lines):"
grep -E 'input: |output: |input_duration: |output_duration:' "$LOGFILE" || true

echo "Done."
