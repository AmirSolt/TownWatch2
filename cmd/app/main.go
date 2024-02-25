package main

import (
	"fmt"
	"townwatch/base"
	"townwatch/services/auth"
	"townwatch/services/payment"
	"townwatch/views/pages"
)

func main() {
	base := base.Base{
		RootDir: "./",
	}

	base.LoadBase()
	auth := auth.LoadAuth(&base)
	payment := payment.LoadPayment(&base, auth)

	pages.RegisterPagesRoutes(&base, auth, payment)

	fmt.Println("=======")
	fmt.Println("http://localhost:8080")
	fmt.Println("=======")

	base.Run()
	base.Kill()
}
