-- Compatible con instalaciones previas: conserva versiones y fechas aplicadas.
DO $registro$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name='schema_migrations') THEN
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name='migraciones_esquema') THEN
            RAISE EXCEPTION 'Existen dos registros de migraciones; resolver antes de continuar';
        END IF;
        ALTER TABLE schema_migrations RENAME TO migraciones_esquema;
        ALTER TABLE migraciones_esquema RENAME COLUMN name TO nombre;
        ALTER TABLE migraciones_esquema RENAME COLUMN applied_at TO aplicado_en;
        ALTER TABLE migraciones_esquema RENAME CONSTRAINT schema_migrations_pkey TO migraciones_esquema_clave_primaria;
    END IF;
END;
$registro$;
CREATE TABLE IF NOT EXISTS migraciones_esquema (
    version BIGINT CONSTRAINT migraciones_esquema_clave_primaria PRIMARY KEY,
    nombre TEXT NOT NULL,
    aplicado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
