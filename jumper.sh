#!/usr/bin/env bash
# jumper shell integration. Source this from ~/.bashrc or ~/.zshrc:
#   source /path/to/jumper.sh

_jumper_add() {
    jumper add "$PWD" >/dev/null 2>&1
}

j() {
    if [ "$#" -eq 0 ]; then
        cd "$HOME" || return
        return
    fi

    local target
    target="$(jumper query "$*")"
    if [ -n "$target" ] && [ -d "$target" ]; then
        cd "$target" || return
    else
        echo "jumper: no match for '$*'" >&2
        return 1
    fi
}

if [ -n "$ZSH_VERSION" ]; then
    autoload -Uz add-zsh-hook
    add-zsh-hook chpwd _jumper_add
elif [ -n "$BASH_VERSION" ]; then
    PROMPT_COMMAND="_jumper_add;${PROMPT_COMMAND}"
fi
