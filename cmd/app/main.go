package main

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/views/pages"
)

func main() {
	base := base.Base{
		RootDir: "./",
	}

	base.LoadBase()
	auth := auth.LoadAuth(&base)

	pages.RegisterPagesRoutes(&base, auth)

	fmt.Println("=======")
	fmt.Println("http://localhost:8080")
	fmt.Println("=======")

	base.Run()
	base.Kill()
}
