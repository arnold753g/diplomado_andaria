// Package testutil supplies isolated PostgreSQL schemas for integration tests.
package testutil

import (
	"crypto/rand"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"starter-backend/internal/database"
	"strings"
	"testing"
)

func Database(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN (PostgreSQL keyword DSN) to run isolated integration tests")
	}
	if strings.HasPrefix(dsn, "postgres:") || strings.HasPrefix(dsn, "postgresql:") {
		t.Fatal("TEST_DATABASE_DSN must use keyword format")
	}
	open := func(dsn string) *gorm.DB {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal("cannot connect to integration database")
		}
		return db
	}
	base := open(dsn)
	var suffix [12]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("module1_test_%x", suffix)
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		if err := base.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("cleanup test schema: %v", err)
		}
		_ = database.Close(base)
	})
	db := open(dsn + " search_path=" + schema)
	t.Cleanup(func() { _ = database.Close(db) })
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migration is not idempotent: %v", err)
	}
	return db
}
