package base

// Webframework, handles mostly routes and requests

import (
	"github.com/gin-gonic/gin"
)

func (base *Base) loadEngine() {
	gin.SetMode(gin.DebugMode)
	// gin.DisableConsoleColor()
	engine := gin.Default()

	base.Engine = engine
}
