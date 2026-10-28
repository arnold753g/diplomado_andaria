package database_test

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"gorm.io/gorm"
	"starter-backend/internal/database"
	"starter-backend/internal/testutil"
)

func TestRuntimeRoleRejectsOwnersAndAllowsOnlyApplicationOperations(t *testing.T) {
	db := testutil.Database(t)
	if err := database.ValidateRuntimeRole(db); err == nil {
		t.Fatal("schema owner accepted as API role")
	}
	var canCreateRoles bool
	if err := db.Raw("SELECT rolsuper OR rolcreaterole FROM pg_roles WHERE rolname=current_user").Scan(&canCreateRoles).Error; err != nil {
		t.Fatal(err)
	}
	if !canCreateRoles {
		t.Skip("integration role cannot create a temporary test role")
	}
	var schema string
	if err := db.Raw("SELECT current_schema()").Scan(&schema).Error; err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^module1_test_[0-9a-f]+$`).MatchString(schema) {
		t.Fatal("unexpected test schema")
	}
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatal(err)
	}
	role := "security_test_" + hex.EncodeToString(bytes)
	rollback := errors.New("rollback test role and grants")
	err := db.Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"CREATE ROLE " + role + " NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS",
			fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s", schema, role),
			fmt.Sprintf("GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA %s TO %s", schema, role),
			fmt.Sprintf("GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA %s TO %s", schema, role),
			fmt.Sprintf("REVOKE INSERT,UPDATE,DELETE ON %s.migraciones_esquema FROM %s", schema, role),
			"SET LOCAL ROLE " + role,
		}
		for _, sql := range statements {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		if err := database.ValidateRuntimeRole(tx); err != nil {
			return err
		}
		if err := tx.Exec("INSERT INTO usuarios(correo,hash_contrasena,nombres,apellidos,rol,estado) VALUES ('limited@example.test','test-hash','Prueba','Permisos','turista','active')").Error; err != nil {
			return err
		}
		for i, sql := range []string{"CREATE TABLE forbidden_table(id INT)", "UPDATE migraciones_esquema SET nombre='forbidden'", "TRUNCATE usuarios", "ALTER TABLE usuarios ADD COLUMN forbidden INT"} {
			point := fmt.Sprintf("denied%d", i)
			if err := tx.SavePoint(point).Error; err != nil {
				return err
			}
			if err := tx.Exec(sql).Error; err == nil {
				return fmt.Errorf("runtime role allowed forbidden operation %d", i)
			}
			if err := tx.RollbackTo(point).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("RESET ROLE").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE ROLE " + role + "_elevated NOLOGIN CREATEROLE").Error; err != nil {
			return err
		}
		if err := tx.Exec("GRANT " + role + "_elevated TO " + role).Error; err != nil {
			return err
		}
		if err := tx.Exec("SET LOCAL ROLE " + role).Error; err != nil {
			return err
		}
		if err := database.ValidateRuntimeRole(tx); err == nil {
			return fmt.Errorf("API can assume an administrative role")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	var count int64
	if err := db.Raw("SELECT count(*) FROM pg_roles WHERE rolname = ?", role).Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("test role persisted")
	}
}
