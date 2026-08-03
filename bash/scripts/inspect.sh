#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "${SCRIPT_DIR}/lib/lifecycle/inspect.sh"

main() {
    local mode="${1:-}"
    local target="${2:-}"
    local target_right="${3:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
        view)
            view "$target"
            ;;
        sync)
            sync "$target" "$target_right"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

modelist() {
    echo view
    echo sync
}

view() {
    local source="${1:-}"
    inspect_prepare
    inspect_view "$source"
}

sync() {
    local left="${1:-}"
    local right="${2:-}"

    inspect_prepare
    inspect_sync "$left" "$right"
}

main "$@"
