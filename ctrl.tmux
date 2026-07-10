#!/usr/bin/env bash

set -e

declare -r CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

get_option() { # <option> <default>
  local value
  if value="$(tmux show-option -gv "$1" 2>/dev/null)"; then
    printf '%s' "$value"
  else
    printf '%s' "$2"
  fi
}

declare -r option_command='@ctrl_command'
declare -r option_default_keybindings='@ctrl_default_keybindings'
declare -r option_session_prev_key='@ctrl_session_prev_key'
declare -r option_session_next_key='@ctrl_session_next_key'
declare -r option_pane_move_key='@ctrl_pane_move_key'
declare -r option_pane_extract_key='@ctrl_pane_extract_key'
declare -r option_pane_extract_command='@ctrl_pane_extract_command'

init_ctrl() {
  if [[ "$(get_option "${option_default_keybindings}" "on")" = "off" ]]; then
    exit 0
  fi

  local cmd
  cmd="$("${CURRENT_DIR}/scripts/ensure-command.sh" "$(get_option "${option_command}" "tmux-ctrl")")"

  local session_prev_key
  session_prev_key="$(get_option "${option_session_prev_key}" "(")"
  if test -n "${session_prev_key}"; then
    tmux bind-key "${session_prev_key}" run-shell -bE "${cmd} session prev"
  fi

  local session_next_key
  session_next_key="$(get_option "${option_session_next_key}" ")")"
  if test -n "${session_next_key}"; then
    tmux bind-key "${session_next_key}" run-shell -bE "${cmd} session next"
  fi

  local pane_move_key
  pane_move_key="$(get_option "${option_pane_move_key}" "M")"
  if test -n "${pane_move_key}"; then
    tmux bind-key "${pane_move_key}" run-shell -bE "${cmd} pane move"
  fi

  local pane_extract_key
  pane_extract_key="$(get_option "${option_pane_extract_key}" "Space")"
  if test -n "${pane_extract_key}"; then
    local pane_extract_command
    pane_extract_command="$(get_option "${option_pane_extract_command}" "tmux load-buffer -")"
    tmux bind-key "${pane_extract_key}" run-shell -b \
      "tmux display-popup -E -B -c #{client_name} -t #{pane_id} -x P -y P -w #{pane_width} -h #{pane_height} \"${cmd} pane extract --overlay -p #{pane_id} | ${pane_extract_command}\""
  fi
}

init_ctrl
