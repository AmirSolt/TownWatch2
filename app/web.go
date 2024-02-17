package app

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type BasicPageLoad struct {
	Title  string
	IsProd bool
}

func (app *App) loadEngine() {
	gin.SetMode(gin.DebugMode)
	// gin.DisableConsoleColor()
	engine := gin.Default()
	engine.LoadHTMLGlob(filepath.Join(app.RootDir, "domains/**/templates/*.tmpl"))

	app.Engine = engine
}
