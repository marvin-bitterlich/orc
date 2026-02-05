#!/bin/bash
# ORC Session Picker TUI - vertical layout

old_stty=$(stty -g)
cleanup() {
    stty "$old_stty"
    tput cnorm
    tput rmcup
}
trap cleanup EXIT INT TERM

stty -icanon -echo min 0 time 1
tput civis
tput smcup

LINES=$(tput lines)
COLS=$(tput cols)
LIST_HEIGHT=$(( (LINES - 3) / 2 ))
PREVIEW_START=$((LIST_HEIGHT + 2))
PREVIEW_HEIGHT=$((LINES - LIST_HEIGHT - 3))

C_RESET=$'\033[0m'
C_BOLD=$'\033[1m'
C_DIM=$'\033[2m'
C_CYAN=$'\033[36m'
C_GREEN=$'\033[32m'
C_RED=$'\033[38;5;167m'
C_MOSS=$'\033[38;5;65m'
C_REV=$'\033[7m'

declare -a TARGETS LABELS

build_list() {
    local idx=0
    while IFS= read -r session; do
        local ctx=$(tmux show-env -t "$session" ORC_CONTEXT 2>/dev/null | cut -d= -f2)
        TARGETS[idx]="${session}:"
        [[ -n "$ctx" ]] && LABELS[idx]="${C_CYAN}${C_BOLD}${session}${C_RESET} ${C_GREEN}${ctx}${C_RESET}" || LABELS[idx]="${C_CYAN}${C_BOLD}${session}${C_RESET}"
        ((idx++))
        
        while IFS=: read -r win_idx win_name; do
            local target="${session}:${win_idx}"
            local agent=$(tmux show-opt -t "$target" -wqv @orc_agent 2>/dev/null)
            local focus=$(tmux show-opt -t "$target" -wqv @orc_focus 2>/dev/null)
            TARGETS[idx]="$target"
            if [[ -n "$agent" ]]; then
                [[ "$agent" == GOBLIN@* ]] && LABELS[idx]="  └─ ${C_MOSS}${agent}${C_RESET}" || LABELS[idx]="  └─ ${C_RED}${agent}${C_RESET}"
                [[ -n "$focus" ]] && LABELS[idx]+=" ${C_DIM}→ ${focus}${C_RESET}"
            else
                LABELS[idx]="  └─ ${C_DIM}${win_name}${C_RESET}"
            fi
            ((idx++))
        done < <(tmux list-windows -t "$session" -F '#{window_index}:#{window_name}')
    done < <(tmux list-sessions -F '#{session_name}')
}

draw_frame() {
    printf '\033[2J\033[H'
    printf "\033[1;1H${C_BOLD}Sessions${C_RESET}"
    printf "\033[$((LIST_HEIGHT + 1));1H${C_DIM}"
    printf '─%.0s' $(seq 1 $((COLS - 1)))
    printf "${C_RESET}"
    printf "\033[$((PREVIEW_START));1H${C_BOLD}Preview${C_RESET}"
    printf "\033[${LINES};1H${C_DIM}jk Navigate  Enter Select  q Quit${C_RESET}"
}

draw_list() {
    local sel=$1 scr=$2
    for ((i=0; i<LIST_HEIGHT-1; i++)); do
        local idx=$((scr + i))
        printf "\033[$((i + 2));1H\033[K"
        ((idx < ${#LABELS[@]})) || continue
        [[ $idx -eq $sel ]] && printf "${C_REV}"
        printf "%b" "${LABELS[idx]}"
        [[ $idx -eq $sel ]] && printf "${C_RESET}"
    done
}

draw_preview() {
    local target=$1 line_num=0
    for ((i=0; i<PREVIEW_HEIGHT-1; i++)); do
        printf "\033[$((PREVIEW_START + 1 + i));1H\033[K"
    done
    line_num=0
    while IFS= read -r line && ((line_num < PREVIEW_HEIGHT-1)); do
        printf "\033[$((PREVIEW_START + 1 + line_num));1H%.${COLS}s" "$line"
        ((line_num++))
    done < <(tmux capture-pane -t "$target" -p 2>/dev/null)
}

# Main
build_list
total=${#TARGETS[@]}
sel=0 scr=0
max_visible=$((LIST_HEIGHT - 1))

draw_frame
draw_list $sel $scr
draw_preview "${TARGETS[$sel]}"

# Simple byte-by-byte read
getkey() {
    local c1 c2 c3
    IFS= read -r -n1 c1 || return 1
    if [[ "$c1" == $'\x1b' ]]; then
        IFS= read -r -t0.1 -n1 c2
        IFS= read -r -t0.1 -n1 c3
        case "$c2$c3" in
            '[A') echo UP; return ;;
            '[B') echo DOWN; return ;;
        esac
        echo ESC; return
    fi
    echo "$c1"
}

while true; do
    key=$(getkey) || continue
    
    case "$key" in
        UP|k) ((sel > 0)) && ((sel--)); ((sel < scr)) && scr=$sel ;;
        DOWN|j) ((sel < total-1)) && ((sel++)); ((sel >= scr+max_visible)) && ((scr++)) ;;
        ''|l) cleanup; tmux switch-client -t "${TARGETS[$sel]}"; exit 0 ;;
        q|ESC) exit 0 ;;
        *) continue ;;
    esac
    
    draw_list $sel $scr
    draw_preview "${TARGETS[$sel]}"
done
