package database

import (
	"fmt"
	"gorm.io/gorm"
)

// ValidateRuntimeRole is deliberately called only by the API, not the migrator.
// PostgreSQL role membership and schema grants must not let a compromised API
// elevate itself to an owner or perform DDL.
func ValidateRuntimeRole(db *gorm.DB) error {
	var unsafe bool
	err := db.Raw(`SELECT
   r.rolsuper OR r.rolcreatedb OR r.rolcreaterole OR r.rolbypassrls OR r.rolreplication
   OR EXISTS (SELECT 1 FROM pg_roles elevated
     WHERE (elevated.rolsuper OR elevated.rolcreatedb OR elevated.rolcreaterole
       OR elevated.rolbypassrls OR elevated.rolreplication OR elevated.rolname LIKE 'pg\_%' ESCAPE '\')
       AND pg_has_role(current_user,elevated.oid,'MEMBER'))
   OR has_schema_privilege(current_user, current_schema(), 'CREATE')
   OR has_database_privilege(current_user, current_database(), 'CREATE')
   OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
     WHERE n.nspname=current_schema() AND c.relkind IN ('r','p')
       AND pg_has_role(current_user,c.relowner,'MEMBER'))
   FROM pg_roles r WHERE r.rolname=current_user`).Scan(&unsafe).Error
	if err != nil {
		return fmt.Errorf("cannot verify runtime database privileges")
	}
	if unsafe {
		return fmt.Errorf("API database role must have no administrative, schema CREATE, or table ownership privileges; use a separate runtime role")
	}
	return nil
}
