#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/lib/os.sh"
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/platform.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/origin.sh"

main() {

    local mode="${1:-}"

    case "$mode" in
        modelist)
            modelist
            ;;
        activate)
            activate "${2:-}"
            echo "[done] ${mode}: ${2:-}"
            ;;
        deactivate)
            deactivate "${2:-}"
            echo "[done] ${mode}"
            ;;
        use)
            use_dist "${2:-}"
            ;;
        current)
            current "${2:-}"
            ;;
        list)
            dist_list
            ;;
        launch)
            shift
            launch "$@"
            ;;
        *)
            usage
            exit 0
            ;;
    esac
}

modelist() {
    echo activate
    echo deactivate
    echo use
    echo current
    echo list
    echo launch
}

current_link() {
    local platform="${1:-}"
    echo "$(path_dist)/current.${platform}"
}

platform_from_dist() {
    local recipe_file="${DIST_DIR}/.meta/recipe.yaml"
    local platform

    [[ -f "$recipe_file" ]] || {
        echo "dist recipe not found: $recipe_file" >&2
        return 1
    }

    platform="$(recipe_load "$recipe_file" >/dev/null; recipe_platform)"
    [[ -n "$platform" && "$platform" != "null" ]] || {
        echo "dist platform missing: $recipe_file" >&2
        return 1
    }
    printf '%s\n' "$platform"
}

active_platforms() {
    local link platform
    while IFS= read -r link
    do
        platform="${link##*/current.}"
        [[ -n "$platform" ]] && printf '%s\n' "$platform"
    done < <(find_link_thin "$(path_dist)" 'current.*' | sort)
}

select_active_platform() {
    local platforms
    platforms="$(active_platforms)"
    [[ -n "$platforms" ]] || return 1

    if [[ "$(printf '%s\n' "$platforms" | wc -l | tr -d ' ')" == "1" ]]
    then
        printf '%s\n' "$platforms"
    else
        printf '%s\n' "$platforms" | fzf
    fi
}

activate() {
    local platform="${1:-}"
    local current_link bak_dir platform_home ext_home code_user ext_dir

    [[ -n "$platform" ]] || {
        echo "usage: codevenv activate <platform>" >&2
        return 1
    }

    platform_home="$(path_platform_home "$platform")"
    ext_home="$(path_platform_extHome "$platform")"
    [[ -n "$platform_home" && -n "$ext_home" ]] || {
        echo "unsupported platform: $platform" >&2
        return 1
    }
    command -v "$platform" >/dev/null || {
        echo "platform command not found: $platform" >&2
        return 1
    }

    current_link="$(current_link "$platform")"
    bak_dir="$(origin_dir "$platform")"
    code_user="${platform_home}/User"
    ext_dir="${ext_home}/extensions"

    [[ -L "$code_user" ]] && return 0
    platform_close_default "$platform"

    if [[ ! -d "$bak_dir" ]]
    then
        origin_create "$platform" "$platform_home"
    fi

    if [[ -d "$code_user" ]]
    then
        rm -rf "$bak_dir/.data"
        mkdir -p "$bak_dir/.data"
        cp -R "$code_user" "$bak_dir/.data/User"
    fi
    if [[ -d "$ext_dir" ]]
    then
        rm -rf "$bak_dir/.ext"
        cp -R "$ext_dir" "$bak_dir/.ext"
    fi

    mkdir -p "$bak_dir/.data/User" "$bak_dir/.ext"
    "${SCRIPT_DIR}/lock.sh" lock "$(dist_name)"
    "${SCRIPT_DIR}/apply.sh" apply "${bak_dir}/.lock"

    rm -rf "$code_user" "$ext_dir"
    os_makelink "$bak_dir"                 "$current_link"
    os_makelink "$current_link/.data/User" "$code_user"
    os_makelink "$current_link/.ext"       "$ext_dir"
}

deactivate() {
    local platform="${1:-}"
    local current_link bak_dir platform_home ext_home code_user ext_dir

    if [[ -z "$platform" ]]
    then
        platform="$(select_active_platform)" || {
            echo "no active platform" >&2
            return 1
        }
    fi

    current_link="$(current_link "$platform")"
    [[ -L "$current_link" ]] || {
        echo "platform is not active: $platform" >&2
        return 1
    }

    platform_home="$(path_platform_home "$platform")"
    ext_home="$(path_platform_extHome "$platform")"
    bak_dir="$(origin_dir "$platform")"
    code_user="${platform_home}/User"
    ext_dir="${ext_home}/extensions"

    platform_close_default "$platform"

    [[ -L "$code_user" ]] && rm "$code_user"
    [[ -d "$code_user" ]] && rm -rf "$code_user"
    [[ -L "$ext_dir" ]] && rm "$ext_dir"
    [[ -d "$ext_dir" ]] && rm -rf "$ext_dir"

    [[ -L "$current_link" ]] && rm "$current_link"
    cp -r "$bak_dir/.data/User" "$platform_home"
    cp -r "$bak_dir/.ext" "$ext_dir"
}

use_dist() {
    local dist_name="${1:-}"
    local platform current_link
    dist_load "$dist_name"
    dist_name=$(dist_name)
    platform="$(platform_from_dist)"
    current_link="$(current_link "$platform")"

    [[ -L "$current_link" ]] || {
        echo "platform is not active: $platform (run: codevenv activate $platform)" >&2
        return 1
    }

    platform_close_default "$platform"
    platform_close_dist "$dist_name"
    os_makelink "$DIST_DIR" "$current_link"
    echo "[changed] ${platform}: $(dist_name)"
}

current() {
    local platform="${1:-}"
    local current_link

    if [[ -n "$platform" ]]
    then
        current_link="$(current_link "$platform")"
        if [[ ! -L "$current_link" ]]
        then
            echo "none"
            return 0
        fi
        basename "$(os_readlink "$current_link")"
        return 0
    fi

    local found=false
    while IFS= read -r platform
    do
        found=true
        printf '%s: %s\n' "$platform" "$(current "$platform")"
    done < <(active_platforms)
    [[ "$found" == true ]] || echo "none"
}

launch() {
    local dist_name="${1:-}"
    (($#)) && shift
    dist_load "$dist_name"
    exec "${DIST_DIR}/run.sh" "$@"
}

main "$@"
