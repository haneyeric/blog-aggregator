package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/haneyeric/blog-aggregator/internal/config"
	"github.com/haneyeric/blog-aggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	cf, err := config.Read()
	if err != nil {
		fmt.Printf("Error getting config: %s", err)
		return
	}

	db, err := sql.Open("postgres", cf.DbUrl)
	if err != nil {
		fmt.Printf("Error opening db: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Printf("Error connecting to db: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	s := &state{cfg: &cf, db: dbQueries}

	commands := commands{
		cmds: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)

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
