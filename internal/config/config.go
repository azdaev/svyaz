package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken     string
	DatabasePath string
	Host         string
	Port         string
	CSRFSecret   string
	CookieDomain string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	c := &Config{
		BotToken:     os.Getenv("BOT_TOKEN"),
		DatabasePath: os.Getenv("DATABASE_PATH"),
		Host:         os.Getenv("HOST"),
		Port:         os.Getenv("PORT"),
		CSRFSecret:   os.Getenv("CSRF_SECRET"),
		CookieDomain: os.Getenv("COOKIE_DOMAIN"),
	}

	if c.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	if c.DatabasePath == "" {
		c.DatabasePath = "./svyaz.db"
	}
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == "" {
		c.Port = "3000"
	}
	if c.CSRFSecret == "" {
		return nil, fmt.Errorf("CSRF_SECRET is required")
	}

	return c, nil
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}
