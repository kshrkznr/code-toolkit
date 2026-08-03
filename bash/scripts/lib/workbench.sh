#!/usr/bin/env bash

workbench_editor() {
    local explicit="${1:-}"
    if [[ -n "$explicit" ]]
    then
        printf '%s\n' "$explicit"
    elif [[ -n "${EDITOR:-}" ]]
    then
        printf '%s\n' "$EDITOR"
    elif command -v code >/dev/null 2>&1
    then
        printf '%s\n' code
    else
        printf '%s\n' vim
    fi
}

workbench_open_path() {
	local target="$1"
	local editor
	editor="$(workbench_editor "${2:-}")"
	echo "[open] $editor : $target"
	if [[ "${3:-}" == "background" ]]
	then
		"$editor" "$target" &
	else
		"$editor" "$target"
	fi
}
