package auth

import (
	"github.com/gin-gonic/gin"
)

func (auth *Auth) OptionalUserMiddleware(ginContext *gin.Context) {
	user, _ := auth.ValidateUser(ginContext)

	ginContext.Set("user", user)
	ginContext.Next()
}

func (auth *Auth) RequireUserMiddleware(ginContext *gin.Context) {
	user, err := auth.ValidateUser(ginContext)
	if err != nil {
		// fmt.Errorf(": %w", err)
		ginContext.Redirect(302, "/")
		return
	}

	ginContext.Set("user", user)
	ginContext.Next()
}
