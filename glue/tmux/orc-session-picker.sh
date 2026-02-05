#!/bin/bash
# ORC Session Picker - custom tmux session browser with ORC context
# Displays sessions with commission context and windows with agent identity
# Bound to prefix+S via ApplyGlobalBindings()

# Color palette for workbench names (hash to pick)
NAME_COLORS=(33 39 142 178 212 81 214 123)  # blue, cyan, purple, orange, pink, green, gold, teal

# Hash name to color index for consistent colors per workbench
name_to_color() {
    local name="$1"
    local hash=0
    for (( i=0; i<${#name}; i++ )); do
        hash=$(( (hash + $(printf '%d' "'${name:$i:1}")) % ${#NAME_COLORS[@]} ))
    done
    echo "${NAME_COLORS[$hash]}"
}

# Format agent string with colors
# GOBLIN@GATE-xxx: mossy green (colour65)
# IMP-{name}@BENCH-xxx: soft red prefix/suffix (colour167), unique name color
format_agent() {
    local agent="$1"
    local focus="$2"

    if [[ "$agent" == GOBLIN@* ]]; then
        # GOBLIN: mossy green
        echo "#[fg=colour65]${agent}#[default]"
    elif [[ "$agent" == IMP-*@* ]]; then
        # IMP-name@BENCH-xxx: red|color|red
        local name=$(echo "$agent" | sed 's/IMP-\(.*\)@.*/\1/')
        local bench=$(echo "$agent" | sed 's/.*\(@.*\)/\1/')
        local color=$(name_to_color "$name")
        local formatted="#[fg=colour167]IMP-#[fg=colour${color}]${name}#[fg=colour167]${bench}#[default]"

        if [[ -n "$focus" ]]; then
            formatted="${formatted} #[dim]→ ${focus}#[default]"
        fi
        echo "$formatted"
    else
        # Unknown agent type - show as-is
        echo "$agent"
    fi
}

# Build menu arguments
menu_args=(-T " ORC Sessions " -s "bg=colour236")

while IFS= read -r session; do
    # Get ORC session environment variables
    context=$(tmux show-environment -t "$session" ORC_CONTEXT 2>/dev/null | cut -d= -f2)

    # Session header: cyan bold name + green commission context
    if [[ -n "$context" ]]; then
        label="#[fg=colour39,bold]${session}#[default] #[fg=colour114]${context}#[default]"
    else
        label="#[fg=colour39,bold]${session}#[default]"
    fi
    menu_args+=("$label" "" "switch-client -t '$session'")

    # Windows under this session
    while IFS= read -r win_info; do
        win_idx=$(echo "$win_info" | cut -d: -f1)
        win_name=$(echo "$win_info" | cut -d: -f2)
        agent=$(tmux show-options -t "${session}:${win_idx}" -wqv @orc_agent 2>/dev/null)
        focus=$(tmux show-options -t "${session}:${win_idx}" -wqv @orc_focus 2>/dev/null)

        if [[ -n "$agent" ]]; then
            # Has ORC agent - format with colors
            formatted=$(format_agent "$agent" "$focus")
            win_label="  └─ ${formatted}"
        else
            # Non-ORC window - show dimmed, just window name
            win_label="  └─ #[dim]${win_name}#[default]"
        fi

        menu_args+=("$win_label" "" "switch-client -t '${session}:${win_idx}'")
    done < <(tmux list-windows -t "$session" -F '#{window_index}:#{window_name}')

    # Separator between sessions
    menu_args+=("" "" "")
done < <(tmux list-sessions -F '#{session_name}')

# Remove trailing separator
unset 'menu_args[-1]' 'menu_args[-1]' 'menu_args[-1]'

# Display the menu (arrow keys to navigate, enter to select)
tmux display-menu "${menu_args[@]}"
