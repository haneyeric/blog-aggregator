package main

import (
	"fmt"

	"github.com/haneyeric/blog-aggregator/internal/config"
)

func main() {
	cf, err := config.Read()
	if err != nil {
		fmt.Printf("Error getting config: %s", err)
		return
	}

	err = cf.SetUser("eric")
	if err != nil {
		fmt.Printf("Error setting user: %s", err)
		return
	}

	cf, err = config.Read()
	if err != nil {
		fmt.Printf("Error getting config: %s", err)
		return
	}

	fmt.Print(cf.DbUrl)

}
