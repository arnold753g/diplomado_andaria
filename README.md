# Andaria

Plataforma turística con aplicación web y Android. Permite explorar atracciones y paquetes, comprar cupos, revisar pagos y gestionar cancelaciones y reembolsos.

## Estructura

- `backend/`: API Go, PostgreSQL, migraciones SQL y pruebas.
- `frontend/`: web Nuxt, Vue, Pinia y PrimeVue.
- `mobile/`: aplicación Flutter para turistas.
- `nginx/`: proxy y cabeceras de seguridad.
- `.github/`: verificaciones automáticas de calidad y seguridad.
- `tools/security/`: comprobación de dependencias.

## Ejecución local

Requisitos: Go 1.27.3, Node.js 22, PostgreSQL y, para Android, Flutter 3.44.8 con Android SDK.

1. Copia `.env.example` a `.env` y configura tu base de datos y secretos.
2. Desde `backend/`, ejecuta `go run ./cmd/migrate` y `go run ./cmd/api`.
3. Desde `frontend/`, ejecuta `npm ci` y `npm run dev`.
4. Para Android, sigue las instrucciones de [mobile/README.md](mobile/README.md).

También puedes iniciar los servicios locales con `docker compose up --build`.

## Funciones

- Acceso por contraseña y Google, perfiles y permisos por rol.
- Agencias, atracciones, favoritos, paquetes y salidas.
- Compras con cupos, comprobantes y revisión manual de pagos.
- Autorización de salidas, cancelaciones y seguimiento de devoluciones.
- Comprobantes PDF y paneles por rol.

Los movimientos bancarios se realizan fuera de la plataforma. El sistema conserva sus evidencias y estados de revisión.

## Verificación

- Backend: `go test ./... -count=1` y `go vet ./...`.
- Web: `npm test`, `npm run typecheck` y `npm run build`.
- Android: `flutter analyze`, `flutter test` y `flutter build apk --debug`.
- Para las pruebas de integración, configura `TEST_DATABASE_DSN` con una base de pruebas: se utilizan esquemas temporales.

## Producción

Configura `.env.production` a partir de `.env.production.example`, secretos propios y un proxy HTTPS. El archivo `backend/scripts/database-roles.sql` permite preparar roles separados para migraciones y ejecución. Aplica las migraciones antes de iniciar la API.

La composición de producción se inicia con `docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d`.

Los entornos reales, claves de firma, compilaciones, auditorías y documentación académica se conservan fuera del repositorio.
