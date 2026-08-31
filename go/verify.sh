#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERIFY_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ctk-verify.XXXXXX")"
trap 'rm -rf "$VERIFY_DIR"' EXIT

UNFORMATTED="$({
    cd "$PROJECT_ROOT"
    git ls-files -z -- '*.go' | xargs -0 gofmt -l
})"
if [[ -n "$UNFORMATTED" ]]; then
    echo "Go files require formatting:" >&2
    echo "$UNFORMATTED" >&2
    exit 1
fi

go -C "$SCRIPT_DIR" test ./...
go -C "$SCRIPT_DIR" vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go -C "$SCRIPT_DIR" build ./...

(
    cd "$PROJECT_ROOT"
    git ls-files -z -- '*.sh' 'bash/bin/ctk' | xargs -0 bash -n
)

REVISION="$(git -C "$PROJECT_ROOT" rev-parse HEAD)"
go -C "$SCRIPT_DIR" run ./cmd/ctk-docbundle \
    -root "$PROJECT_ROOT" \
    -output "$VERIFY_DIR/documentation-bundle.zip" \
    -revision "$REVISION" >/dev/null

"$SCRIPT_DIR/third-party-notices.sh" "$VERIFY_DIR/THIRD_PARTY_NOTICES"
normalize_notices() {
    sed -E \
        's/^(Component: Go toolchain and standard library) \(go[0-9.]+\)$/\1 (go-version)/' \
        "$1"
}
if ! cmp -s \
    <(normalize_notices "$PROJECT_ROOT/THIRD_PARTY_NOTICES") \
    <(normalize_notices "$VERIFY_DIR/THIRD_PARTY_NOTICES"); then
    echo "THIRD_PARTY_NOTICES is not current; regenerate it with go/third-party-notices.sh" >&2
    diff -u \
        <(normalize_notices "$PROJECT_ROOT/THIRD_PARTY_NOTICES") \
        <(normalize_notices "$VERIFY_DIR/THIRD_PARTY_NOTICES") || true
    exit 1
fi

git -C "$PROJECT_ROOT" diff HEAD --check
