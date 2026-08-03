#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/cookbook/recipe.sh"
source "${SCRIPT_DIR}/lib/lifecycle/find.sh"
source "${SCRIPT_DIR}/lib/lifecycle/dist.sh"
source "${SCRIPT_DIR}/lib/platform.sh"

LOCK_DIR=""

lock_load() {
    local lock_dir="${1:-}"
    [[ -d "$lock_dir" ]] || {
        echo "lock not found: $lock_dir" >&2
        return 1
    }
    LOCK_DIR="$(cd "$lock_dir" && pwd)"
}

lock_create() {
    local dist_dir="${1:-}"
    LOCK_DIR="${dist_dir}/.lock"
    mkdir -p "${LOCK_DIR}"
}

lock_recipe() {
    local src="${DIST_DIR}/.meta/recipe.yaml"
    local lock="${LOCK_DIR}/recipe.yaml"

    if [[ ! -f "${src}" ]]; then
      echo "recipe.yaml not found: ${src}"
      exit 1
    fi

    cp "${src}" "${lock}"
}

lock_setting() {
    local src="${DIST_DIR}/.data/User/settings.json"
    local lock="${LOCK_DIR}/settings.jsonc"

    if [[ -f "${src}" ]]
    then
        cp "${src}" "${lock}"
    else
        printf '{}\n' > "${lock}"
    fi
}

lock_profile_settings() {
    local profile settings_default

    while read -r profile
    do
        settings_default="$(recipe_profile_default_flag "$profile" settings)" || return 1
        if [[ "$settings_default" == true ]]
        then
            rm -f "$(lock_profile_settings_file "$profile")"
            continue
        fi
        lock_profile_setting "$profile" || return 1
    done < <(recipe_profiles)
}

lock_profile_setting() {
    local profile="$1"
    local src lock
    src="$(platform_profile_settings_file "${DIST_DIR}" "$profile")"
    lock="$(lock_profile_settings_file "$profile")"

    if [[ -f "$src" ]]
    then
        cp "$src" "$lock"
    else
        printf '{}\n' > "$lock"
    fi
}

lock_extensions() {
    recipe_file_set "$(dist_recipe)"
    lock_runtime_extensions || return 1

    local profile

    while read -r profile
    do
        local profile_lock="${LOCK_DIR}/${profile}.extensions.lock"
        lock_extension "${profile}" "${profile_lock}" || return 1
    done < <(recipe_profiles)
}

lock_runtime_extensions() {
    local output="${LOCK_DIR}/runtime.extensions.lock"

    echo "[runtime lock] extensions"
    "$DIST_DIR/run.sh" \
        --list-extensions \
        --show-versions \
    > "${output}"
}

lock_extension() {
    local profile="$1"
    local output="${LOCK_DIR}/${profile}.extensions.lock"
    echo "[profile lock] ${profile}"
    "$DIST_DIR/run.sh" \
        --profile "${profile}" \
        --list-extensions \
        --show-versions \
    > "${output}"
}

lock_recipe_file() {
    echo "${LOCK_DIR}/recipe.yaml"
}

lock_settings_file() {
    echo "${LOCK_DIR}/settings.jsonc"
}

lock_profile_settings_file() {
    local profile="$1"
    echo "${LOCK_DIR}/${profile}.settings.jsonc"
}

lock_extensions_lock_file() {
    local profile="$1"
    echo "${LOCK_DIR}/${profile}.extensions.lock"
}

lock_runtime_extensions_lock_file() {
    echo "${LOCK_DIR}/runtime.extensions.lock"
}

lock_ensure_lock() {
    local dist_name="$1"

    recipe_file_set "$(dist_recipe)" >/dev/null
    lock_after_mutation "$dist_name"
}

lock_after_mutation() {
    local dist_name="$1"
    local past_lock_dir="${2:-${DIST_DIR}/.lock}"
    local mode
    mode="$(recipe_lock_mode)"

    case "$mode" in
        refresh)
            lock_refresh "$dist_name"
            ;;
        reuse)
            lock_use_past "$past_lock_dir"
            ;;
        ask)
            lock_ask "$dist_name" "$past_lock_dir"
            ;;
        *)
            echo "invalid lock-mode: $mode (expected: refresh, reuse, ask)" >&2
            return 1
            ;;
    esac
}

lock_refresh() {
    local dist_name="$1"

    "${SCRIPT_DIR}/lock.sh" lock "$dist_name"
    lock_load "${DIST_DIR}/.lock"
}

lock_use_past() {
    local past_lock_dir="$1"

    lock_validate "$past_lock_dir" || return 1
    if [[ "$(cd "$past_lock_dir" && pwd)" != "${DIST_DIR}/.lock" ]]
    then
        rm -rf "${DIST_DIR}/.lock"
        cp -R "$past_lock_dir" "${DIST_DIR}/.lock"
    fi
    lock_load "${DIST_DIR}/.lock"
    echo "[lock] reusing trusted Lock: ${LOCK_DIR}"
}

lock_validate() {
    local lock_dir="$1"
    local previous_recipe="${RECIPE_FILE:-}"
    local profile settings_default required
    local valid=true

    for required in recipe.yaml settings.jsonc runtime.extensions.lock
    do
        if [[ ! -f "${lock_dir}/${required}" ]]
        then
            echo "incomplete Lock: ${lock_dir}/${required} not found" >&2
            valid=false
        fi
    done

    if [[ "$valid" == true ]]
    then
        recipe_file_set "${lock_dir}/recipe.yaml" >/dev/null || valid=false
    fi

    if [[ "$valid" == true ]]
    then
        while read -r profile
        do
            [[ -z "$profile" ]] && continue

            if [[ ! -f "${lock_dir}/${profile}.extensions.lock" ]]
            then
                echo "incomplete Lock: ${lock_dir}/${profile}.extensions.lock not found" >&2
                valid=false
            fi

            settings_default="$(recipe_profile_default_flag "$profile" settings)" || {
                valid=false
                continue
            }
            if [[ "$settings_default" != true \
                && ! -f "${lock_dir}/${profile}.settings.jsonc" ]]
            then
                echo "incomplete Lock: ${lock_dir}/${profile}.settings.jsonc not found" >&2
                valid=false
            fi
        done < <(recipe_profiles)
    fi

    RECIPE_FILE="$previous_recipe"
    [[ "$valid" == true ]] || {
        echo "trusted Lock is incomplete: $lock_dir" >&2
        return 1
    }
}

lock_ask() {
    local dist_name="$1"
    local past_lock_dir="$2"
    local answer

    [[ -t 0 && -t 1 ]] || {
        echo "lock-mode ask requires a TTY" >&2
        return 1
    }

    while true
    do
        read -r -p "[ask] Lock after ${dist_name}: [r]efresh / re[u]se / [n]o (fail): " answer
        case "$answer" in
            r|refresh)
                lock_refresh "$dist_name"
                return
                ;;
            u|reuse)
                lock_use_past "$past_lock_dir"
                return
                ;;
            n|no)
                echo "Lock update declined" >&2
                return 1
                ;;
            *)
                echo "select yes, no, or past" >&2
                ;;
        esac
    done
}

lock_pool_update() {
    vsix_pool_update "$(recipe_platform)" "$LOCK_DIR"
}

lock_resolve_extension_locks() {
    local lock_dir="$1"
     find_extension_lock "$lock_dir"
}

lock_extension_merge() {
    sort -u
}

lock_unlock_extension_files() {
    local lock_dir="$1"
    while read -r lock_file
    do
        lock_unlock_extension "${lock_file}" \
            > "${lock_file%.lock}"
    done < <(find_extension_lock "$lock_dir")
}

lock_unlock_extension() {
    sed 's/@.*//' "$1"
}
