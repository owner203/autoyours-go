# autoyours-go

Save your time from lunch reservation. A reimplementation of [autoyours](https://github.com/owner203/autoyours) in Go.

## Features

- Automates lunch reservations on the GMO Yours booking system
- Concurrent booking requests via goroutines
- Automatic retry (up to 5 attempts) for login and booking requests
- Booking confirmation via reservation list query
- Supports current week and next two weeks (Mon-Fri)

## Requirements

- Go 1.24+

## Build & Run

```bash
# Build
go build -o autoyours-go main.go

# Run
go run main.go
```

## Configuration

Copy `config.toml` to the same directory as the binary and fill in your details:

```toml
[account]
login_id = "your-email@example.com"
password = "PASSWORD"
customer_id = "C0012345"
customer_company_name = "Your Company"
customer_name = "Your Name"
customer_email = "your-email@example.com"

[setups]
service_id = "S001"          # Tokyo(S001)
service_menu_id = "S0000063" # Cerulean(S0000063) Fukulass(S0000064)
current_monday = 0           # Current Monday
current_tuesday = 0          # Current Tuesday
current_wednesday = 0        # Current Wednesday
current_thursday = 0         # Current Thursday
current_friday = 0           # Current Friday
next_monday1 = 1230          # Next Monday
next_tuesday1 = 1200         # Next Tuesday
next_wednesday1 = 1300       # Next Wednesday
next_thursday1 = 0           # Next Thursday
next_friday1 = 0             # Next Friday
next_monday2 = 0             # Monday in two weeks
next_tuesday2 = 1215         # Tuesday in two weeks
next_wednesday2 = 1315       # Wednesday in two weeks
next_thursday2 = 0           # Thursday in two weeks
next_friday2 = 0             # Friday in two weeks
```

### Time slot values

Set the time you want to book in HHMM format. Use `0` to skip a day.

| Value | Time  |
|-------|-------|
| 1200  | 12:00 |
| 1215  | 12:15 |
| 1230  | 12:30 |
| 1245  | 12:45 |
| 1300  | 13:00 |
| 1315  | 13:15 |

## License

[Apache License 2.0](LICENSE)
