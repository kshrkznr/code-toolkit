#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/lifecycle/archive.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"
source "${SCRIPT_DIR}/lib/vsix.sh"

main() {
    local mode="${1:-}"
    local dist_name="${2:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
        archive)
            archive "${dist_name}"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

modelist() {
    echo archive
}

archive() {
    local dist_name="${1:-}"
    dist_load "$dist_name"
    [[ -z "$dist_name" ]] && dist_name=$(dist_name)
    arc_create "$dist_name" "$DIST_DIR"
    archive_download_vsix
    echo
    echo "[done] archived $dist_name"
}

archive_download_vsix() {
    local lock_dir="${ARC_DIR}/lock"
    local platform
    platform="$(recipe_load "$(dist_recipe)" >/dev/null; recipe_platform)"
    [[ -n "$platform" && "$platform" != "null" ]] || {
        echo "[error] archive platform missing" >&2
        return 1
    }

    echo "[resolve vsix from Pool] ${ARC_DIR}/vsix"
    vsix_pool_extensions "$lock_dir" | vsix_pool_download "$platform"
    vsix_pool_extensions "$lock_dir" | vsix_pool_copy "$platform" "${ARC_DIR}/vsix"
}

main "$@"
