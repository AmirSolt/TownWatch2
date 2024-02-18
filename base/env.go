package base

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	HOME_URL     string `validate:"url"`
	IS_PROD      bool   `validate:"boolean"`
	DATABASE_URL string `validate:"url"`

	STRIPE_PRIVATE_KEY string `validate:"required"`
	STRIPE_WEBHOOK_KEY string `validate:"required"`

	ARCGIS_TORONTO_URL string `validate:"url"`

	EMAIL_CF_WORKER_URL     string `validate:"url"`
	EMAIL_CF_WORKER_API_KEY string `validate:"required"`
	EMAIL_SECRET_KEY        string `validate:"required"`

	JWE_SECRET_KEY string `validate:"required"`
}

func (base *Base) loadEnv() {

	if err := godotenv.Load(filepath.Join(base.RootDir, ".env")); err != nil {
		log.Fatal("Error .env:", err)
	}

	env := Env{
		HOME_URL:                os.Getenv("HOME_URL"),
		IS_PROD:                 strToBool(os.Getenv("IS_PROD")),
		DATABASE_URL:            os.Getenv("DATABASE_URL"),
		STRIPE_PRIVATE_KEY:      os.Getenv("STRIPE_PRIVATE_KEY"),
		STRIPE_WEBHOOK_KEY:      os.Getenv("STRIPE_WEBHOOK_KEY"),
		ARCGIS_TORONTO_URL:      os.Getenv("ARCGIS_TORONTO_URL"),
		EMAIL_CF_WORKER_URL:     os.Getenv("EMAIL_CF_WORKER_URL"),
		EMAIL_CF_WORKER_API_KEY: os.Getenv("EMAIL_CF_WORKER_API_KEY"),
		JWE_SECRET_KEY:          os.Getenv("JWE_SECRET_KEY"),
		EMAIL_SECRET_KEY:        os.Getenv("EMAIL_SECRET_KEY"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(env)
	if err != nil {
		log.Fatal("Error .env:", err)
	}

	base.Env = &env
}

func strToBool(s string) bool {
	return s == "true"
}
