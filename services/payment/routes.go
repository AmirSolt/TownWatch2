package payment


func (payment *Payment) registerPaymentRoutes() {
	payment.paymentRoutes()

}

func (payment *Payment) paymentRoutes() {

	// payment.base.GET("/pricing", payment.auth.OptionalUserMiddleware, func(c *gin.Context) {

	// 	c.HTML(http.StatusOK, "pricing.tmpl", gin.H{
	// 		"data": pricingLoad{
	// 			pageLoad: pageLoad{
	// 				Title: "Pricing",
	// 			},
	// 			Tier1: models.TierT1,
	// 			Tier2: models.TierT2,
	// 		},
	// 	})

	// })

	// payment.base.GET("/checkout/:tier", payment.auth.RequireUserMiddleware, func(c *gin.Context) {
	// 	c.String(200, c.Param("tier"))
	// })

	// payment.base.GET("/wallet", payment.auth.RequireUserMiddleware, func(c *gin.Context) {

	// 	c.HTML(http.StatusOK, "wallet.tmpl", gin.H{
	// 		"data": walletLoad{
	// 			pageLoad: pageLoad{
	// 				Title: "Wallet",
	// 			},
	// 			TierDisplays: &tierDisplays,
	// 		},
	// 	})

	// })

}
