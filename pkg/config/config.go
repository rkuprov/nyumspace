package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Cfg struct {
	PG         *Postgres   `json:"postgres"`
	HTTPServer *HTTPServer `json:"http_server"`
}
type Postgres struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"db_name"`
}
type HTTPServer struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func NewConfig() (Cfg, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Cfg{}, err
	}
	err = godotenv.Load(filepath.Join(wd, "..", "..", "deployments", "env", "local.env"))
	if err != nil {
		return Cfg{}, err
	}
	cfg := Cfg{
		PG: &Postgres{
			Host:     os.Getenv("PGHOST"),
			Port:     os.Getenv("PGPORT"),
			User:     os.Getenv("PGUSER"),
			Password: os.Getenv("PGPASSWORD"),
			DbName:   os.Getenv("PGDATABASE"),
		},
		HTTPServer: &HTTPServer{
			Host: os.Getenv("HTTP_HOST"),
			Port: os.Getenv("HTTP_PORT"),
		},
	}

	fmt.Println(cfg)

	return cfg, nil
}
