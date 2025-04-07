package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Account struct {
	LoginID             string `toml:"login_id"`
	Password            string `toml:"password"`
	CustomerID          string `toml:"customer_id"`
	CustomerCompanyName string `toml:"customer_company_name"`
	CustomerName        string `toml:"customer_name"`
	CustomerEmail       string `toml:"customer_email"`
}

type Setups struct {
	ServiceID      string `toml:"service_id"`
	ServiceMenuID  string `toml:"service_menu_id"`
	NextMonday1    int    `toml:"next_monday1"`
	NextTuesday1   int    `toml:"next_tuesday1"`
	NextWednesday1 int    `toml:"next_wednesday1"`
	NextThursday1  int    `toml:"next_thursday1"`
	NextFriday1    int    `toml:"next_friday1"`
	NextMonday2    int    `toml:"next_monday2"`
	NextTuesday2   int    `toml:"next_tuesday2"`
	NextWednesday2 int    `toml:"next_wednesday2"`
	NextThursday2  int    `toml:"next_thursday2"`
	NextFriday2    int    `toml:"next_friday2"`
}

func configLoad() (Account, Setups, error) {
	var account Account
	var setups Setups

	file, err := os.Open("config.toml")
	if err != nil {
		return account, setups, err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if _, err := decoder.Decode(&account); err != nil {
		return account, setups, err
	}
	if _, err := decoder.Decode(&setups); err != nil {
		return account, setups, err
	}

	return account, setups, nil
}

func main() {
	account, setups, err := configLoad()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Account: %+v\n", account)
	fmt.Printf("Setups: %+v\n", setups)
}
