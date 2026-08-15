package handlers

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http/httptest"
	"starter-backend/internal/config"
	"starter-backend/internal/models"
	"starter-backend/internal/testutil"
	"sync"
	"testing"
	"time"
)

func TestConcurrentAdminChangesKeepOneActiveAdmin(t *testing.T) {
	db := testutil.Database(t)
	users := make([]models.User, 2)
	for i := range users {
		users[i] = models.User{Email: fmt.Sprintf("admin%d@example.test", i), FirstName: "Admin", LastName: "Test", Role: models.RoleAdmin, Status: models.StatusActive}
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			results <- db.Transaction(func(tx *gorm.DB) error {
				if err := guardLastAdmin(tx, id, models.RoleUser, models.StatusActive); err != nil {
					return err
				}
				return tx.Model(&models.User{}).Where("id = ?", id).Update("role", models.RoleUser).Error
			})
		}(user.ID)
	}
	wg.Wait()
	close(results)
	success, blocked := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, errLastAdmin) {
			blocked++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || blocked != 1 {
		t.Fatalf("success=%d blocked=%d", success, blocked)
	}
}

func TestSessionCannotBeIssuedFromStaleCredentials(t *testing.T) {
	db := testutil.Database(t)
	user := models.User{Email: "stale@example.test", PasswordHash: "old-hash", FirstName: "Stale", LastName: "Test", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.User{}).Where("id = ?", user.ID).Update("password_hash", "new-hash").Error; err != nil {
		t.Fatal(err)
	}
	h, err := NewAuthHandler(db, config.Config{PasswordBcryptCost: 10, SessionTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	if _, err := h.issueSession(w, httptest.NewRequest("POST", "/auth/login", nil), &user); err == nil {
		t.Fatal("issued session using stale credentials")
	}
	var count int64
	db.Model(&models.Session{}).Count(&count)
	if count != 0 || len(w.Result().Cookies()) != 0 {
		t.Fatal("stale login created an active session")
	}
}
