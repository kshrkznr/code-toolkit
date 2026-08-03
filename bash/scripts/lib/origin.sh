#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/platform.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"

origin_dir() {
    local platform="${1:-}"
    echo "$(path_dist)/origin.${platform}"
}

origin_os() {
    case "$(uname -s)" in
        Darwin) echo macos ;;
        MINGW*|MSYS*) echo windows ;;
        *) return 1 ;;
    esac
}

origin_recipe_create() {
    local platform="$1"
    local platform_home="$2"
    local recipe_file="$3"
    local recipe_name="$4"

    {
        printf 'name: %s\n\n' "$recipe_name"
        printf 'os: %s\n\n' "$(origin_os)"
        printf 'platform: %s\n\n' "$platform"
        printf 'runtime:\n'
        printf '  - draft\n\n'
        printf 'profile:\n'
        platform_profiles "$platform_home" | sed 's/^/  - /'
        origin_profile_strategy "$platform_home"
    } > "$recipe_file"
}

origin_profile_strategy() {
    local platform_home="$1"
    local profile
    local found=false
    local entries=""

    while read -r profile
    do
        [[ "$(platform_profile_settings_strategy "$platform_home" "$profile")" == profile ]] || continue
        found=true
        entries+="    ${profile}:\n      settings: profile\n"
    done < <(platform_profiles "$platform_home")

    [[ "$found" == true ]] || return 0
    printf '\nconfig:\n  profile-strategy:\n%b' "$entries"
}

origin_create() {
    local platform="$1"
    local platform_home="$2"
    local dist_name
    local recipe_file

    dist_name="$(basename "$(origin_dir "$platform")")"
    recipe_file="$(mktemp "${TMPDIR:-/tmp}/ctk-origin-recipe.XXXXXX")" || return 1

    origin_recipe_create "$platform" "$platform_home" "$recipe_file" "$dist_name" || {
        rm -f "$recipe_file"
        return 1
    }
    dist_create "$dist_name" "$recipe_file" "$platform" || {
        rm -f "$recipe_file"
        return 1
    }
    rm -f "$recipe_file"
}
