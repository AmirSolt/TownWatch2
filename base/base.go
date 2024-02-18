package base

type Base struct {
	RootDir string
	*Env
	*DB
	*Config
}

func (base *Base) LoadBase() {

	base.loadEnv()
	base.loadDB()
	base.loadConfig()
}

func (base *Base) Kill() {
	base.killDB()
}
