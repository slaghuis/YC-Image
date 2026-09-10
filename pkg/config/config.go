package config

import (
  "log"
  "os"
  "time"
)

// Configurations exported
type Configurations struct {
	Server       ServerConfigurations
  Security     JwtConfigurations
	Database     DatabaseConfigurations
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Port string
}

type JwtConfigurations struct {
  JWTIssuer       string
  JWTSigningKey   string // HS256 secret
}

// DatabaseConfigurations exported
type DatabaseConfigurations struct {
	DBHost		 string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func must(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatalf("missing required env %s", key)
    }
    return v
}

func LoadConfig() Configurations {
    return Configurations {
        Server : ServerConfigurations{
        Port:              getenv("PORT", "8082"),
      },
      Security :    JwtConfigurations {
        JWTIssuer:         getenv("JWT_ISSUER", "identity"),
        JWTSigningKey:     must("JWT_SIGNING_KEY"),
      },
      Database :    DatabaseConfigurations{
        DBHost:            must("DATABASE_HOST"),
        DBPort:            must("DATABASE_PORT"),
        DBName:            must("DATABASE_NAME"),
        DBUser:            must("DATABASE_USER"),
        DBPassword:        must("DATABASE_PWD"),
      },
    }
}

func mustDur(key, def string) time.Duration {
    d := getenv(key, def)
    v, err := time.ParseDuration(d)
    if err != nil {
        log.Fatalf("bad duration for %s: %v", key, err)
    }
    return v
}
