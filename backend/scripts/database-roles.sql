-- Run with psql as the existing database administrator, on this application's
-- dedicated database only. Credentials are read from environment, not SQL args.
\set ON_ERROR_STOP on
\getenv runtime_password ANDARIA_DB_RUNTIME_PASSWORD
\getenv migration_password ANDARIA_DB_MIGRATION_PASSWORD
SELECT length(:'runtime_password') >= 24 AND length(:'migration_password') >= 24
 AND :'runtime_password' <> :'migration_password' AS valid_passwords \gset
\if :valid_passwords
\else
 \echo 'Two distinct passwords of at least 24 characters are required.'
 \quit 3
\endif
BEGIN;
-- Bootstrap only the migration ledger, so a fresh database can use the limited
-- migrator from its first application migration onward.
SET LOCAL search_path TO public;
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
SELECT 'CREATE ROLE andaria_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS NOREPLICATION'
 WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname='andaria_migrator') \gexec
SELECT 'CREATE ROLE andaria_api LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS NOREPLICATION'
 WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname='andaria_api') \gexec
ALTER ROLE andaria_migrator NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS NOREPLICATION;
ALTER ROLE andaria_api NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS NOREPLICATION;
SELECT format('ALTER ROLE andaria_migrator PASSWORD %L', :'migration_password') \gexec
SELECT format('ALTER ROLE andaria_api PASSWORD %L', :'runtime_password') \gexec
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO andaria_migrator;
GRANT USAGE ON SCHEMA public TO andaria_api;
SELECT format('GRANT CONNECT ON DATABASE %I TO andaria_api, andaria_migrator',current_database()) \gexec
SELECT format('REVOKE CREATE ON DATABASE %I FROM PUBLIC, andaria_api',current_database()) \gexec
SELECT format('ALTER TABLE %I.%I OWNER TO andaria_migrator',n.nspname,c.relname)
 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname='public' AND c.relkind IN ('r','p') \gexec
SELECT format('ALTER SEQUENCE %I.%I OWNER TO andaria_migrator',n.nspname,c.relname)
 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname='public' AND c.relkind='S' \gexec
SELECT format('ALTER FUNCTION %s OWNER TO andaria_migrator',p.oid::regprocedure)
 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
 WHERE n.nspname='public' AND p.prokind='f' \gexec
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO andaria_api;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO andaria_api;
ALTER DEFAULT PRIVILEGES FOR ROLE andaria_migrator IN SCHEMA public
 GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO andaria_api;
ALTER DEFAULT PRIVILEGES FOR ROLE andaria_migrator IN SCHEMA public
 GRANT USAGE, SELECT ON SEQUENCES TO andaria_api;
REVOKE INSERT, UPDATE, DELETE ON public.migraciones_esquema FROM andaria_api;
COMMIT;
