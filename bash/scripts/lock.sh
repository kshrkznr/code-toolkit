#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"

main() {
    local mode="${1:-}"
    local dist_name="${2:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
        lock)
            lock "${dist_name}"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

modelist() {
    echo lock
}

lock() {
    local dist_name="${1:-}"
    local final_lock staging_lock backup_lock attempt
    dist_load "${dist_name}"

    final_lock="${DIST_DIR}/.lock"
    for attempt in 1 2 3
    do
        staging_lock="$(mktemp -d "${DIST_DIR}/.lock.staging.XXXXXX")"
        LOCK_DIR="$staging_lock"

        if lock_collect && lock_validate "$staging_lock"
        then
            backup_lock="${DIST_DIR}/.lock.previous.$$"
            rm -rf "$backup_lock"
            if [[ -d "$final_lock" ]] && ! mv "$final_lock" "$backup_lock"
            then
                rm -rf "$staging_lock"
                echo "failed to preserve existing Lock: $final_lock" >&2
                return 1
            fi
            if mv "$staging_lock" "$final_lock"
            then
                rm -rf "$backup_lock"
                LOCK_DIR="$final_lock"
                echo
                echo "[done] locked $dist_name"
                return
            fi

            [[ ! -d "$backup_lock" ]] || mv "$backup_lock" "$final_lock"
        fi

        rm -rf "$staging_lock"
        if [[ "$attempt" -lt 3 ]]
        then
            echo "[lock retry] ${attempt}/3 failed" >&2
            sleep "$attempt"
        fi
    done

    echo "lock failed after 3 attempts: $dist_name" >&2
    return 1
}

lock_collect() {
    lock_recipe || return 1
    recipe_file_set "$(dist_recipe)" >/dev/null || return 1
    lock_setting || return 1
    lock_profile_settings || return 1
    lock_extensions || return 1
}

main "$@"
