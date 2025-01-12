package main

import (
	"fmt"
	"os"

	"github.com/haneyeric/blog-aggregator/internal/config"
)

func main() {
	cf, err := config.Read()
	if err != nil {
		fmt.Printf("Error getting config: %s", err)
		return
	}
	s := &state{cfg: &cf}
	commands := commands{
		cmds: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	cmd := command{name: args[1], args: args[2:]}

	err = commands.run(s, cmd)
	if err != nil {
		fmt.Printf("Error running command: %s\n", err)
		os.Exit(1)
	}
}
