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

type Config struct {
	Account Account `toml:"account"`
	Setups  Setups  `toml:"setups"`
}

const (
	configFilePath = "."
	configFileName = "config.toml"
)

var (
	config Config
)

func configLoad() error {
	log.Println("Loading configuration...")

	file, err := os.Open(configFilePath + "/" + configFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if _, err := decoder.Decode(&config); err != nil {
		return err
	}

	return nil
}

func main() {
	err := configLoad()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Config: %+v\n", config)
}
