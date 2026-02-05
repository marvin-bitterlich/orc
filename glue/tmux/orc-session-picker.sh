#!/bin/bash
# ORC Session Picker TUI - optimized version

tput civis
tput smcup
trap 'tput cnorm; tput rmcup; exit' EXIT INT TERM

LINES=$(tput lines)
COLS=$(tput cols)
LIST_WIDTH=$((COLS / 3))
PREVIEW_START=$((LIST_WIDTH + 3))
PREVIEW_WIDTH=$((COLS - LIST_WIDTH - 4))
MAX_VISIBLE=$((LINES - 4))

# Colors
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
    clear
    # Simple header/footer
    printf "\033[1;1H${C_BOLD}Sessions${C_RESET}%*s${C_BOLD}Preview${C_RESET}" $((LIST_WIDTH - 4)) ""
    printf "\033[$((LINES));1H${C_DIM}j/k Navigate  Enter Select  q/Esc Cancel${C_RESET}"
}

draw_list() {
    local sel=$1 scr=$2
    for ((i=0; i<MAX_VISIBLE; i++)); do
        local idx=$((scr + i))
        printf "\033[$((i + 2));1H\033[K"  # Move and clear line
        ((idx < ${#LABELS[@]})) || continue
        [[ $idx -eq $sel ]] && printf "${C_REV}"
        printf "%.${LIST_WIDTH}b" "${LABELS[idx]}"
        [[ $idx -eq $sel ]] && printf "${C_RESET}"
    done
}

draw_preview() {
    local target=$1 line_num=0
    while ((line_num < MAX_VISIBLE)); do
        printf "\033[$((line_num + 2));${PREVIEW_START}H\033[K"
        ((line_num++))
    done
    line_num=0
    while IFS= read -r line && ((line_num < MAX_VISIBLE)); do
        printf "\033[$((line_num + 2));${PREVIEW_START}H%.${PREVIEW_WIDTH}s" "$line"
        ((line_num++))
    done < <(tmux capture-pane -t "$target" -p 2>/dev/null)
}

# Main
build_list
total=${#TARGETS[@]}
sel=0 scr=0

draw_frame
draw_list $sel $scr
draw_preview "${TARGETS[$sel]}"

while IFS= read -rsn1 key; do
    case "$key" in
        $'\x1b') read -rsn2 -t 0.1 rest
            [[ "$rest" == "[A" ]] && key=k
            [[ "$rest" == "[B" ]] && key=j
            [[ -z "$rest" ]] && exit 0 ;;
    esac
    case "$key" in
        k) ((sel > 0)) && ((sel--)); ((sel < scr)) && scr=$sel ;;
        j) ((sel < total-1)) && ((sel++)); ((sel >= scr+MAX_VISIBLE)) && ((scr++)) ;;
        ''|l) tmux switch-client -t "${TARGETS[$sel]}"; exit 0 ;;
        q) exit 0 ;;
        *) continue ;;
    esac
    draw_list $sel $scr
    draw_preview "${TARGETS[$sel]}"
done
