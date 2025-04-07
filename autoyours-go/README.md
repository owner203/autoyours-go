# autoyours-go

## Overview
`autoyours-go` is a Go application designed to manage configurations for accounts and setups. It reads configuration data from a TOML file and provides structured access to this data within the application.

## Project Structure
```
autoyours-go
├── config
│   └── config.go
├── config.toml
├── go.mod
├── main.go
└── README.md
```

## Setup Instructions

1. **Clone the Repository**
   ```bash
   git clone <repository-url>
   cd autoyours-go
   ```

2. **Install Dependencies**
   Ensure you have Go installed on your machine. Run the following command to download the necessary dependencies:
   ```bash
   go mod tidy
   ```

3. **Create the Configuration File**
   Create a `config.toml` file in the root directory with the following structure:
   ```toml
   [account]
   login_id = "your_login_id"
   password = "your_password"
   customer_id = "your_customer_id"
   customer_company_name = "your_company_name"
   customer_name = "your_name"
   customer_email = "your_email"

   [setups]
   service_id = "your_service_id"
   service_menu_id = "your_service_menu_id"
   next_monday1 = 1
   next_tuesday1 = 2
   next_wednesday1 = 3
   next_thursday1 = 4
   next_friday1 = 5
   next_monday2 = 6
   next_tuesday2 = 7
   next_wednesday2 = 8
   next_thursday2 = 9
   next_friday2 = 10
   ```

## Usage
To run the application, execute the following command:
```bash
go run main.go
```

This will load the configuration from `config.toml` and perform the necessary operations as defined in `main.go`.

## Contributing
Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for details.