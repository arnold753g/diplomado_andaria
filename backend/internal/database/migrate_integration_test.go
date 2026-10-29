package database_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	"starter-backend/internal/database"
	"starter-backend/internal/models"
	"starter-backend/internal/testutil"
)

type schemaNames struct {
	Tables  map[string]string `json:"tables"`
	Columns map[string]string `json:"columns"`
}

func assertSpanishSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	raw, err := os.ReadFile("nombres_esquema.json")
	if err != nil {
		t.Fatal(err)
	}
	var names schemaNames
	if err := json.Unmarshal(raw, &names); err != nil {
		t.Fatal(err)
	}
	allowedTables, allowedColumns := map[string]bool{}, map[string]bool{}
	for _, name := range names.Tables {
		allowedTables[name] = true
	}
	for _, name := range names.Columns {
		allowedColumns[name] = true
	}
	var columns []struct {
		TableName  string
		ColumnName string
	}
	if err := db.Raw("SELECT table_name, column_name FROM information_schema.columns WHERE table_schema=current_schema()").Scan(&columns).Error; err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, c := range columns {
		if !allowedTables[c.TableName] || !allowedColumns[c.ColumnName] {
			t.Errorf("untranslated schema identifier: %s.%s", c.TableName, c.ColumnName)
		}
		seen[c.TableName] = true
	}
	if len(seen) != len(allowedTables) {
		t.Fatalf("tables: got %d want %d", len(seen), len(allowedTables))
	}
	var objects []string
	if err := db.Raw(`SELECT c.relname FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname=current_schema() AND c.relkind IN ('r','p','i','S')
 UNION ALL SELECT c.conname FROM pg_constraint c JOIN pg_namespace n ON n.oid=c.connamespace WHERE n.nspname=current_schema()
 UNION ALL SELECT p.proname FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname=current_schema()
 UNION ALL SELECT t.tgname FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname=current_schema() AND NOT t.tgisinternal`).Scan(&objects).Error; err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	for _, name := range objects {
		if !pattern.MatchString(name) || len(name) > 63 {
			t.Errorf("invalid identifier: %q", name)
		}
		translated := name == "establecer_actualizado_en" || name == "validar_autorizacion_salida"
		for table := range allowedTables {
			if name == table || strings.HasPrefix(name, table+"_") {
				translated = true
			}
		}
		if !translated {
			t.Errorf("untranslated object: %s", name)
		}
	}
	var count int64
	if err := db.Table("migraciones_esquema").Count(&count).Error; err != nil || count != 17 {
		t.Fatalf("migration ledger: %d %v", count, err)
	}
}

func TestSpanishSchemaFreshInstall(t *testing.T) {
	db := testutil.Database(t)
	assertSpanishSchema(t, db)
	if err := database.ValidateSchemaVersion(db); err != nil {
		t.Fatal(err)
	}
	// The existing user-facing API keys do not depend on persistence names.
	user := models.User{Email: "esquema@example.test", FirstName: "Prueba", LastName: "Esquema", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	old := user.UpdatedAt
	if err := db.Model(&user).Update("nombres", "Nombre actualizado").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&user, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if user.FirstName != "Nombre actualizado" || user.UpdatedAt.Before(old) {
		t.Fatal("translated model/updated timestamp not persisted")
	}
	raw, err := json.Marshal(user.Public())
	if err != nil {
		t.Fatal(err)
	}
	var public map[string]any
	json.Unmarshal(raw, &public)
	if public["first_name"] != "Nombre actualizado" || public["email"] != "esquema@example.test" || public["nombres"] != nil || public["correo"] != nil {
		t.Fatalf("API changed: %s", raw)
	}
}

func TestSpanishSchemaUpgradesEnglishDatabaseWithoutLosingData(t *testing.T) {
	db := testutil.UnmigratedDatabase(t)
	if err := database.ValidateSchemaVersion(db); err == nil {
		t.Fatal("empty database accepted by API")
	}
	if err := db.Exec(`CREATE TABLE schema_migrations(version BIGINT PRIMARY KEY,name TEXT NOT NULL,applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`).Error; err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	for i, path := range paths {
		if i >= 14 {
			break
		}
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("INSERT INTO schema_migrations(version,name) VALUES(?,?)", i+1, filepath.Base(path)).Error; err != nil {
			t.Fatal(err)
		}
	}
	fixture := `INSERT INTO users(email,password_hash,first_name,last_name,google_subject) VALUES ('legacy@example.test','hash-prueba','Nombre','Anterior','google-sub-prueba');
 INSERT INTO auth_sessions(id,user_id,token_hash,expires_at) VALUES ('sesion-prueba',1,'token-prueba',NOW()+INTERVAL '1 day');
 INSERT INTO oauth_attempts(state_hash,verifier,user_id,session_id,expires_at) VALUES ('estado-prueba','verificador-prueba',1,'sesion-prueba',NOW()+INTERVAL '1 hour');
 INSERT INTO agencies(name,city,address,phone,email,manager_id) VALUES ('Agencia prueba','Tarija','Direccion prueba','000000','agencia@example.test',1);
 INSERT INTO attractions(name,category,department,city,address,manager_id) VALUES ('Atraccion prueba','Natural','Tarija','Tarija','Direccion prueba',1);
 INSERT INTO attraction_favorites(user_id,attraction_id) VALUES(1,1);
 INSERT INTO attraction_photos(attraction_id,position,image) VALUES(1,0,decode('0102','hex'));
 INSERT INTO attraction_classifications(attraction_id,subcategory_id,position) VALUES(1,14,0);
 INSERT INTO tour_packages(agency_id,name,national_price_cents) VALUES(1,'Paquete prueba',10000);
 INSERT INTO tour_package_photos(package_id,position,image) VALUES(1,0,decode('0102','hex'));
 INSERT INTO tour_package_itinerary_days(package_id,day_number,title) VALUES(1,1,'Dia prueba');
 INSERT INTO tour_package_itinerary_attractions(itinerary_day_id,attraction_id,position) VALUES(1,1,0);
 INSERT INTO tour_package_schedules(package_id,frequency_type,valid_from,valid_until,departure_time,meeting_time,default_min_capacity,default_max_capacity,booking_cutoff_hours,maximum_advance_days) VALUES(1,'single',CURRENT_DATE+10,CURRENT_DATE+10,'08:00','07:45',1,5,24,180);
 INSERT INTO tour_package_departures(package_id,schedule_id,starts_at,meeting_at,booking_opens_at,booking_closes_at,min_capacity,max_capacity,held_capacity) VALUES(1,1,NOW()+INTERVAL '10 days',NOW()+INTERVAL '10 days'-INTERVAL '15 minutes',NOW()-INTERVAL '1 day',NOW()+INTERVAL '9 days',1,5,1);
 INSERT INTO tour_package_purchases(reference,tourist_id,agency_id,package_id,departure_id,payment_method,national_adults,capacity_count,national_unit_price_cents,foreign_surcharge_cents,total_cents,tourist_name,tourist_email,payment_proof) VALUES('COMPRA-PRUEBA',1,1,1,1,'qr',1,1,10000,0,10000,'Nombre Anterior','legacy@example.test',decode('0102','hex'));
 INSERT INTO purchase_requests(tourist_id,request_key,payload_hash,purchase_id) VALUES(1,'clave-prueba','hash-prueba',1);
 INSERT INTO pending_registrations(email,token_hash,password_hash,first_name,last_name,created_at,expires_at) VALUES('pendiente@example.test','token-pendiente','hash-prueba','Pendiente','Prueba',NOW(),NOW()+INTERVAL '1 hour');`
	if err := db.Exec(fixture).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.ValidateSchemaVersion(db); err == nil {
		t.Fatal("legacy English schema accepted by API")
	}
	var applied time.Time
	if err := db.Raw("SELECT applied_at FROM schema_migrations WHERE version=1").Scan(&applied).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	assertSpanishSchema(t, db)
	if err := database.ValidateSchemaVersion(db); err != nil {
		t.Fatal(err)
	}
	var current time.Time
	if err := db.Raw("SELECT aplicado_en FROM migraciones_esquema WHERE version=1").Scan(&current).Error; err != nil || !current.Equal(applied) {
		t.Fatal("original ledger timestamp lost", err)
	}
	var purchase models.TourPackagePurchase
	if err := db.Preload("Package").Preload("Departure").Preload("Agency").First(&purchase, 1).Error; err != nil {
		t.Fatal(err)
	}
	if purchase.Reference != "COMPRA-PRUEBA" || purchase.TotalCents != 10000 || len(purchase.PaymentProof) != 2 || purchase.Package.Name != "Paquete prueba" || purchase.Agency.ManagerID != 1 || purchase.Departure.HeldCapacity != 1 {
		t.Fatalf("purchase graph lost: %+v", purchase)
	}
	if purchase.ReceiptSnapshot.Origin != "migracion" || purchase.ReceiptSnapshot.PackageName != "Paquete prueba" || purchase.ReceiptSnapshot.CapturedAt.IsZero() {
		t.Fatal("legacy receipt provenance missing")
	}
	var user models.User
	if err := db.First(&user, 1).Error; err != nil {
		t.Fatal(err)
	}
	if user.GoogleSubject == nil || *user.GoogleSubject != "google-sub-prueba" || user.PasswordHash != "hash-prueba" {
		t.Fatal("Google/password identity lost")
	}
	// Exercise the PL/pgSQL trigger directly, bypassing GORM's timestamp updates.
	before := user.UpdatedAt
	if err := db.Exec("UPDATE usuarios SET nombres=? WHERE id=?", "Renombrado", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&user, 1).Error; err != nil || !user.UpdatedAt.After(before) {
		t.Fatal("updated timestamp trigger broken", err)
	}
	// Identity sequences and restrictive foreign keys must survive the rename.
	next := models.User{Email: "nuevo@example.test", FirstName: "Nuevo", LastName: "Usuario", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&next).Error; err != nil || next.ID != 2 {
		t.Fatal("sequence broken", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	deleteErr := tx.Exec("DELETE FROM usuarios WHERE id=1").Error
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	if deleteErr == nil {
		t.Fatal("foreign key no longer restricts deletion")
	}
}
