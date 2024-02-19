package auth

import (
	"net/http"
)

// routes:
// 1. signin
// 2. signout
// 3. signintest

func (auth *Auth) registerAuthRoutes() {
	auth.authRoutes()

	if !auth.base.IS_PROD {
		auth.authTestRoutes()
	}
}

func (auth *Auth) authRoutes() {

	http.Handle("/join/signin", auth.RequireGuestMiddleware(auth.signinHandler()))
	http.Handle("/join/resendverif", auth.RequireGuestMiddleware(auth.resendVerifHandler()))
	http.Handle("/join/signout", auth.RequireUserMiddleware(auth.signoutHandler()))

}

func (auth *Auth) authTestRoutes() {

	http.Handle("/join-test/signin", auth.signinTestHandler())

}
