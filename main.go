package main

import (
	"fmt"
	"net/http"
	"townwatch/app"
	"townwatch/domains/auth"
)

func main() {
	app := app.App{
		RootDir: "./",
	}

	app.LoadApp()
	auth.LoadAuth(&app)

	fmt.Println("=======")
	fmt.Println("http://localhost:8080")
	fmt.Println("=======")

	http.ListenAndServe(":8080", nil)
	app.Kill()
}
