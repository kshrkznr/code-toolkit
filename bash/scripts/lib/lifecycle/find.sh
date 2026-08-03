#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/find.sh"

find_dist() {
    find_dir_thin "$(path_dist)"
}

find_archive() {
    find_dir_thin "$(path_archive)"
}

find_draft() {
    find_file_thin "$(path_draft)" "*.draft"
}

find_extension_lock() {
    stdin_or_args "$@" | while read -r dir; do
        find_file_thin "${dir}" "*.extensions.lock"
    done
}
