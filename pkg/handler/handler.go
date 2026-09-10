package handler

import (
  "encoding/json"
  "fmt"
  "strconv"
  "gorm.io/gorm"
  "github.com/slaghuis/YC-Image/pkg/models"
)

type Handler struct {
  db *gorm.DB
}

func NewHandler(database *gorm.DB) *Handler {
  return &Handler{
    db: database,
  }
}

func (h *Handler) Migrate() (err error) {
  h.db.AutoMigrate(
    &models.Image{},
  )

  return nil
}

func (h *Handler) IsNumeric(str string) bool {
  _, err := strconv.ParseUint(str, 10, 64)
  return err == nil
}

func (h *Handler) StrToUint(str string) (uint, error) {
  // Convert string to uint64 first
	// Base 10, 64-bit size
	u64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		//log.Println("Error during conversion:", err)
		return 0, err
	}

	// Explicitly convert the uint64 to a uint
	// Note: The size of uint is implementation defined (either 32 or 64 bits)
	// You may lose data if the number is too large for the target uint type
	// on a 32-bit system
	u := uint(u64)
  return u, err
}

func (h *Handler) StrToInt(str string, def int) (int) {
  num, err := strconv.Atoi(str)
  if err != nil {
    return def
  }
  return num
}

func (h *Handler) LogJson(data any) error {
  // Marshal the struct with indentation (empty prefix, tab indent)
  jsonData, err := json.MarshalIndent(data, "", "  ")
  if err != nil {
    return err
  }

  // Print the pretty-printed JSON string
  fmt.Println(string(jsonData))

  return nil
}
