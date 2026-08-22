
// internal/config/config.go
package config

import (
  "log"
  "os"

  "github.com/joho/godotenv"
)

type Config struct {
  DBUser     string
  DBPassword string
  DBName     string
  DBHost     string
  DBPort     string
  HTTPPort   string
}

func Load() *Config {
  if err := godotenv.Load(); err != nil {
    log.Println("No .env file found, falling back to environment variables")
  }

  return &Config{
    DBUser:     os.Getenv("DB_USER"),
    DBPassword: os.Getenv("DB_PASSWORD"),
    DBName:     os.Getenv("DB_NAME"),
    DBHost:     os.Getenv("DB_HOST"),   // e.g. "localhost"
    DBPort:     os.Getenv("DB_PORT"),   // e.g. "5432"
    HTTPPort:   os.Getenv("HTTP_PORT"), // e.g. ":8080"
  }
}
