#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/cookbook/extension.sh"
source "${SCRIPT_DIR}/lib/cookbook/settings.sh"
source "${SCRIPT_DIR}/lib/lifecycle/archive.sh"

resolve_load_recipe() {
    local name="$1"
    local _os="$2"
    local _from="$(recipe_file)"
    while read -r recipe_name
    do
        echo $recipe_name >&2
        [[ -z "$recipe_name" ]] && continue
        recipe_load "$recipe_name"
        [[ "$name" == $(recipe_name) ]] && [[ "$_os" == $(recipe_os) ]] && return

    done < <(recipe_list)
    recipe_load "$_from"
}

resolve_profile_extensions() {
    local profile="$1"
    local install_type="${2:-}"

    case "$install_type" in
    archive)
        local file="$(arc_extensions_lock_file "${profile}")"
        [[ -f "${file}" ]] || return 0
        cat "${file}" | resolve_vsix_extensions "${ARC_DIR}/vsix"
        ;;
    lock)
        local file="$(lock_extensions_lock_file "${profile}")"
        [[ -f "${file}" ]] || return 0
        lock_unlock_extension "${file}"
        ;;
    *)
        {
            :
            # future.capabillity mocule
            # recipe_runtimes | resolve_runtime_modules
            # resolve_profile_modules "$profile"
        }|
        sort -u |
        while read -r module
        do
            :
            # future.capabillity
            # resolve_recipe_content extensions_module_file "$module"
        done

        resolve_extensions_layer recipe_runtimes  extension_runtime_file
        resolve_recipe_content extension_profile_file "$profile"
        ;;
    esac
}

resolve_runtime_extensions() {
    local install_type="${1:-}"

    case "$install_type" in
    lock)
        local file="$(lock_runtime_extensions_lock_file)"
        [[ -f "${file}" ]] || return 0
        lock_unlock_extension "${file}"
        ;;
    *)
        resolve_extensions_layer recipe_runtimes extension_runtime_file
        ;;
    esac
}

resolve_extensions_layer() {
    local recipe_func="$1"
    local extensions_func="$2"

    "$recipe_func" |
    while read -r name
    do
        resolve_recipe_content "$extensions_func" "$name"
    done
}

resolve_vsix_extensions() {
    local vsix_dir="$1"

    while read -r extension
    do
        [[ -z "$extension" ]] && continue
        name="${extension%@*}"
        ver="${extension#*@}"
        vsix="${vsix_dir}/${name}-${ver}.vsix"
        if [[ ! -f "${vsix}" ]]; then
            continue
        fi
        echo "${vsix}"
    done
}

resolve_settings() {
    local install_type="${1:-}"

    case "$install_type" in
    archive)
        lock_settings_file
        ;;
    lock)
        lock_settings_file
        ;;
    *)
        resolve_default_settings
        ;;
    esac
}

resolve_default_settings() {
    local profile settings_default

    resolve_settings_base
    while read -r profile
    do
        settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
        [[ "$settings_default" == true ]] || continue
        resolve_profile_settings_content "$profile"
    done < <(recipe_profiles)
}

resolve_profile_settings() {
    local profile="$1"
    local install_type="${2:-}"

    case "$install_type" in
    archive)
        local archive_settings="$(arc_profile_settings_file "$profile")"
        [[ -f "$archive_settings" ]] || {
            echo "profile settings missing from Archive: $archive_settings" >&2
            return 1
        }
        echo "$archive_settings"
        ;;
    lock)
        local lock_settings="$(lock_profile_settings_file "$profile")"
        [[ -f "$lock_settings" ]] || {
            echo "profile settings missing from Lock: $lock_settings" >&2
            return 1
        }
        echo "$lock_settings"
        ;;
    *)
        resolve_default_settings
        resolve_profile_settings_content "$profile"
        ;;
    esac
}

resolve_settings_base() {
    resolve_settings_layer recipe_os         settings_os_file
    resolve_settings_layer recipe_platform   settings_platform_file
    resolve_runtime_extension_settings
    resolve_settings_layer recipe_runtimes   settings_runtime_file
}

resolve_runtime_extension_settings() {
    resolve_runtime_extensions \
        | extension_read recipe \
        | extension_merge \
        | while read -r extension
          do
              resolve_recipe_content settings_extension_file "$extension"
          done
}

resolve_profile_settings_content() {
    local profile="$1"

    resolve_profile_extensions "$profile" \
        | extension_read recipe \
        | extension_merge \
        | while read -r extension
          do
              resolve_recipe_content settings_extension_file "$extension"
          done
    resolve_recipe_content settings_profile_file "$profile"
}

resolve_settings_layer() {
    local recipe_func="$1"
    local settings_func="$2"

    "$recipe_func" |
    while read -r name
    do
        resolve_recipe_content "$settings_func" "$name"
    done
}

resolve_recipe_content() {
    local resolver="$1"
    local name="$2"
    local decorator="${3:-echo}"

    "$resolver" "$name" |
    while read -r file
    do
        "$decorator" "$file"
    done
}

stdin_or_args() {
    if (($#)); then
        printf '%s\n' "$@"
    else
        cat
    fi
}
