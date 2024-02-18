package auth

import (
	"townwatch/base"
	"townwatch/services/auth/authmodels"
)

type Auth struct {
	Queries *authmodels.Queries
	base    *base.Base
}

func LoadAuth(base *base.Base) *Auth {

	queries := authmodels.New(base.Conn)
	auth := Auth{
		Queries: queries,
		base:    base,
	}
	auth.registerAuthRoutes()

	return &auth
}
