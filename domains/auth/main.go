package auth

import (
	"townwatch/app"
	"townwatch/domains/auth/authmodels"
)

type Auth struct {
	Queries *authmodels.Queries
	app     *app.App
}

func LoadAuth(app *app.App) *Auth {

	queries := authmodels.New(app.Conn)
	auth := Auth{
		Queries: queries,
		app:     app,
	}
	auth.registerAuthRoutes()

	return &auth
}
