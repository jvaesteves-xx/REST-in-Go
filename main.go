package main

import "os"

func main() {
	app := App{}
	app.Initialize(
		os.Getenv("APP_DB_USERNAME_SANDBOX"),
		os.Getenv("APP_DB_PASSWORD_SANDBOX"),
		os.Getenv("APP_DB_NAME_SANDBOX"),
	)

	app.Run(":8080")
}
