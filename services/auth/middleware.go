package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (auth *Auth) RequireGuestMiddleware(ctx *gin.Context) {
	user, _ := auth.ValidateUser(ctx)
	if user != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	ctx.Next()
}

func (auth *Auth) OptionalUserMiddleware(ctx *gin.Context) {
	user, err := auth.ValidateUser(ctx)
	if err != nil {
		fmt.Printf("\nOptionalUserMiddleware Error:%w \n\n", err)
	}
	ctx.Set("user", user)
	ctx.Next()
}

func (auth *Auth) RequireUserMiddleware(ctx *gin.Context) {

	user, err := auth.ValidateUser(ctx)
	if err != nil || user == nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	ctx.Set("user", user)
	ctx.Next()

}
