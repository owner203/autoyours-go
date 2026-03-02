package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

type Setup struct {
	ServiceID        string `toml:"service_id"`
	ServiceMenuID    string `toml:"service_menu_id"`
	CurrentMonday    int    `toml:"current_monday"`
	CurrentTuesday   int    `toml:"current_tuesday"`
	CurrentWednesday int    `toml:"current_wednesday"`
	CurrentThursday  int    `toml:"current_thursday"`
	CurrentFriday    int    `toml:"current_friday"`
	NextMonday1      int    `toml:"next_monday1"`
	NextTuesday1     int    `toml:"next_tuesday1"`
	NextWednesday1   int    `toml:"next_wednesday1"`
	NextThursday1    int    `toml:"next_thursday1"`
	NextFriday1      int    `toml:"next_friday1"`
	NextMonday2      int    `toml:"next_monday2"`
	NextTuesday2     int    `toml:"next_tuesday2"`
	NextWednesday2   int    `toml:"next_wednesday2"`
	NextThursday2    int    `toml:"next_thursday2"`
	NextFriday2      int    `toml:"next_friday2"`
}

type Config struct {
	Account Account `toml:"account"`
	Setup   Setup   `toml:"setups"`
}

type App struct {
	config Config
	todo   []int64
	cookie string
	client *http.Client
}

func newApp() *App {
	return &App{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

const (
	loginURL            = "https://gmoyours.dt-r.com/customer/ajaxLogin.php"
	bookingURL          = "https://gmoyours.dt-r.com/reservation/ajaxBooking.php"
	fetchBookingListURL = "https://gmoyours.dt-r.com/customer/reservation/ajaxViewList.php"
)

func retry(maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxRetries-1 {
			fmt.Println("Retrying...")
			time.Sleep(delay)
		}
	}
	return err
}

func (a *App) loadConfig(configPath string) error {
	log.Println("[loadConfig]Begin")

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if _, err := decoder.Decode(&a.config); err != nil {
		return err
	}

	a.printConfig()
	log.Println("[loadConfig]End")
	return nil
}

func (a *App) printConfig() {
	fmt.Println("[account]")
	fmt.Println("login_id:", a.config.Account.LoginID)
	fmt.Println("password:", "****")
	fmt.Println("customer_id:", a.config.Account.CustomerID)
	fmt.Println("customer_company_name:", a.config.Account.CustomerCompanyName)
	fmt.Println("customer_name:", a.config.Account.CustomerName)
	fmt.Println("customer_email:", a.config.Account.CustomerEmail)
	fmt.Println("[setups]")
	fmt.Printf("service_id: %s, service_menu_id: %s\n",
		a.config.Setup.ServiceID, a.config.Setup.ServiceMenuID)
	fmt.Printf("current:  Mon=%d Tue=%d Wed=%d Thu=%d Fri=%d\n",
		a.config.Setup.CurrentMonday, a.config.Setup.CurrentTuesday,
		a.config.Setup.CurrentWednesday, a.config.Setup.CurrentThursday,
		a.config.Setup.CurrentFriday)
	fmt.Printf("next_w1:  Mon=%d Tue=%d Wed=%d Thu=%d Fri=%d\n",
		a.config.Setup.NextMonday1, a.config.Setup.NextTuesday1,
		a.config.Setup.NextWednesday1, a.config.Setup.NextThursday1,
		a.config.Setup.NextFriday1)
	fmt.Printf("next_w2:  Mon=%d Tue=%d Wed=%d Thu=%d Fri=%d\n",
		a.config.Setup.NextMonday2, a.config.Setup.NextTuesday2,
		a.config.Setup.NextWednesday2, a.config.Setup.NextThursday2,
		a.config.Setup.NextFriday2)
}

func (a *App) generateTodo() {
	log.Println("[generateTodo]Begin")

	type entry struct {
		weekday    time.Weekday
		weekOffset int
		timeCode   int
	}

	s := a.config.Setup
	entries := []entry{
		{time.Monday, 0, s.CurrentMonday},
		{time.Tuesday, 0, s.CurrentTuesday},
		{time.Wednesday, 0, s.CurrentWednesday},
		{time.Thursday, 0, s.CurrentThursday},
		{time.Friday, 0, s.CurrentFriday},
		{time.Monday, 1, s.NextMonday1},
		{time.Tuesday, 1, s.NextTuesday1},
		{time.Wednesday, 1, s.NextWednesday1},
		{time.Thursday, 1, s.NextThursday1},
		{time.Friday, 1, s.NextFriday1},
		{time.Monday, 2, s.NextMonday2},
		{time.Tuesday, 2, s.NextTuesday2},
		{time.Wednesday, 2, s.NextWednesday2},
		{time.Thursday, 2, s.NextThursday2},
		{time.Friday, 2, s.NextFriday2},
	}

	for _, e := range entries {
		date := getTargetWeekdayDate(e.weekday).AddDate(0, 0, 7*e.weekOffset)
		if ts, ok := generateUnixTime(date, e.timeCode); ok {
			a.todo = append(a.todo, ts)
		}
	}

	if len(a.todo) == 0 {
		log.Fatalf("Null todo list.")
	}

	fmt.Println(a.todo)
	log.Println("[generateTodo]End")
}

func getTargetWeekdayDate(target time.Weekday) time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	currentMonday := now.AddDate(0, 0, -(weekday - 1))

	targetInt := int(target)
	if targetInt == 0 {
		targetInt = 7
	}

	offset := targetInt - 1
	return currentMonday.AddDate(0, 0, offset)
}

func generateUnixTime(date time.Time, val int) (int64, bool) {
	if val == 0 {
		return 0, false
	}

	hour := val / 100
	minute := val % 100

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		log.Printf("Invalid time code: %d", val)
		return 0, false
	}

	t := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location())
	fmt.Println("ToDo:", t)
	return t.Unix(), true
}

func (a *App) login() error {
	log.Println("[login]Begin")

	params := url.Values{}
	params.Add("action", "login")
	params.Add("login_id", a.config.Account.LoginID)
	params.Add("password", a.config.Account.Password)
	fullURL := loginURL + "?" + params.Encode()

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Thunder Client (https://www.thunderclient.com)")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("Error server access: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("error HTTP status (%s)", resp.Status)
		log.Printf("Error HTTP status: %v", err)
		return err
	}

	rawCookie := resp.Header.Get("Set-Cookie")
	if idx := strings.Index(rawCookie, ";"); idx != -1 {
		a.cookie = rawCookie[:idx]
	} else {
		a.cookie = rawCookie
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}
	bodyStr := string(bodyBytes)

	if bodyStr == "" {
		err = fmt.Errorf("empty response body")
		log.Printf("Unexpected response: %v", err)
		return err
	} else if bodyStr == "1" {
		fmt.Println("Login succeeded.")
		fmt.Println("Cookie:", a.cookie)
		log.Println("[login]End")
		return nil
	} else {
		err = fmt.Errorf("login failed")
		log.Printf("Error: %v", err)
		return err
	}
}

func (a *App) requestBooking(startUnixTime int64) error {
	endUnixTime := startUnixTime + 1800
	calendarID := fmt.Sprintf("%s.%s..%d.%d", a.config.Setup.ServiceID, a.config.Setup.ServiceMenuID, startUnixTime, endUnixTime)

	log.Printf("[requestBooking]Begin (%s)\n", calendarID)

	params := url.Values{}
	params.Add("action", "regist")
	params.Add("booking_data[calendar_id]", calendarID)
	params.Add("booking_data[service_id]", a.config.Setup.ServiceID)
	params.Add("booking_data[service_menu_id]", a.config.Setup.ServiceMenuID)
	params.Add("booking_data[start_unixtime]", fmt.Sprintf("%d", startUnixTime))
	params.Add("booking_data[end_unixtime]", fmt.Sprintf("%d", endUnixTime))
	params.Add("booking_data[num]", "1")
	params.Add("booking_data[customer_id]", a.config.Account.CustomerID)
	params.Add("booking_data[customer_company_name]", a.config.Account.CustomerCompanyName)
	params.Add("booking_data[customer_name]", a.config.Account.CustomerName)
	params.Add("booking_data[customer_email]", a.config.Account.CustomerEmail)
	params.Add("confirm", "1")

	fullURL := bookingURL + "?" + params.Encode()

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create booking request: %v", err)
		return err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Thunder Client (https://www.thunderclient.com)")
	req.Header.Set("Cookie", a.cookie)

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("Error server access: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("error HTTP status (%s)", resp.Status)
		log.Printf("Error HTTP status: %v", err)
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}

	bodyStr := string(bodyBytes)
	log.Printf("[requestBooking] Raw response for %s:\n%s", calendarID, bodyStr)
	fmt.Printf("  [Booking] %s\n", parseBookingResponse(bodyStr, calendarID))
	log.Printf("[requestBooking]End (%s)\n", calendarID)
	return nil
}

func parseBookingResponse(html string, calendarID string) string {
	var parts []string

	dateRe := regexp.MustCompile(`(\d{4}/\d{2}/\d{2})\s*（([月火水木金土日])）`)
	if m := dateRe.FindStringSubmatch(html); len(m) > 2 {
		parts = append(parts, fmt.Sprintf("%s(%s)", m[1], m[2]))
	}

	timeRe := regexp.MustCompile(`(\d{2}:\d{2})\s*～\s*(\d{2}:\d{2})`)
	if m := timeRe.FindStringSubmatch(html); len(m) > 2 {
		parts = append(parts, fmt.Sprintf("%s-%s", m[1], m[2]))
	}

	menuRe := regexp.MustCompile(`(セルリアン|フクラス)`)
	if m := menuRe.FindStringSubmatch(html); len(m) > 1 {
		parts = append(parts, m[1])
	}

	summary := strings.Join(parts, " ")
	if summary == "" {
		summary = calendarID
	}

	// Check for reservation ID (success indicator)
	idRe := regexp.MustCompile(`(?s)予約ID:\s*</dt>\s*<dd>\s*(R\d+)`)
	if m := idRe.FindStringSubmatch(html); len(m) > 1 {
		summary += " — " + m[1]
		return summary
	}

	// Check for error messages
	wsRe := regexp.MustCompile(`\s+`)
	errRe := regexp.MustCompile(`(?s)<p class="error_message">\s*(.*?)\s*</p>`)
	matches := errRe.FindAllStringSubmatch(html, -1)
	var errMsgs []string
	for _, m := range matches {
		msg := strings.TrimSpace(wsRe.ReplaceAllString(m[1], " "))
		if msg != "" && !strings.Contains(msg, "まだ手続は完了しておりません") {
			errMsgs = append(errMsgs, msg)
		}
	}

	if len(errMsgs) > 0 {
		summary += " — " + strings.Join(errMsgs, "; ")
	} else {
		summary += " — unknown response"
	}

	return summary
}

func parseBookingList(html string) {
	countRe := regexp.MustCompile(`(\d+)件の予約があります`)
	if m := countRe.FindStringSubmatch(html); len(m) > 1 {
		fmt.Printf("Booking List: %s reservations\n", m[1])
	} else {
		fmt.Println("Booking List: 0 reservations")
		return
	}

	idRe := regexp.MustCompile(`予約番号:\s*(R\d+)`)
	dateRe := regexp.MustCompile(`(\d{4}/\d{2}/\d{2})\s+\(([月火水木金土日])\)`)
	timeRe := regexp.MustCompile(`(\d{2}:\d{2})\s+-\s+(\d{2}:\d{2})`)
	menuRe := regexp.MustCompile(`(セルリアン|フクラス)`)

	items := strings.Split(html, `class="list_item"`)
	for i := 1; i < len(items); i++ {
		item := items[i]

		id := ""
		if m := idRe.FindStringSubmatch(item); len(m) > 1 {
			id = m[1]
		}
		date, day := "", ""
		if m := dateRe.FindStringSubmatch(item); len(m) > 2 {
			date = m[1]
			day = m[2]
		}
		startTime, endTime := "", ""
		if m := timeRe.FindStringSubmatch(item); len(m) > 2 {
			startTime = m[1]
			endTime = m[2]
		}
		menu := ""
		if m := menuRe.FindStringSubmatch(item); len(m) > 1 {
			menu = m[1]
		}

		fmt.Printf("  %s  %s(%s) %s-%s  %s\n", id, date, day, startTime, endTime, menu)
	}
}

func (a *App) fetchBookingList() error {
	log.Println("[fetchBookingList]Begin")

	req, err := http.NewRequest("GET", fetchBookingListURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Thunder Client (https://www.thunderclient.com)")
	req.Header.Set("Cookie", a.cookie)

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("Error server access: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("error HTTP status (%s)", resp.Status)
		log.Printf("Error HTTP status: %v", err)
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}

	parseBookingList(string(bodyBytes))
	log.Println("[fetchBookingList]End")
	return nil
}

func main() {
	configPath := flag.String("c", "./config.toml", "path to config file")
	flag.Parse()

	configDir := filepath.Dir(*configPath)
	logPath := filepath.Join(configDir, "autoyours-go.log")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	app := newApp()

	if err := app.loadConfig(*configPath); err != nil {
		log.Fatalf("Bad config file: %v", err)
	}

	app.generateTodo()

	if err := retry(5, 5*time.Second, app.login); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var bookingErrors []error

	for _, startUnixTime := range app.todo {
		wg.Add(1)
		go func(t int64) {
			defer wg.Done()
			if err := retry(5, 5*time.Second, func() error { return app.requestBooking(t) }); err != nil {
				mu.Lock()
				bookingErrors = append(bookingErrors, fmt.Errorf("booking for unix time %d failed: %v", t, err))
				mu.Unlock()
			}
		}(startUnixTime)
	}
	wg.Wait()

	if len(bookingErrors) > 0 {
		for _, e := range bookingErrors {
			log.Printf("ERROR: %v", e)
		}
		log.Printf("%d out of %d booking(s) failed", len(bookingErrors), len(app.todo))
	}

	if err := retry(5, 5*time.Second, app.fetchBookingList); err != nil {
		log.Fatalf("Booking list failed: %v", err)
	}
}
