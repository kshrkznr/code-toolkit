#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_PATH="${1:-$PROJECT_ROOT/THIRD_PARTY_NOTICES}"
NOTICE_TMP="$(mktemp "${TMPDIR:-/tmp}/ctk-third-party-notices.XXXXXX")"
LICENSE_CACHE="$(mktemp -d "${TMPDIR:-/tmp}/ctk-license-cache.XXXXXX")"
trap 'rm -f "$NOTICE_TMP"; rm -rf "$LICENSE_CACHE"' EXIT

append_file() {
    local label="$1"
    local path="$2"

    {
        printf '\nFile: %s\n\n' "$label"
        cat "$path"
        printf '\n'
    } >> "$NOTICE_TMP"
}

{
    cat <<'EOF'
THIRD-PARTY SOFTWARE NOTICES

CTK includes third-party software. The following notices are generated from
the Go standard library and non-standard modules linked into ./cmd/ctk.

These notices do not change the MIT License that applies to CTK itself.
EOF
} > "$NOTICE_TMP"

GO_ROOT="$(go env GOROOT)"
GO_LICENSE_ROOT=""
for candidate in "$GO_ROOT" "$(dirname "$GO_ROOT")"; do
    if [[ -f "$candidate/LICENSE" ]]; then
        GO_LICENSE_ROOT="$candidate"
        break
    fi
done
if [[ -z "$GO_LICENSE_ROOT" ]]; then
    echo "Go LICENSE not found from GOROOT: $GO_ROOT" >&2
    exit 1
fi

{
    printf '\n%s\n' '================================================================================'
    printf 'Component: Go toolchain and standard library (%s)\n' "$(go env GOVERSION)"
    printf 'Source: https://go.dev/\n'
} >> "$NOTICE_TMP"
append_file "LICENSE" "$GO_LICENSE_ROOT/LICENSE"
if [[ -f "$GO_ROOT/PATENTS" ]]; then
    append_file "PATENTS" "$GO_ROOT/PATENTS"
elif [[ -f "$GO_LICENSE_ROOT/PATENTS" ]]; then
    append_file "PATENTS" "$GO_LICENSE_ROOT/PATENTS"
fi

for target in darwin/arm64 darwin/amd64 windows/amd64; do
    target_os="${target%/*}"
    target_arch="${target#*/}"
    GOCACHE="$LICENSE_CACHE" GOOS="$target_os" GOARCH="$target_arch" CGO_ENABLED=0 \
        go -C "$SCRIPT_DIR" list -deps \
        -f '{{with .Module}}{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}{{end}}' \
        ./cmd/ctk
done | LC_ALL=C sort -u |
while IFS='|' read -r module_path module_version module_dir; do
    [[ -n "$module_path" ]] || continue

    {
        printf '\n%s\n' '================================================================================'
        printf 'Component: %s %s\n' "$module_path" "$module_version"
        printf 'Source: https://%s\n' "$module_path"
    } >> "$NOTICE_TMP"

    found=false
    while IFS= read -r license_path; do
        [[ -n "$license_path" ]] || continue
        found=true
        append_file "$(basename "$license_path")" "$license_path"
    done < <(find "$module_dir" -maxdepth 1 -type f \
        \( -iname 'LICENSE*' -o -iname 'COPYING*' -o -iname 'NOTICE*' \) \
        -print | LC_ALL=C sort)

    if [[ "$found" != true ]]; then
        echo "third-party license file not found: $module_path $module_version" >&2
        exit 1
    fi
done

mkdir -p "$(dirname "$OUTPUT_PATH")"
mv "$NOTICE_TMP" "$OUTPUT_PATH"
