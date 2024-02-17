package main

import (
	"fmt"
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
	fmt.Println("http://localhost:3000")
	fmt.Println("=======")

	app.Engine.Run()
	app.Kill()
}
