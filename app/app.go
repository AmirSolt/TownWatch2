package app

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	RootDir string
	*gin.Engine
	*Env
	*DB
	*Config
}

func (app *App) LoadApp() {

	gin.SetMode(gin.DebugMode)
	// gin.DisableConsoleColor()

	app.loadEnv()
	app.loadEngine()
	app.loadDB()
	app.loadConfig()
}

func (app *App) Kill() {
	app.killDB()
}
