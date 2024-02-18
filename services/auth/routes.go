package auth

func (auth *Auth) registerAuthRoutes() {
	auth.registerJoinRoute()
	// auth.joinVerify()

	if !auth.base.Env.IS_PROD {
		// auth.registerTestJoin()
	}
}

func (auth *Auth) registerJoinRoute() {

	// auth.base.Engine.GET("/join", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "join.tmpl", gin.H{
	// 		"data": joinLoad{
	// 			BasicPageLoad: base.BasicPageLoad{
	// 				Title: "Join",
	// 			},
	// 		},
	// 	})

	// })

	// http.Handle("/join", templ.Handler(authtemplates.JoinPage()))

	// auth.base.Engine.GET("/join", func(c *gin.Context) {

	// 	var buf bytes.Buffer
	// 	authtemplates.Hello("Amir").Render(c, &buf)

	// 	// err := templ.Render(&buf, authtemplates.Hello("Amir"))

	// 	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())

	// })
}

// func (auth *Auth) joinVerify() {

// 	auth.base.Engine.GET("/join/verify", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": verifyLoad{
// 				BasicPageLoad: base.BasicPageLoad{
// 					Title: "Join",
// 				},
// 			},
// 		})

// 	})
// }

// func (auth *Auth) registerTestJoin() {

// 	auth.base.Engine.POST("/join/test/singin", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": "",
// 		})

// 	})

// 	auth.base.Engine.POST("/join/test/singout", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": "",
// 		})

// 	})
// }
