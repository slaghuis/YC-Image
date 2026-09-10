package main

import (
  "log"
  "os"
  "time"
  "github.com/joho/godotenv"
  "github.com/slaghuis/YC-Image/pkg/config"
  "github.com/slaghuis/YC-Image/pkg/database"
  "github.com/slaghuis/YC-Image/pkg/gateway"
  "github.com/slaghuis/YC-Image/pkg/handler"
  "github.com/slaghuis/YC-Image/pkg/router"
)

var cfg config.Configurations

func init() {
  // Set the file name of the configurations file
  mode := os.Getenv("GIN_MODE")
  if len(mode) == 0 {
    log.Println("GIN_MODE env var not set, assuming development")
    mode = "development"
  }
  log.Printf("Starting Micro Service in %s mode", mode)
}

func main() {
  // Load environment variables from .env file
  err := godotenv.Load()
  if err != nil {
    log.Fatalf("Error loading .env file: %v", err)
  }
  cfg = config.LoadConfig()

  // Initialize Database
  db, err := database.Connect(cfg.Database)
  if err != nil {
    log.Fatal(err)
    panic("Cannot connect to DB")
  }
  log.Println("Connected to Database!")

  handler := handler.NewHandler(db)
  handler.Migrate()

  // Register after a slight delay to ensure server is up
  go func() {
      time.Sleep(2 * time.Second)
      gateway.RegisterWithGateway()
  }()

  router := router.NewRouter(handler, cfg.Security)
  router.Run(":" + cfg.Server.Port)
}
