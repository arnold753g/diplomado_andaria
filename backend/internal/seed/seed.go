package seed

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"starter-backend/internal/models"
	"starter-backend/internal/security"

	"gorm.io/gorm"
)

func Development(db *gorm.DB, appEnv string, bcryptCost int) error {
	if appEnv == "production" {
		return errors.New("development seed is disabled in production")
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_DEVELOPMENT")), "true") {
		return errors.New("set SEED_DEVELOPMENT=true to confirm development seed")
	}
	accounts := []struct {
		emailKey, passwordKey, firstName, lastName, role string
	}{
		{"DEV_ADMIN_EMAIL", "DEV_ADMIN_PASSWORD", "Admin", "Andaria", models.RoleAdmin},
		{"DEV_USER_EMAIL", "DEV_USER_PASSWORD", "Turista", "Andaria", models.RoleUser},
	}
	for _, account := range accounts {
		email := strings.ToLower(strings.TrimSpace(os.Getenv(account.emailKey)))
		password := os.Getenv(account.passwordKey)
		if email == "" || password == "" {
			return fmt.Errorf("%s and %s are required", account.emailKey, account.passwordKey)
		}
		hash, err := security.HashPassword(password, bcryptCost)
		if err != nil {
			return fmt.Errorf("%s: %w", account.passwordKey, err)
		}
		user := models.User{Email: email, PasswordHash: hash, FirstName: account.firstName,
			LastName: account.lastName, Role: account.role, Status: models.StatusActive}
		var existing models.User
		result := db.Where("email = ?", email).Limit(1).Find(&existing)
		err = result.Error
		switch {
		case err == nil && result.RowsAffected == 0:
			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("create %s: %w", account.role, err)
			}
		case err != nil:
			return err
		}
	}
	return nil
}
