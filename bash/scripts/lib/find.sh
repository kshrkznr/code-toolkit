#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"

find_dist() {
    find_dir_thin "$(path_dist)"
}

find_archive() {
    find_dir_thin "$(path_archive)"
}

find_dir_thin() {
    local dir=${1:-}
    [[ ! -d "$dir" ]] && return
    find "${dir}" -mindepth 1 -maxdepth 1 -type d ! -path '*/.*' 2>/dev/null
}

find_extension_lock() {
    stdin_or_args "$@" | while read -r dir; do
        find_file_thin "${dir}" "*.extensions.lock"
    done
}

find_file_thin() {
    local dir=${1:-}
    local name=${2:-*}
    [[ ! -d "$dir" ]] && return
    find "${dir}" -name "$name" -mindepth 1 -maxdepth 1 -type f ! -name '.*'  2>/dev/null
}

find_link_thin() {
    local dir=${1:-}
    local name=${2:-*}
    [[ ! -d "$dir" ]] && return
    find "${dir}" -name "$name" -mindepth 1 -maxdepth 1 -type l 2>/dev/null
}

stdin_or_args() {
    if (($#)); then
        printf '%s\n' "$@"
    else
        cat
    fi
}
