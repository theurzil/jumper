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

    case "$1" in
        -h|--help)
            echo "j - jump to a tracked directory by frecency"
            echo ""
            echo "Usage:"
            echo "  j <term>   cd to the best-ranked directory matching <term>"
            echo "  j          cd to \$HOME"
            echo "  j --help   show this help"
            echo ""
            echo "History commands: jumper add|query|list|--help"
            return 0
            ;;
    esac

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

    _jumper_complete() {
        local term="${words[2]}"
        local -a matches
        matches=("${(@f)$(jumper complete "$term" 2>/dev/null)}")
        compadd -a matches
    }
    compdef _jumper_complete j
elif [ -n "$BASH_VERSION" ]; then
    case ";$PROMPT_COMMAND;" in
        *";_jumper_add;"*) ;;
        *) PROMPT_COMMAND="_jumper_add;${PROMPT_COMMAND}" ;;
    esac

    _jumper_complete() {
        local term="${COMP_WORDS[1]}"
        local IFS=$'\n'
        COMPREPLY=($(jumper complete "$term" 2>/dev/null))
    }
    complete -F _jumper_complete j
fi
