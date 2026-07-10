#!/usr/bin/env bash

# Prints the command the keybindings should invoke: the resolved name when
# present, otherwise the absolute path where the background install will land
# (so bindings work post-install without relying on $PATH).
#
# Usage: ensure-command.sh <cmd>

set -e

declare -r default_command="tmux-ctrl"

cmd="${1:-${default_command}}"

if command -v "${cmd}" >/dev/null 2>&1; then
  printf '%s' "${cmd}"
  exit 0
fi

if [[ "${cmd}" != "${default_command}" ]]; then
  printf '%s' "${cmd}"
  exit 0
fi

scripts_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
install_dir="${HOME}/.local/bin"
fallback="${install_dir}/${default_command}"

# Already installed into the fallback dir but that dir isn't on $PATH.
if [[ -x "${fallback}" ]]; then
  printf '%s' "${fallback}"
  exit 0
fi

log="${TMPDIR:-/tmp}/tmux-ctrl-install.log"
# Atomic lock dir: repeated plugin loads (e.g. config reload) must not stack
# concurrent installs.
lock="${TMPDIR:-/tmp}/tmux-ctrl-install.lock"
# reap an abandoned lock: treat a lock older than 10 minutes as stale.
if [[ -d "${lock}" ]] && find "${lock}" -maxdepth 0 -mmin +10 2>/dev/null | grep -q .; then
  rmdir "${lock}" 2>/dev/null || true
fi
# Best-effort dispatch: a failure here must not abort (set -e) before we print
# the fallback path below, or the keybindings would get an empty command.
if mkdir "${lock}" 2>/dev/null; then
  tmux run-shell -b "tmux display-message 'tmux-ctrl: installing…'; if INSTALL_DIR='${install_dir}' '${scripts_dir}/install.sh' >'${log}' 2>&1; then tmux display-message 'tmux-ctrl: installed'; else tmux display-message 'tmux-ctrl: install failed — see ${log}'; fi; rmdir '${lock}'" || rmdir "${lock}" 2>/dev/null || true
fi

printf '%s' "${fallback}"
