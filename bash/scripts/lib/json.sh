#!/usr/bin/env bash

json_merge() {
    jq -S -s 'reduce .[] as $item ({}; . * $item)'
}

json_read() {
    local file

    if [[ $# -gt 0 ]]; then
        for file in "$@"
        do
            if [[ ! -s "$file" ]]; then
                echo '{}'
            else
                json5 "$file"
            fi
        done
    else
        while read -r file
        do
            if [[ ! -s "$file" ]]; then
                echo '{}'
            else
                json5 "$file"
            fi
        done
    fi
}

json_normalize() {
    json5
}

json_gron(){
    gron
}

json_ungron(){
    gron -u
}

json_empty_to_json(){
    gron -u
}


json_gron_group() {
    local current=""
    local group

    while IFS= read -r line
    do
        group="$(json_gron_group_name "$line")"

        if [[ "$group" != "$current" ]]; then
            [[ -n "$current" ]] && echo
            echo "// ${group}"
            current="$group"
        fi

        echo "$line"
    done < "${1:-/dev/stdin}"
}
json_gron_group_name() {
    local line="$1"

    if [[ "$line" =~ ^json\[[\"\'](\[[^]]+\])[\"\']\] ]]; then
        printf '%s\n' "${BASH_REMATCH[1]}"
    elif [[ "$line" =~ ^json\[[\"\']([^.\"]+) ]]; then
        printf '%s\n' "${BASH_REMATCH[1]}"
    else
        printf 'other\n'
    fi
}

