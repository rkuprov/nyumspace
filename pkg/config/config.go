package config

import (
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type Cfg struct {
	PG *Postgres `json:"postgres"`
}
type Postgres struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"db_name"`
}

func NewConfig() (Cfg, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Cfg{}, err
	}
	err = godotenv.Load(filepath.Join(wd, "..", "deployments", "env", "local.env"))
	if err != nil {
		return Cfg{}, err
	}
	cfg := Cfg{
		PG: &Postgres{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DbName:   os.Getenv("POSTGRES_DB"),
		},
	}

	return cfg, nil
}
