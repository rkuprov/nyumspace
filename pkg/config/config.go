package config

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	switch os.Getenv("NYUMSPACE_ENV") {
	case "local":
		err := loadFromLocal()
		if err != nil {
			return Cfg{}, err
		}
	default:
		log.Printf("Parsing settings from the %s environment\n", os.Getenv("NYUMSPACE_ENV"))
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

	return cfg, nil
}

func loadFromLocal() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parts := strings.Split(wd, string(filepath.Separator))
	ind := slices.Index(parts, "nyumspace")
	parts = parts[0 : ind+1]
	err = godotenv.Load(string(filepath.Separator) + filepath.Join(append(parts, "deployments", "env", "local.env")...))
	if err != nil {
		return err
	}

	return nil
}
