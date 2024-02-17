package auth

import (
	"bytes"
	"net/http"
	"townwatch/app"
	"townwatch/domains/auth/authtemplates"

	"github.com/gin-gonic/gin"
)

type joinLoad struct {
	app.BasicPageLoad
}

type verifyLoad struct {
	app.BasicPageLoad
}

func (auth *Auth) registerAuthRoutes() {
	auth.registerJoinRoute()
	// auth.joinVerify()

	if !auth.app.Env.IS_PROD {
		// auth.registerTestJoin()
	}
}

func (auth *Auth) registerJoinRoute() {

	// auth.app.Engine.GET("/join", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "join.tmpl", gin.H{
	// 		"data": joinLoad{
	// 			BasicPageLoad: app.BasicPageLoad{
	// 				Title: "Join",
	// 			},
	// 		},
	// 	})

	// })

	auth.app.Engine.GET("/join", func(c *gin.Context) {

		var buf bytes.Buffer
		authtemplates.Hello("Amir").Render(c, &buf)

		// err := templ.Render(&buf, authtemplates.Hello("Amir"))

		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())

	})
}

// func (auth *Auth) joinVerify() {

// 	auth.app.Engine.GET("/join/verify", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": verifyLoad{
// 				BasicPageLoad: app.BasicPageLoad{
// 					Title: "Join",
// 				},
// 			},
// 		})

// 	})
// }

// func (auth *Auth) registerTestJoin() {

// 	auth.app.Engine.POST("/join/test/singin", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": "",
// 		})

// 	})

// 	auth.app.Engine.POST("/join/test/singout", func(c *gin.Context) {

// 		c.HTML(http.StatusOK, "verify.tmpl", gin.H{
// 			"data": "",
// 		})

// 	})
// }
