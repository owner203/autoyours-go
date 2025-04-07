package main

import (
	"fmt"
	"log"

	"autoyours-go/config"
)

func main() {
	account, setups, err := config.configLoad()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Account: %+v\n", account)
	fmt.Printf("Setups: %+v\n", setups)
}