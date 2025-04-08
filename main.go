package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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

const (
	loginURL = "https://gmoyours.dt-r.com/customer/ajaxLogin.php"
)

var (
	config Config
	todo   []int64
	cookie string
)

func configLoad() error {
	log.Println("[configLoad]Begin")

	file, err := os.Open(configFilePath + "/" + configFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if _, err := decoder.Decode(&config); err != nil {
		return err
	}

	configPrint()
	log.Println("[configPrint]End")
	return nil
}

func configPrint() {
	fmt.Println("[account]")
	fmt.Println("login_id:", config.Account.LoginID)
	fmt.Println("password:", config.Account.Password)
	fmt.Println("customer_id:", config.Account.CustomerID)
	fmt.Println("customer_company_name:", config.Account.CustomerCompanyName)
	fmt.Println("customer_name:", config.Account.CustomerName)
	fmt.Println("customer_email:", config.Account.CustomerEmail)
	fmt.Println("[setups]")
	fmt.Println("service_id:", config.Setups.ServiceID)
	fmt.Println("service_menu_id:", config.Setups.ServiceMenuID)
	fmt.Println("next_monday1:", config.Setups.NextMonday1)
	fmt.Println("next_tuesday1:", config.Setups.NextTuesday1)
	fmt.Println("next_wednesday1:", config.Setups.NextWednesday1)
	fmt.Println("next_thursday1:", config.Setups.NextThursday1)
	fmt.Println("next_friday1:", config.Setups.NextFriday1)
	fmt.Println("next_monday2:", config.Setups.NextMonday2)
	fmt.Println("next_tuesday2:", config.Setups.NextTuesday2)
	fmt.Println("next_wednesday2:", config.Setups.NextWednesday2)
	fmt.Println("next_thursday2:", config.Setups.NextThursday2)
	fmt.Println("next_friday2:", config.Setups.NextFriday2)
}

func todoGenerate() {
	log.Println("[todoGenerate]Begin")

	nextMondayDate1 := getTargetWeekdayDate(time.Monday)
	nextTuesdayDate1 := getTargetWeekdayDate(time.Tuesday)
	nextWednesdayDate1 := getTargetWeekdayDate(time.Wednesday)
	nextThursdayDate1 := getTargetWeekdayDate(time.Thursday)
	nextFridayDate1 := getTargetWeekdayDate(time.Friday)

	nextMondayDate2 := nextMondayDate1.AddDate(0, 0, 7)
	nextTuesdayDate2 := nextTuesdayDate1.AddDate(0, 0, 7)
	nextWednesdayDate2 := nextWednesdayDate1.AddDate(0, 0, 7)
	nextThursdayDate2 := nextThursdayDate1.AddDate(0, 0, 7)
	nextFridayDate2 := nextFridayDate1.AddDate(0, 0, 7)

	todo = append(todo, generateUnixTime(nextMondayDate1, config.Setups.NextMonday1)...)
	todo = append(todo, generateUnixTime(nextTuesdayDate1, config.Setups.NextTuesday1)...)
	todo = append(todo, generateUnixTime(nextWednesdayDate1, config.Setups.NextWednesday1)...)
	todo = append(todo, generateUnixTime(nextThursdayDate1, config.Setups.NextThursday1)...)
	todo = append(todo, generateUnixTime(nextFridayDate1, config.Setups.NextFriday1)...)

	todo = append(todo, generateUnixTime(nextMondayDate2, config.Setups.NextMonday2)...)
	todo = append(todo, generateUnixTime(nextTuesdayDate2, config.Setups.NextTuesday2)...)
	todo = append(todo, generateUnixTime(nextWednesdayDate2, config.Setups.NextWednesday2)...)
	todo = append(todo, generateUnixTime(nextThursdayDate2, config.Setups.NextThursday2)...)
	todo = append(todo, generateUnixTime(nextFridayDate2, config.Setups.NextFriday2)...)

	if len(todo) == 0 {
		log.Fatalf("Null todo list.")
	} else {
		fmt.Println(todo)
		log.Println("[todoGenerate]End")
		return
	}
}

func getTargetWeekdayDate(target time.Weekday) time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	currentMonday := now.AddDate(0, 0, -(weekday - 1))
	nextMonday := currentMonday.AddDate(0, 0, 7)

	targetInt := int(target)
	if targetInt == 0 {
		targetInt = 7
	}

	offset := targetInt - 1
	return nextMonday.AddDate(0, 0, offset)
}

func generateUnixTime(date time.Time, val int) []int64 {
	var result []int64
	var hour, minute int
	switch val {
	case 1200:
		hour, minute = 12, 0
	case 1215:
		hour, minute = 12, 15
	case 1230:
		hour, minute = 12, 30
	case 1245:
		hour, minute = 12, 45
	case 1300:
		hour, minute = 13, 0
	case 1315:
		hour, minute = 13, 15
	default:
		return result
	}

	t := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location())
	fmt.Println("ToDo:", t)
	result = append(result, t.Unix())
	return result
}

func accountLogin() error {
	log.Println("[accountLogin]Begin")

	params := url.Values{}
	params.Add("action", "login")
	params.Add("login_id", config.Account.LoginID)
	params.Add("password", config.Account.Password)
	fullURL := loginURL + "?" + params.Encode()

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Thunder Client (https://www.thunderclient.com)")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error server access:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("error HTTP status (%s)", resp.Status)
		log.Println("Error HTTP status:", err)
		return err
	}

	cookie = resp.Header.Get("Set-Cookie")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return err
	}
	bodyStr := string(bodyBytes)

	if bodyStr == "" {
		err = fmt.Errorf("empty response body")
		log.Println("Unexpected response:", err)
		return err
	} else if bodyStr == "1" {
		fmt.Println("Login succeeded.")
		fmt.Println("Cookie:", cookie)
		log.Println("[accountLogin]End")
		return nil
	} else {
		err = fmt.Errorf("login failed")
		log.Println("Error:", err)
		return err
	}
}

func main() {
	err := configLoad()
	if err != nil {
		log.Fatalf("Bad config file: %v", err)
	}

	todoGenerate()

	err = accountLogin()
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
}
