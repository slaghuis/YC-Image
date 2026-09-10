package database
import (
  "fmt"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "github.com/slaghuis/YC-Image/pkg/config"
)

func Connect(db config.DatabaseConfigurations) (instance *gorm.DB, err error) {
  connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.DBUser, db.DBPassword, db.DBHost, db.DBPort, db.DBName)
  instance, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
  return
}
