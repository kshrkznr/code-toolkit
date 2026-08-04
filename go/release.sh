#!/usr/bin/env bash

set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" || ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    echo "usage: go/release.sh <version>" >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_ROOT="${CTK_RELEASE_OUTPUT:-$PROJECT_ROOT/release}"
OUTPUT_DIR="$OUTPUT_ROOT/$VERSION"
if [[ -e "$OUTPUT_DIR" ]]; then
    echo "release output already exists: $OUTPUT_DIR" >&2
    exit 1
fi

STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ctk-release.XXXXXX")"
trap 'rm -rf "$STAGING_DIR"' EXIT
mkdir -p "$OUTPUT_DIR"

THIRD_PARTY_NOTICES="$STAGING_DIR/THIRD_PARTY_NOTICES"
"$SCRIPT_DIR/third-party-notices.sh" "$THIRD_PARTY_NOTICES"

COMMIT="$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_DATE="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
LDFLAGS="-s -w -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Version=$VERSION -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Commit=$COMMIT -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Date=$BUILD_DATE"

build() {
    local goos="$1"
    local goarch="$2"
    local extension="$3"
    local platform_dir="$STAGING_DIR/${goos}_${goarch}"
    mkdir -p "$platform_dir"
    cp "$PROJECT_ROOT/LICENSE" "$platform_dir/LICENSE"
    cp "$THIRD_PARTY_NOTICES" "$platform_dir/THIRD_PARTY_NOTICES"
    GOCACHE="${GOCACHE:-$STAGING_DIR/go-build-cache}" \
        GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
        go -C "$SCRIPT_DIR" build -trimpath -ldflags "$LDFLAGS" \
        -o "$platform_dir/ctk$extension" ./cmd/ctk
}

build darwin arm64 ""
build darwin amd64 ""
build windows amd64 ".exe"

tar -C "$STAGING_DIR/darwin_arm64" -czf "$OUTPUT_DIR/ctk_${VERSION}_darwin_arm64.tar.gz" ctk LICENSE THIRD_PARTY_NOTICES
tar -C "$STAGING_DIR/darwin_amd64" -czf "$OUTPUT_DIR/ctk_${VERSION}_darwin_amd64.tar.gz" ctk LICENSE THIRD_PARTY_NOTICES
(
    cd "$STAGING_DIR/windows_amd64"
    zip -q "$OUTPUT_DIR/ctk_${VERSION}_windows_amd64.zip" ctk.exe LICENSE THIRD_PARTY_NOTICES
)

if command -v sha256sum >/dev/null 2>&1; then
    (cd "$OUTPUT_DIR" && sha256sum ctk_*) > "$OUTPUT_DIR/checksums.txt"
else
    (cd "$OUTPUT_DIR" && shasum -a 256 ctk_*) > "$OUTPUT_DIR/checksums.txt"
fi

echo "$OUTPUT_DIR"
