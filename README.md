# Andaria â€” versiÃ³n para el diplomado

MÃ³dulos disponibles: usuarios, cuatro roles, perfiles y acceso mediante correo/contraseÃ±a o Google; agencias, encargados y configuraciÃ³n de tarifas y medios de pago; atracciones, clasificaciÃ³n, horarios, catÃ¡logo, mapas y favoritos. Paquetes incorpora borradores, tarifas, fotografÃ­as, itinerarios, publicaciÃ³n y programaciÃ³n de salidas Ãºnicas, diarias o por dÃ­as especÃ­ficos con ventanas de compra. El turista puede comprar cupos, adjuntar y corregir su comprobante, cancelar dentro de la polÃ­tica y elegir QR o cuenta bancaria para la devoluciÃ³n. La agencia revisa pagos, confirma cancelaciones por cupo mÃ­nimo y registra cada reembolso con comprobante dentro de 72 horas. La portada pÃºblica reÃºne el catÃ¡logo y cada rol recibe un panel con sus alertas y cifras operativas. Consulta las guÃ­as de [usuarios](docs/modulo-01-usuarios.md), [agencias](docs/modulo-02-agencias.md), [atracciones](docs/modulo-03-atracciones.md), [paquetes](docs/modulo-04-paquetes.md), [compras](docs/modulo-05-compras.md), [cancelaciones y reembolsos](docs/modulo-06-cancelaciones-reembolsos.md) y [home y paneles](docs/modulo-07-home-paneles.md).

## Arquitectura

```text
Browser
  â””â”€ Nuxt 4 / Vue 3 / Pinia / PrimeVue
       â””â”€ API JSON Go 1.27 / Gorilla Mux / GORM
            â””â”€ PostgreSQL 16
```

- `frontend/`: interfaz SSR, layouts `/app` y `/admin`, guards cliente y tema centralizado.
- `mobile/`: app Android Flutter + Dart para turistas, con API compartida. Consulta [instalaciÃ³n y pruebas](mobile/README.md) y [contrato mÃ³vil](docs/app-android.md).
- `backend/cmd/api`: servidor HTTP.
- `backend/cmd/migrate`: migraciones SQL versionadas e idempotentes.
- `backend/cmd/seed`: dos usuarios configurables, solo para desarrollo/pruebas.
- `backend/internal/httpapi`: composiciÃ³n explÃ­cita de rutas y middleware.
- `backend/internal/handlers`: autenticaciÃ³n, perfiles, usuarios, agencias, atracciones y favoritos.
- `backend/internal/middleware`: sesiÃ³n, CSRF, RBAC, rate limiting, logging y cabeceras.
- `backend/internal/database/migrations`: fuente de verdad del esquema.
- `nginx/`: reverse proxy de producciÃ³n con servicios internos no publicados.

## Requisitos

- Go 1.27.3, alineado con el toolchain, Docker y CI.
- Node.js 22 y npm.
- PostgreSQL 16 o 17.
- Docker Compose, opcional pero recomendado.

## InstalaciÃ³n rÃ¡pida con Docker

1. Crea la configuraciÃ³n local:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Sustituye en `.env` al menos `DB_PASSWORD` y `SESSION_SECRET`. Usa valores aleatorios; `SESSION_SECRET` debe tener 32 caracteres o mÃ¡s.

3. Inicia los servicios:

   ```powershell
   docker compose up --build
   ```

La base se publica Ãºnicamente en `127.0.0.1`, igual que frontend y backend. En desarrollo, el backend aplica las migraciones al arrancar porque `DB_AUTO_MIGRATE=true` en Compose.

## Desarrollo local

Primero crea la base PostgreSQL indicada por `DB_NAME` y copia `.env.example` como `.env` en la raÃ­z. El backend carga `../.env` cuando se ejecuta desde su carpeta.

Backend:

```powershell
Set-Location backend
go mod download
go run ./cmd/migrate
go run ./cmd/api
```

Frontend, en otra terminal:

```powershell
Set-Location frontend
npm ci
npm run dev
```

Direcciones predeterminadas:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080/api/v1`
- Salud: `http://localhost:8080/health`

## Variables de entorno

La lista completa y segura estÃ¡ en [`.env.example`](./.env.example). No se debe versionar `.env`.

| Grupo | Variables principales | PropÃ³sito |
| --- | --- | --- |
| App | `APP_ENV`, `APP_NAME`, `APP_SHORT_NAME`, `LOG_LEVEL` | Entorno, nombre y nivel de logs |
| Frontend | `NUXT_PUBLIC_API_BASE`, `NUXT_PUBLIC_REGISTRATION_ENABLED`, `FRONTEND_PORT` | URL pÃºblica de API y comportamiento pÃºblico |
| Backend | `SERVER_HOST`, `SERVER_PORT`, `BACKEND_PORT`, `ALLOWED_ORIGINS`, `TRUST_PROXY` | Escucha HTTP, CORS y proxies confiables |
| Base de datos | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_AUTO_MIGRATE` | ConexiÃ³n y estrategia de migraciÃ³n |
| Sesiones | `SESSION_SECRET`, `SESSION_TTL`, `SESSION_IDLE_TIMEOUT` | Firma CSRF y vencimientos absoluto/inactivo |
| Cookies | `AUTH_COOKIE_SECURE`, `AUTH_COOKIE_SAME_SITE` | Seguridad de la cookie HttpOnly |
| ContraseÃ±as | `PASSWORD_BCRYPT_COST`, `LOGIN_MAX_ATTEMPTS`, `LOGIN_LOCKOUT` | Hashing y bloqueo contra fuerza bruta |
| LÃ­mites | `AUTH_RATE_LIMIT_PER_MINUTE`, `API_RATE_LIMIT_PER_MINUTE` | LÃ­mites por IP y proceso |
| Seed | `SEED_DEVELOPMENT`, `DEV_*` | Usuarios configurables no productivos |

La aplicaciÃ³n falla al arrancar si falta una variable crÃ­tica o si producciÃ³n intenta usar una cookie no segura, un secreto corto/placeholder, CORS comodÃ­n o una combinaciÃ³n `SameSite=None` sin `Secure`.

## Base de datos y migraciones

El esquema se crea Ãºnicamente desde SQL versionado. No depende de cambios manuales ni de `AutoMigrate` de GORM.

```powershell
Set-Location backend
go run ./cmd/migrate
```

El comando puede ejecutarse repetidamente: registra versiones en `schema_migrations` y aplica cada archivo una sola vez dentro de una transacciÃ³n. Para agregar una migraciÃ³n, crea el siguiente archivo secuencial en `backend/internal/database/migrations`, por ejemplo `000002_feature.sql`.

En producciÃ³n, el servicio `migrate` de `docker-compose.prod.yml` debe terminar correctamente antes de que el backend inicie.

## Seed de desarrollo

El seed crea exactamente un `admin` y un `turista`, usando correos y contraseÃ±as definidos por variables:

```powershell
$env:SEED_DEVELOPMENT='true'
Set-Location backend
go run ./cmd/seed
```

Con Docker:

```powershell
docker compose run --rm backend /app/seed
```

El comando rechaza `APP_ENV=production`, exige confirmaciÃ³n mediante `SEED_DEVELOPMENT=true`, valida la polÃ­tica de contraseÃ±a y nunca contiene contraseÃ±as codificadas en el repositorio.

Para cargar tres paquetes demostrativos con programaciÃ³n Ãºnica, diaria y por dÃ­as especÃ­ficos, ejecuta despuÃ©s de crear la agencia de desarrollo:

```powershell
Set-Location backend
go run ./cmd/seed-demo-packages
```

Este segundo comando tambiÃ©n estÃ¡ limitado a desarrollo. Es idempotente mientras sus paquetes no tengan compras y no modifica datos comerciales ajenos.

## AutenticaciÃ³n

No se utilizan JWT ni tokens en `localStorage`/`sessionStorage`.

1. Login genera 256 bits aleatorios.
2. El valor crudo vive solo en una cookie `HttpOnly` de mismo origen. Su alcance `/` permite que Nuxt valide la sesiÃ³n durante SSR sin exponerla a JavaScript.
3. PostgreSQL almacena Ãºnicamente SHA-256 del token.
4. El frontend conserva en memoria un token CSRF HMAC; toda mutaciÃ³n autenticada debe enviarlo en `X-CSRF-Token`.
5. La sesiÃ³n tiene expiraciÃ³n absoluta e inactividad deslizante.
6. Logout, cambio de contraseÃ±a, cambio de rol y desactivaciÃ³n revocan sesiones en servidor.

La cookie usa `Secure=true` obligatoriamente en producciÃ³n. `SameSite=lax` es el valor recomendado para un frontend y API del mismo sitio.

## Usuarios y roles

- `admin`: `/admin`, dashboard, listado/detalle de usuarios, cambio controlado de rol/estado y perfil propio.
- `turista`: `/app`, perfil y cambio de contraseÃ±a.
- `encargado_agencia`: `/app`, perfil y configuraciÃ³n de su agencia asignada.
- `encargado_atraccion`: `/app`, perfil y ediciÃ³n de una o varias atracciones asignadas.

El registro pÃºblico, cuando estÃ¡ habilitado, siempre crea `turista`. El frontend orienta la navegaciÃ³n, pero el backend es la autoridad: todas las rutas `/admin/*` pasan por autenticaciÃ³n y `RequireRoles("admin")`. Un administrador tampoco puede cambiar su propio rol o estado desde la tabla administrativa.

## Rutas base

Frontend:

```text
/                          (portada pÃºblica)
/login
/register                 (solo si el registro estÃ¡ habilitado)
/app
/app/profile
/app/favorites
/app/purchases
/admin
/admin/users
/admin/agencies
/admin/attractions
/agency/packages
/agency/packages/new
/agency/packages/:id
/agency/purchases
/managed-attractions
/attractions
/attractions/:id
/packages
/packages/:id
/packages/:id/purchase
/403
/cualquier-ruta-invalida  (404)
```

API:

```text
GET    /health
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/auth/session
POST   /api/v1/auth/logout
GET    /api/v1/me
GET    /api/v1/me/dashboard
PATCH  /api/v1/me
POST   /api/v1/me/password
GET    /api/v1/admin/dashboard
GET    /api/v1/admin/users
GET    /api/v1/admin/users/:id
PATCH  /api/v1/admin/users/:id/role
PATCH  /api/v1/admin/users/:id/status
GET    /api/v1/attractions
GET    /api/v1/attractions/:id
GET    /api/v1/admin/attractions
POST   /api/v1/admin/attractions
GET    /api/v1/managed-attractions
GET    /api/v1/me/favorites
PUT    /api/v1/me/favorites/:id
DELETE /api/v1/me/favorites/:id
GET    /api/v1/packages/options
GET    /api/v1/packages
GET    /api/v1/packages/:id
GET    /api/v1/packages/:id/photos/:photo
GET    /api/v1/agency/packages
POST   /api/v1/agency/packages
GET    /api/v1/agency/packages/attractions
GET    /api/v1/agency/packages/:id
PUT    /api/v1/agency/packages/:id
GET    /api/v1/agency/packages/:id/photos/:photo
PATCH  /api/v1/agency/packages/:id/departures/:departure
POST   /api/v1/agency/packages/:id/departures/:departure/cancel
POST   /api/v1/agency/packages/:id/departures/:departure/minimum-refund
GET    /api/v1/me/packages/:id/payment-options
GET    /api/v1/me/purchases
POST   /api/v1/me/purchases
GET    /api/v1/me/purchases/:id
GET    /api/v1/me/purchases/:id/proof
PATCH  /api/v1/me/purchases/:id/proof
POST   /api/v1/me/purchases/:id/cancel
PATCH  /api/v1/me/purchases/:id/refund-destination
GET    /api/v1/me/purchases/:id/refund-qr
GET    /api/v1/me/purchases/:id/refund-proof
GET    /api/v1/agency/purchases
GET    /api/v1/agency/purchases/:id
GET    /api/v1/agency/purchases/:id/proof
GET    /api/v1/agency/purchases/:id/refund-qr
GET    /api/v1/agency/purchases/:id/refund-proof
PATCH  /api/v1/agency/purchases/:id/review
PATCH  /api/v1/agency/purchases/:id/refund
```

## Tema visual

La fuente de verdad es [`frontend/assets/theme.css`](./frontend/assets/theme.css).

Para cambiar colores, edita una vez:

```css
--color-primary          /* #BAFE06: CTA, enlaces e indicadores */
--color-secondary        /* #AACA02: hover y degradados */
--color-accent           /* #7D9505: brillo y decoraciÃ³n */
--color-accent-dark      /* #333D04: superficies teÃ±idas */
--color-background       /* #080C0F: fondo principal */
--color-background-deep  /* #010100: secciones profundas */
--color-surface          /* #111517: cards */
--color-surface-elevated /* #181C1D: inputs y elevaciÃ³n */
--color-text             /* #F8F9F2: contenido principal */
--color-text-muted       /* #B9BABA: contenido secundario */
--color-text-disabled    /* #757B7A: contenido desactivado */
--color-border           /* #303432: bordes y separadores */
--color-success
--color-warning
--color-danger
```

La plantilla utiliza una identidad oscura Ãºnica y la clase raÃ­z `.app-dark` permite que PrimeVue consuma su variante oscura. Los componentes consumen tokens semÃ¡nticos y no contienen paletas de cliente dispersas.

Para cambiar fuentes, edita:

```css
--font-primary
--font-secondary
--font-size-base
--font-weight-regular
--font-weight-medium
--font-weight-semibold
--font-weight-bold
--font-weight-extrabold
```

La fuente configurada es Outfit y se carga en `frontend/nuxt.config.ts` Ãºnicamente en los pesos 400, 500, 600, 700 y 800. Si agregas fuentes propias, define `@font-face` en el mismo archivo y mantÃ©n los archivos bajo `frontend/public/fonts/`.

## Branding

- Nombre completo: `APP_NAME` / `NUXT_PUBLIC_APP_NAME`.
- Nombre corto: `APP_SHORT_NAME` / `NUXT_PUBLIC_APP_SHORT_NAME`.
- Logo: reemplaza `frontend/public/branding/logo.svg` conservando la ruta.
- Favicon: reemplaza `frontend/public/branding/favicon.svg`.
- Colores y fuentes: `frontend/assets/theme.css`.

No hay referencias al logo, imÃ¡genes ni textos comerciales del proyecto original.

## Componentes reutilizables

PrimeVue cubre Button, Input, Password, Select, Dialog, ConfirmDialog, Card, DataTable, Tag, Message, Toast, Spinner y Paginator. Se aÃ±adieron `AppShell`, `EmptyState` y `LoadingState` solo donde existe valor comÃºn real. Esto evita mantener duplicados de la librerÃ­a UI.

## Calidad y pruebas

```powershell
# Backend
Set-Location backend
go test ./...
go vet ./...
go build ./cmd/api
go build ./cmd/migrate
go build ./cmd/seed

# Frontend
Set-Location ../frontend
npm test
npm run typecheck
npm run build
npm audit --omit=dev
```

No hay script de lint separado; `go vet`, `gofmt` y el typecheck estricto cubren las comprobaciones estÃ¡ticas disponibles en este starter.

La Ãºltima revisiÃ³n funcional y tÃ©cnica estÃ¡ resumida en [auditorÃ­a integral](docs/auditoria-integral.md), con el resultado detallado de Medusa Auditor en `.medusa-auditor/results/integrated-release-report.md`.

## ProducciÃ³n

La guÃ­a operativa completa estÃ¡ en [despliegue de producciÃ³n](docs/despliegue-produccion.md) y la plantilla segura en [`.env.production.example`](./.env.production.example).

1. Usa secretos aleatorios externos al repositorio.
2. Configura `APP_ENV=production`, orÃ­genes HTTPS exactos y `NUXT_PUBLIC_API_BASE` pÃºblico.
3. MantÃ©n `AUTH_COOKIE_SECURE=true`; el Compose de producciÃ³n lo fuerza.
4. Termina TLS en un balanceador/reverse proxy o amplÃ­a Nginx con certificados administrados. No publiques HTTP directo a Internet.
5. Usa `DB_SSLMODE=require` cuando PostgreSQL estÃ© fuera de la red privada.
6. MantÃ©n `REGISTRATION_ENABLED=false` si no necesitas altas pÃºblicas.
7. Cambia el rate limiter en memoria por Redis si despliegas varias rÃ©plicas.
8. Ejecuta backups, monitorizaciÃ³n y rotaciÃ³n de secretos segÃºn tu infraestructura.

Comando base despuÃ©s de crear `.env.production`:

```powershell
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
```

PostgreSQL, backend y frontend no publican puertos en el Compose de producciÃ³n. Nginx se enlaza por defecto a `127.0.0.1:8088` para recibir trÃ¡fico de un proxy TLS.

## Crear un proyecto nuevo desde esta base

1. Copia esta carpeta sin `.git`, `.env`, `node_modules`, `.nuxt`, `.output` ni binarios.
2. Crea un `.env` nuevo desde `.env.example` y genera secretos nuevos.
3. Cambia nombre, logo, favicon, colores y fuentes en los puntos descritos arriba.
4. Conserva `users`, `auth_sessions`, `/me` y RBAC como infraestructura.
5. Agrega cada nuevo dominio en paquetes backend y pÃ¡ginas/componentes frontend propios; no mezcles reglas comerciales dentro de autenticaciÃ³n.
6. AÃ±ade migraciones secuenciales y tests de autorizaciÃ³n para cada endpoint nuevo.
7. Inicializa un repositorio Git nuevo y ejecuta toda la baterÃ­a de calidad.

## Alcance deliberadamente excluido

La plantilla no inventa correo transaccional, verificaciÃ³n de email, recuperaciÃ³n de contraseÃ±a, MFA, almacenamiento de archivos ni permisos granulares. Esas capacidades requieren decisiones de proveedor y producto. AÃ±Ã¡delas cuando el nuevo sistema tenga requisitos concretos.

## RevisiÃ³n de documento y salud (11/10/2026)

La monografÃ­a organiza cuatro mÃ³dulos acadÃ©micos, con 12 RF y 3 roles aprobados por el tutor segÃºn confirmaciÃ³n del autor. El cliente mÃ³vil es transversal para el turista; el cÃ³digo conserva funciones heredadas fuera de la defensa. Consulte [requisitos y pruebas](docs/requisitos-y-pruebas-2026-10-11.md), el [contrato](docs/contrato-api.md) y la [colecciÃ³n Postman](docs/api/Andaria.postman_collection.json).

El backend corregido conserva `/health` y aÃ±ade `/api/v1/health` y `/api/v1/salud`; todas consultan PostgreSQL con `SELECT 1`. La actualizaciÃ³n del VPS y los dos monitores UptimeRobot siguen la [guÃ­a operativa](docs/monitoreo-y-despliegue-salud.md). La disponibilidad mensual no estÃ¡ medida.

Cada nueva revisiÃ³n se registra antes de editar y se cierra despuÃ©s de verificar. Abra [el historial visual](docs/revisiones/index.html) y siga [su procedimiento](docs/revisiones/README.md).

## Esquema de base de datos en español

Los comprobantes PDF de compra se generan en el backend y se consultan desde web y APK. La migración 17 conserva los datos comerciales de cada compra; consulte la [guía de comprobantes](docs/comprobantes-pdf.md).

La persistencia usa identificadores en español mediante la migración 15, conservando el contrato HTTP de la web y la APK. Véanse el [procedimiento de actualización y despliegue](docs/base-datos-espanol.md) y el [diccionario físico actual](docs/diccionario-datos.md). Las migraciones 1–14 se mantienen como historial.
