#!/usr/bin/env bash

set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" || ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    echo "usage: go/release.sh <version>" >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
if [[ -n "$(git -C "$PROJECT_ROOT" status --porcelain --untracked-files=all)" ]]; then
    echo "release requires a clean checkout" >&2
    exit 1
fi
COMMIT="$(git -C "$PROJECT_ROOT" rev-parse HEAD)"
if ! TAG_COMMIT="$(git -C "$PROJECT_ROOT" rev-parse "$VERSION^{commit}" 2>/dev/null)"; then
    echo "release tag does not exist: $VERSION" >&2
    exit 1
fi
if [[ "$TAG_COMMIT" != "$COMMIT" ]]; then
    echo "release tag does not identify HEAD: $VERSION" >&2
    exit 1
fi

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

BUILD_DATE="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
LDFLAGS="-s -w -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Version=$VERSION -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Commit=$COMMIT -X github.com/kshrkznr/code-toolkit/go/internal/buildinfo.Date=$BUILD_DATE"
DOCBUNDLE_TOOL="$STAGING_DIR/ctk-docbundle"
DOCUMENTATION_BUNDLE="$STAGING_DIR/documentation-bundle.zip"
TAG_SOURCE="$STAGING_DIR/tag-source"

GOCACHE="${GOCACHE:-$STAGING_DIR/go-build-cache}" \
    go -C "$SCRIPT_DIR" build -trimpath -o "$DOCBUNDLE_TOOL" ./cmd/ctk-docbundle
"$DOCBUNDLE_TOOL" -root "$PROJECT_ROOT" -output "$DOCUMENTATION_BUNDLE" \
    -version "$VERSION" -revision "$COMMIT" -tag "$VERSION" >/dev/null

mkdir -p "$TAG_SOURCE"
git -C "$PROJECT_ROOT" archive --format=tar "$VERSION" | tar -xf - -C "$TAG_SOURCE"
"$DOCBUNDLE_TOOL" -root "$TAG_SOURCE" -output "$STAGING_DIR/tag-documentation-bundle.zip" \
    -version "$VERSION" -revision "$COMMIT" -tag "$VERSION" >/dev/null
if ! cmp -s "$DOCUMENTATION_BUNDLE" "$STAGING_DIR/tag-documentation-bundle.zip"; then
    echo "Documentation Bundle does not reproduce from tag: $VERSION" >&2
    exit 1
fi

DOCUMENTATION_MANIFEST="$STAGING_DIR/documentation-manifest.txt"
"$DOCBUNDLE_TOOL" -input "$DOCUMENTATION_BUNDLE" \
    -expect-version "$VERSION" -expect-revision "$COMMIT" -expect-tag "$VERSION" \
    > "$DOCUMENTATION_MANIFEST"
CONTENT_SHA256="$(awk '$1 == "content-sha256:" { print $2 }' "$DOCUMENTATION_MANIFEST")"
if [[ -z "$CONTENT_SHA256" ]]; then
    echo "Documentation Bundle content digest is missing" >&2
    exit 1
fi

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
    "$DOCBUNDLE_TOOL" -input "$DOCUMENTATION_BUNDLE" \
        -append-to "$platform_dir/ctk$extension" \
        -expect-version "$VERSION" -expect-revision "$COMMIT" \
        -expect-tag "$VERSION" -expect-content-sha256 "$CONTENT_SHA256" >/dev/null
    "$DOCBUNDLE_TOOL" -verify-executable "$platform_dir/ctk$extension" \
        -expect-version "$VERSION" -expect-revision "$COMMIT" \
        -expect-tag "$VERSION" -expect-content-sha256 "$CONTENT_SHA256" >/dev/null
}

build darwin arm64 ""
build darwin amd64 ""
build windows amd64 ".exe"

verify_native_cli() {
    local native_binary=""
    case "$(uname -s):$(uname -m)" in
        Darwin:arm64)
            native_binary="$STAGING_DIR/darwin_arm64/ctk"
            ;;
        Darwin:x86_64)
            native_binary="$STAGING_DIR/darwin_amd64/ctk"
            ;;
        MINGW*:x86_64|MSYS*:x86_64)
            native_binary="$STAGING_DIR/windows_amd64/ctk.exe"
            ;;
    esac
    if [[ -z "$native_binary" ]]; then
        echo "native packaged CLI verification skipped on $(uname -s) $(uname -m)" >&2
        return
    fi

    local verification_dir="$STAGING_DIR/native-verification"
    mkdir -p "$verification_dir"
    (
        cd "$verification_dir"
        env -u CTK_HOME "$native_binary" version | grep -F "$VERSION" >/dev/null
        env -u CTK_HOME "$native_binary" completion bash | grep -F "__start_ctk" >/dev/null
        env -u CTK_HOME "$native_binary" completion zsh | grep -F "compdef _ctk ctk" >/dev/null
        env -u CTK_HOME "$native_binary" completion fish | grep -F "complete -c ctk" >/dev/null
        env -u CTK_HOME "$native_binary" completion powershell | grep -F "Register-ArgumentCompleter" >/dev/null
        env -u CTK_HOME "$native_binary" init "$verification_dir/initialized" --exclude-sample >/dev/null
        test -d "$verification_dir/initialized/cookbook/recipe"
        test -d "$verification_dir/initialized/cookbook/ingredient"
        env -u CTK_HOME "$native_binary" docs >/dev/null
        env -u CTK_HOME "$native_binary" docs status | grep -F "source: packaged" >/dev/null
        env -u CTK_HOME "$native_binary" docs core >/dev/null
        env -u CTK_HOME "$native_binary" docs resolve "Settings Variant precedence" | grep -F "Knowledge.note.variant.md" >/dev/null
        env -u CTK_HOME "$native_binary" docs toc "Knowledge.core.cookbook.md" | grep -F "doc/core/core.cookbook.md#responsibility-1" >/dev/null
        env -u CTK_HOME "$native_binary" docs show "Knowledge.note.leaving-ctk.md#restore-an-activated-platform-first" >/dev/null
        env -u CTK_HOME "$native_binary" docs show "Knowledge.core.cookbook.md#responsibility-1" --depth -1..1 >/dev/null
        source_status="$(env -u CTK_HOME "$native_binary" docs --source "$PROJECT_ROOT" status)"
        grep -F "revision-match: match" <<<"$source_status" >/dev/null
        grep -F "definition-match: match" <<<"$source_status" >/dev/null
        grep -F "content-match: match" <<<"$source_status" >/dev/null
        grep -F "selected-path-dirty: clean" <<<"$source_status" >/dev/null
        env -u CTK_HOME "$native_binary" docs --source "$PROJECT_ROOT" show "Knowledge.core.md#responsibility" >/dev/null 2>/dev/null
        env -u CTK_HOME "$native_binary" docs export "$verification_dir/exported" >/dev/null
        test -f "$verification_dir/exported/$DOCUMENTATION_MANIFEST_PATH"
    )
}

DOCUMENTATION_MANIFEST_PATH=".ctk-docs/manifest.json"
verify_native_cli

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
