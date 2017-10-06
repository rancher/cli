package cmd

import (
	"fmt"
	"github.com/urfave/cli"
)

func CompleteCommand() cli.Command {
	return cli.Command{
		Name:      "complete",
		Usage:     "Print auto complete bash functions",
		Action:    complete,
		ArgsUsage: "None",
	}
}

func complete(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 {
		if args[0] == "zsh" {
			v := `#! /bin/bash

autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit
: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _cli_bash_autocomplete $PROG

unset PROG
`
			fmt.Print(v)
		} else if args[0] == "bash" {
			v := `#! /bin/bash

: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _cli_bash_autocomplete $PROG

unset PROG`
			fmt.Print(v)
		}
	}
	return nil
}
