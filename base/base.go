package base

import "github.com/gin-gonic/gin"

type Base struct {
	RootDir string
	*Env
	*DB
	*Config
	*gin.Engine
}

func (base *Base) LoadBase() {

	base.loadEnv()
	base.loadDB()
	base.loadConfig()
	base.loadEngine()
	base.loadLogging()
}

func (base *Base) Kill() {
	base.killDB()
	base.killLogging()
}
