#!/usr/bin/env bash

FAILED_EXTENSIONS=()

platform_create() {
    local dist_dir="$1"
    "${dist_dir}/run.sh" < /dev/null
    platform_close_dist "$(dist_name "${dist_dir}")"
    "${dist_dir}/run.sh" --list-extensions >/dev/null
}

platform_create_profile() {

    local dist_dir="$1"
    local profile="$2"

    if platform_profile_exists "${profile}" ; then
        echo "[profile exists] ${profile}"
        return
    fi
    echo "[profile create] ${profile}"
    "${dist_dir}/run.sh" --profile "${profile}" < /dev/null
    platform_close_dist "$(dist_name "${dist_dir}")"
    "${dist_dir}/run.sh" --list-extensions >/dev/null
}

platform_profile_exists() {
    local profile="$1"
    local storage_file="$(path_platform_strage "${dist_dir}/.data")"

    [[ ! -f "${storage_file}" ]] && return 1
    jq -e --arg profile "${profile}" \
        'any((.userDataProfiles // [])[]; .name == $profile)' \
        "${storage_file}" >/dev/null
}

platform_profiles() {
    local user_data_dir="${1:-}"
    local storage_file
    storage_file="$(path_platform_strage "$user_data_dir")"

    [[ -f "$storage_file" ]] || return 0
    jq -r '.userDataProfiles[]?.name // empty' "$storage_file" | tr -d '\r'
}

platform_profile_location() {
    local dist_dir="$1"
    local profile="$2"
    platform_profile_location_from_home "${dist_dir}/.data" "$profile"
}

platform_profile_location_from_home() {
    local user_data_dir="$1"
    local profile="$2"
    local storage_file
    storage_file="$(path_platform_strage "$user_data_dir")"

    jq -er --arg profile "${profile}" '
        (.userDataProfiles // [])[]
        | select(.name == $profile)
        | .location
    ' "${storage_file}"
}

platform_profile_directory() {
    local dist_dir="$1"
    local profile="$2"
    platform_profile_directory_from_home "${dist_dir}/.data" "$profile"
}

platform_profile_directory_from_home() {
    local user_data_dir="$1"
    local profile="$2"
    local location
    location="$(platform_profile_location_from_home "$user_data_dir" "$profile")"
    path_platform_profile "$user_data_dir" "$location"
}

platform_profile_settings_file() {
    local dist_dir="$1"
    local profile="$2"

    echo "$(platform_profile_directory "$dist_dir" "$profile")/settings.json"
}

platform_profile_settings_strategy() {
    local user_data_dir="$1"
    local profile="$2"
    local storage_file
    storage_file="$(path_platform_strage "$user_data_dir")"

    jq -e --arg profile "$profile" '
        any((.userDataProfiles // [])[];
            .name == $profile and (.useDefaultFlags.settings // false))
    ' "$storage_file" >/dev/null \
        && echo default \
        || echo profile
}

platform_profile_contents() {

    local dist_dir="$1"
    local profile="$2"
    local settings_default="${3:-true}"
    local keybindings_default="${4:-true}"
    local tasks_default="${5:-true}"
    local mcp_default="${6:-true}"
    local snippets_default="${7:-true}"

    echo "[profile content] ${profile}"
    echo "[profile ${profile} [settings_default:${settings_default}] [keybindings_default:${keybindings_default}] [tasks_default:${tasks_default}] [mcp_default:${mcp_default}] [snippets_default:${snippets_default}]"

    local storage_file="$(path_platform_strage "${dist_dir}/.data")"

    [[ ! -f "${storage_file}" ]] && return 0

    tmp_file="$(mktemp)"

    jq --arg profile "${profile}" \
        --argjson settings_default "${settings_default}" \
        --argjson keybindings_default "${keybindings_default}" \
        --argjson tasks_default "${tasks_default}" \
        --argjson mcp_default "${mcp_default}" \
        --argjson snippets_default "${snippets_default}" \
    '
        .userDataProfiles |= map(
        if .name == $profile then
            .useDefaultFlags = (
                (.useDefaultFlags // {})
                | if $settings_default     then .settings = true     else del(.settings)     end
                | if $keybindings_default  then .keybindings = true  else del(.keybindings)  end
                | if $tasks_default        then .tasks = true        else del(.tasks)        end
                | if $mcp_default          then .mcp = true          else del(.mcp)          end
                | if $snippets_default     then .snippets = true     else del(.snippets)     end
            )
            | if (.useDefaultFlags | length) == 0 then del(.useDefaultFlags) else . end
        else
            .
        end
        )
    ' "${storage_file}" > "${tmp_file}"

    mv "${tmp_file}" "${storage_file}"
}

platform_extension_recover() {
    local dist_dir="$1"

    rm -f "${dist_dir}/.ext/extensions.json"
}

platform_run() {
    local dist_dir="$1"
    local profile="$2"
    shift 2

    if [[ -n "$profile" ]]
    then
        "${dist_dir}/run.sh" --profile "$profile" "$@"
    else
        "${dist_dir}/run.sh" "$@"
    fi
}

platform_install_extensions() {
    local dist_dir="$1"
    local profile="$2"
    local update_extension="${3:-true}"
    local extension extension_id
    declare -A installed
    while IFS= read -r ext
    do
        installed["$ext"]=1
    done < <(platform_run "${dist_dir}" "${profile}" --list-extensions)

    while read -r extension
    do

        [[ -z "${extension}" ]] && continue

        extension_id="$extension"
        if [[ -f "$extension" ]]
        then
            extension_id="$(basename "$extension")"
            extension_id="${extension_id%-*}"
        fi

        [[ -n ${installed["$extension_id"]:-} ]] && \
            echo "[exist in $profile] ${extension_id}" && continue

        echo "[install $profile] ${extension}"
        if ! platform_run "${dist_dir}" "${profile}" \
            --install-extension "${extension}" \
            < /dev/null
        then
            echo "  [warn] failed: ${extension}"
            FAILED_EXTENSIONS+=(
                "${profile}:${extension}"
            )
            continue
        fi
    done
    if [[ "$update_extension" == true ]]
    then
        platform_run "${dist_dir}" "${profile}" --update-extensions 2>/dev/null
    fi
}

platform_uninstall_extensions() {
    local dist_dir="$1"
    local profile="$2"
    local update_extension="${3:-true}"

    declare -A expects
    while read -r ext
    do
        [[ -z "${ext}" ]] && continue
        [[ ! -f "${ext}" ]] && expects["${ext%@*}"]=1 && continue
        local ext_nm=$(basename ${ext})
        expects["${ext_nm%-*}"]=1
    done

    while IFS= read -r extension
    do
        [[ -z "${extension}" ]] && continue
        [[ -n ${expects["$extension"]:-} ]] && continue

        echo "[remove $profile] ${extension}"
        if ! platform_run "${dist_dir}" "${profile}" \
            --uninstall-extension "${extension}" \
            < /dev/null
        then
            echo "  [warn] failed: ${extension}"
            FAILED_EXTENSIONS+=(
                "${profile}:${extension}"
            )
            continue
        fi
    done < <(platform_run "${dist_dir}" "${profile}" --list-extensions)
    if [[ "$update_extension" == true ]]
    then
        platform_run "${dist_dir}" "${profile}" --update-extensions >/dev/null
    fi
}

platform_print_summary() {

    echo

    if [[ ${#FAILED_EXTENSIONS[@]} -eq 0 ]]
    then
        echo "[extensions] all installed"
        return
    fi

    echo "[extensions] failed"

    for ext in "${FAILED_EXTENSIONS[@]}"
    do
        echo "  - ${ext}"
    done
}

platform_vsix_repositories() {
    local platform="${1:-}"

    case "$platform" in
    kiro)
        printf '%s\n' open-vsx visual-studio-marketplace
        ;;
    *)
        printf '%s\n' visual-studio-marketplace
        ;;
    esac
}

platform_download_vsix() {

    local repository="$1"
    local out_dir="$2"
    local extension name ver pub pkg out url

    while read -r extension
    do

        [[ -z "${extension}" ]] && continue

        # split:
        # vscodevim.vim@1.32.4
        name="${extension%@*}"
        ver="${extension#*@}"

        # split publisher / extension
        pub="${name%%.*}"
        pkg="${name#*.}"

        # output filename
        out="${out_dir}/${name}-${ver}.vsix"

        # skip existing file
        if [[ -f "${out}" ]]; then
          echo "[skip] ${name}@${ver}"
          continue
        fi

        case "$repository" in
        visual-studio-marketplace)
            url="https://marketplace.visualstudio.com/_apis/public/gallery/publishers/${pub}/vsextensions/${pkg}/${ver}/vspackage"
            ;;
        open-vsx)
            url="https://open-vsx.org/api/${pub}/${pkg}/${ver}/file/${name}-${ver}.vsix"
            ;;
        *)
            echo "[warn] unsupported VSIX repository: ${repository}" >&2
            continue
            ;;
        esac

        echo "[download ${repository}] ${name}@${ver}"

        if ! curl \
            -L --fail --compressed -A "Mozilla/5.0" \
            -o "${out}" "${url}" \
            --ssl-no-revoke
        then
            rm -f "${out}"
            echo "[warn] VSIX unavailable: ${name}@${ver}"
        fi

    done
}

platform_pid_dist() {
    local dist_name="$1"
    local platform="${2:-}"

    if [[ -z "$platform" ]]
    then
        platform="$(platform_dist_platform "$dist_name")" || return 0
    fi

    os_pid_dist "$platform" "$dist_name"
}

platform_pid_default() {
    os_pid_default "${1:-}"
}

platform_dist_platform() {
    local dist_name="$1"
    local dist_dir recipe_file

    if [[ -d "$dist_name" ]]
    then
        dist_dir="$dist_name"
    else
        dist_dir="$(path_dist)/${dist_name}"
    fi

    recipe_file="${dist_dir}/.meta/recipe.yaml"
    [[ -f "$recipe_file" ]] || return 1
    yq -r '.platform // ""' "$recipe_file"
}

platform_close_dist() {
    local dist_name="$1"
    local dist_dir platform

    if [[ -d "$dist_name" ]]
    then
        dist_dir="$dist_name"
    else
        dist_dir="$(path_dist)/${dist_name}"
    fi

    platform="$(platform_dist_platform "$dist_dir")"
    [[ -n "$platform" && "$platform" != "null" ]] || {
        echo "dist platform missing: ${dist_dir}/.meta/recipe.yaml" >&2
        return 1
    }

    platform_pid_dist "$dist_name" "$platform" | platform_close_pid

    if dist_is_current "$dist_dir"
    then
        platform_pid_default "$platform" | platform_close_pid
    fi
}

platform_close_default() {
    local platform="${1:-}"
    platform_pid_default "$platform" | platform_close_pid
    platform_close_current "$platform"
}

platform_close_current() {
    local platform="${1:-}"
    local dist_name=$("${SCRIPT_DIR}/codevenv.sh" "current" "$platform")
    [[ "$dist_name" != "none" ]] || return 0
    platform_pid_dist "$dist_name" "$platform" | platform_close_pid
}

platform_close_pid() {
    while read -r pid
    do

        [[ -z "$pid" ]] && continue
        os_kill "$pid" || continue
        while os_pid_exists "$pid"
        do
            sleep 0.2
        done
    done
}
