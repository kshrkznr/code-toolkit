#!/usr/bin/env bash

os_makelink(){
    local source=${1:-}
    local target=${2:-}
    [[ -L "$target" ]] && rm "$target"
    source="$(cd "$source" && pwd)"
    source=$(cygpath -w "${source}")
    target=$(cygpath -w "${target}")
    {
        cmd.exe //c mklink //J "${target}" "${source}"
    }> /dev/null
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
    local exec_file="${DIST_DIR}/$script_name.cmd"
    cat > "${exec_file}" <<'EOF'
@echo off
set BASE_DIR=%~dp0
bash "%BASE_DIR%run.sh" %*
EOF
    echo "${exec_file}"
}

os_pid_dist() {
    local platform="$1"
    local dist_name="$2"
    local process

    case "$platform" in
        code) process="Code.exe" ;;
        kiro) process="Kiro.exe" ;;
        *) return 0 ;;
    esac

    powershell.exe -NoProfile -Command "
        Get-CimInstance Win32_Process |
        Where-Object {
            \$_.Name -eq '${process}' -and
            \$_.CommandLine -match '--user-data-dir' -and
            \$_.CommandLine -match [regex]::Escape('${dist_name}') -and
            \$_.CommandLine -notmatch '--type='
        } |
        Select-Object -ExpandProperty ProcessId
    "
}

os_pid_default() {
    local platform="${1:-code}"
    local process

    case "$platform" in
        code) process="Code.exe" ;;
        kiro) process="Kiro.exe" ;;
        *) return 0 ;;
    esac

    powershell.exe -NoProfile -Command "
        Get-CimInstance Win32_Process |
        Where-Object {
            \$_.Name -eq '${process}' -and
            \$_.CommandLine -notmatch '--user-data-dir' -and
            \$_.CommandLine -notmatch '--type='
        } |
        Select-Object -ExpandProperty ProcessId
    "
}

os_kill() {
    local pid="$1"
    powershell.exe -NoProfile -Command "Stop-Process -Force -Id $pid"
}

os_pid_exists() {
    local pid="$1"

    powershell.exe -NoProfile -Command "(Get-Process -Id $pid -ErrorAction SilentlyContinue) -ne \$null" |
        grep -q True
}
