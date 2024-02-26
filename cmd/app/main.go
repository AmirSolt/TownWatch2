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
	defer base.Kill()
	auth := auth.LoadAuth(&base)
	payment := payment.LoadPayment(&base, auth)

	pages.RegisterPagesRoutes(&base, auth, payment)

	fmt.Println("=======")
	fmt.Println(base.DOMAIN)
	fmt.Println("=======")

	base.Run()
}
