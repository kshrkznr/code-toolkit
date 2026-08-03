#!/usr/bin/env bash
source "${SCRIPT_DIR}/lib/path.sh"
source "${SCRIPT_DIR}/lib/lifecycle/lock.sh"
source "${SCRIPT_DIR}/lib/platform.sh"

vsix_pool_dir() {
    local repository="$1"
    echo "$(path_vsix)/${repository}"
}

vsix_pool_repositories() {
    local platform="$1"
    platform_vsix_repositories "$platform"
}

vsix_pool_artifact() {
    local platform="$1"
    local extension="$2"
    local name="${extension%@*}"
    local version="${extension#*@}"
    local repository pool_dir artifact

    artifact="${name}-${version}.vsix"
    while IFS= read -r repository
    do
        pool_dir="$(vsix_pool_dir "$repository")"
        [[ -f "${pool_dir}/${artifact}" ]] || continue
        printf '%s\n' "${pool_dir}/${artifact}"
        return 0
    done < <(vsix_pool_repositories "$platform")

    return 1
}

vsix_pool_current_artifact() {
    local platform="$1"
    local extension="$2"
    local repository pool_dir artifact

    while IFS= read -r repository
    do
        pool_dir="$(vsix_pool_dir "$repository")"
        [[ -d "$pool_dir" ]] || continue
        artifact="$(find "$pool_dir" -maxdepth 1 -type f -name "${extension}-*.vsix" -print -quit 2>/dev/null)"
        [[ -n "$artifact" ]] || continue
        printf '%s\n' "$artifact"
        return 0
    done < <(vsix_pool_repositories "$platform")

    return 1
}

vsix_pool_resolve() {
    local platform="$1"
    local marketplace="${2:-true}"
    local extension artifact

    while IFS= read -r extension
    do
        [[ -z "$extension" ]] && continue
        [[ -f "$extension" ]] && {
            printf '%s\n' "$extension"
            continue
        }

        if artifact="$(vsix_pool_current_artifact "$platform" "$extension")"
        then
            echo "[pool resolve] ${extension}" >&2
            printf '%s\n' "$artifact"
            continue
        fi

        if [[ "$marketplace" == true ]]
        then
            printf '%s\n' "$extension"
            continue
        fi

        echo "[warn] extension unavailable from Pool: ${extension}" >&2
    done
}

vsix_pool_extensions() {
    local lock_dir="$1"

    lock_resolve_extension_locks "$lock_dir" |
    while IFS= read -r lock_file
    do
        cat "$lock_file"
    done |
    lock_extension_merge
}

vsix_pool_update() {
    local platform="$1"
    local lock_dir="$2"

    echo "[update vsix Pool] $(path_vsix)"
    vsix_pool_extensions "$lock_dir" | vsix_pool_download "$platform"
}

vsix_pool_download() {
    local platform="$1"
    local pool_dir

    local extension name version artifact repository tmp_dir stored=false
    while IFS= read -r extension
    do
        [[ -z "$extension" ]] && continue

        name="${extension%@*}"
        version="${extension#*@}"
        artifact="${name}-${version}.vsix"

        if vsix_pool_artifact "$platform" "$extension" >/dev/null
        then
            echo "[pool hit] ${extension}"
            continue
        fi

        stored=false
        while IFS= read -r repository
        do
            pool_dir="$(vsix_pool_dir "$repository")"
            mkdir -p "$pool_dir"
            tmp_dir="$(mktemp -d "${pool_dir}/.download.XXXXXX")"
            printf '%s\n' "$extension" | platform_download_vsix "$repository" "$tmp_dir"

            if [[ ! -f "${tmp_dir}/${artifact}" ]]
            then
                echo "[warn] VSIX download failed: ${extension} (${repository})" >&2
                rmdir "$tmp_dir"
                continue
            fi

            find "$pool_dir" -maxdepth 1 -type f -name "${name}-*.vsix" -delete
            mv "${tmp_dir}/${artifact}" "$pool_dir/"
            rmdir "$tmp_dir"
            echo "[pool store ${repository}] ${extension}"
            stored=true
            break
        done < <(vsix_pool_repositories "$platform")

        [[ "$stored" == true ]] || echo "[warn] VSIX unavailable from configured repositories: ${extension}" >&2
    done
}

vsix_pool_copy() {
    local platform="$1"
    local destination="$2"
    mkdir -p "$destination"

    local extension artifact missing=false
    while IFS= read -r extension
    do
        [[ -z "$extension" ]] && continue

        if ! artifact="$(vsix_pool_artifact "$platform" "$extension")"
        then
            echo "[error] VSIX missing from Pool: ${extension}" >&2
            missing=true
            continue
        fi

        cp "$artifact" "$destination/"
    done

    [[ "$missing" == false ]]
}
