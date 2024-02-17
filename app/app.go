package app

type App struct {
	RootDir string
	*Env
	*DB
	*Config
}

func (app *App) LoadApp() {

	app.loadEnv()
	app.loadDB()
	app.loadConfig()
}

func (app *App) Kill() {
	app.killDB()
}
