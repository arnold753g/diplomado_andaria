package database

import (
	"fmt"

	"gorm.io/gorm"
)

// ValidateSchemaVersion prevents serving requests with the old persistence
// schema when production migrations have not yet been run by the migrator.
func ValidateSchemaVersion(db *gorm.DB) error {
	var ready bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema=current_schema() AND table_name='usuarios' AND column_name='correo'
	) AND EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema=current_schema() AND table_name='migraciones_esquema' AND column_name='nombre'
	) AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='salidas_paquete' AND column_name='autorizado_en')
      AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='compras_paquete' AND column_name='datos_comprobante')`).Scan(&ready).Error; err != nil || !ready {
		return fmt.Errorf("database schema is not ready; run the migration service before starting the API")
	}
	var count int64
	if err := db.Table("migraciones_esquema").Where("version = ?", 17).Count(&count).Error; err != nil || count != 1 {
		return fmt.Errorf("database migration 17 is required; run the migration service before starting the API")
	}
	return nil
}
