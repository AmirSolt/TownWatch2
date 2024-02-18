package main

import (
	"fmt"
	"net/http"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/views/pages"
)

func main() {
	base := base.Base{
		RootDir: "./",
	}

	base.LoadBase()
	auth.LoadAuth(&base)

	pages.RegisterPagesRoutes()

	fmt.Println("=======")
	fmt.Println("http://localhost:8080")
	fmt.Println("=======")

	http.ListenAndServe(":8080", nil)
	base.Kill()
}
