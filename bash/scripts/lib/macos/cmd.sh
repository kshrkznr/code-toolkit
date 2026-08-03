#!/usr/bin/env bash

os_makelink(){
    local source=${1:-}
    local target=${2:-}
    source="$(cd "$source" && pwd)"
    ln -sfn "$source" "$target"
}

os_readlink(){
    readlink "${1:-}"
}

os_run_sh() {
    platform="${1:-code}"

    cat <<EOF
#!/usr/bin/env bash

BASE_DIR="\$(cd "\$(dirname "\$0")" && pwd)"
PLATFORM_BIN=\$(command -v "${platform}")

"\$PLATFORM_BIN" \\
  --user-data-dir "\$BASE_DIR/.data" \\
  --extensions-dir "\$BASE_DIR/.ext" \\
  "\$@"
EOF
}

os_exec_file() {
    local script_name="$1"
    local exec_file="${DIST_DIR}/$script_name"
    cat > "${exec_file}" <<'EOF'
#!/usr/bin/env bash
BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$BASE_DIR/run.sh" "$@"
EOF
    echo "${exec_file}"
}

os_pid_dist() {
    local platform="$1"
    local dist_name="$2"
    local app

    case "$platform" in
        code) app="Code" ;;
        kiro) app="Kiro" ;;
        *) return 0 ;;
    esac

    ps -axo pid,args |
        grep -- "--user-data-dir.*${dist_name}/.data" |
        grep -v "${app} Helper" |
        grep -v "grep" |
        awk '{print $1}' || true
}

os_pid_default() {
    local platform="${1:-code}"
    local executable helper

    case "$platform" in
        code)
            executable="Visual Studio Code.app/Contents/MacOS/Code"
            helper="Code Helper"
            ;;
        kiro)
            executable="Kiro.app/Contents/MacOS/Electron"
            helper="Kiro Helper"
            ;;
        *) return 0 ;;
    esac

    ps -axo pid,args |
        grep "${executable} " |
        grep -v "${helper}" |
        grep -v -- "--user-data-dir" |
        grep -v "grep" |
        awk '{print $1}' || true
}

os_kill() {
    kill "$1" 2>/dev/null
}

os_pid_exists() {
    kill -0 "$1" 2>/dev/null
}
