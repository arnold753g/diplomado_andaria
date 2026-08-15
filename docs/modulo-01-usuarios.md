# Módulo 1: usuarios y acceso

## Alcance entregado

- Registro público de turistas con nombre, apellido, correo y contraseña; teléfono, documento y nacionalidad opcionales.
- Login, logout, restauración de sesión y perfil editable.
- Cambio de contraseña con revocación de sesiones; cuentas creadas con Google pueden establecer una contraseña desde su sesión.
- Roles: `admin`, `turista`, `encargado_agencia`, `encargado_atraccion`.
- Administración: alta, listado paginado, búsqueda, filtros, detalle y edición de datos, rol y estado.
- Solo el administrador asigna roles. El registro público rechaza un campo `role`.
- Cambio de rol y desactivación revocan sesiones. No se permite modificar el propio rol/estado ni dejar el sistema sin administrador activo.
- Google: alta de turistas, ingreso de cuentas vinculadas y vinculación explícita desde el perfil.
- Logo de Andaria reutilizado; tema oscuro y paleta de la base conservados.

Las asignaciones de agencias y atracciones pertenecen a los módulos siguientes. Tampoco se incluyen aún recuperación de contraseña por correo, compras, favoritos ni la app Flutter.

## Revisión local

La sesión de desarrollo de esta entrega utiliza:

- Interfaz: `http://127.0.0.1:3000/login`.
- Backend: `http://127.0.0.1:8082/api/v1`.

Se eligieron estas direcciones para evitar interferir con las instancias existentes en localhost/IPv6 y el puerto 8080. Usa el mismo hostname para frontend y API: mezclar `127.0.0.1` y `localhost` impide el funcionamiento esperado de las cookies SameSite.

Para reiniciar esta vista, en dos terminales desde la raíz:

```powershell
Set-Location backend
$env:SERVER_HOST='127.0.0.1'
$env:SERVER_PORT='8082'
$env:ALLOWED_ORIGINS='http://localhost:3000,http://127.0.0.1:3000'
go run ./cmd/migrate
go run ./cmd/api
```

```powershell
Set-Location frontend
$env:NUXT_PUBLIC_API_BASE='http://127.0.0.1:8082/api/v1'
npm run dev -- --host 127.0.0.1 --port 3000
```

Si Windows restringe la caché global de Go, define `GOCACHE` dentro del backend:

```powershell
$env:GOCACHE = Join-Path (Get-Location) '.cache\go-build'
```

Las cuentas de desarrollo se crean con los valores `DEV_ADMIN_EMAIL`, `DEV_ADMIN_PASSWORD`, `DEV_USER_EMAIL` y `DEV_USER_PASSWORD` de tu `.env`:

```powershell
Set-Location backend
$env:SEED_DEVELOPMENT='true'
go run ./cmd/seed
```

El seed no modifica cuentas existentes y está bloqueado en producción. No uses las credenciales de desarrollo para el despliegue final.

### Recorrido sugerido

1. Registrar un turista; comprobar confirmación de contraseña y correo duplicado.
2. Ingresar, completar el perfil y recargar para verificar persistencia.
3. Ingresar como administrador; crear una cuenta de encargado, buscarla y editarla.
4. Ingresar como encargado en otra sesión de navegador; comprobar que no accede a `/admin/users`.
5. Desactivar ese encargado o cambiar su rol; su próxima solicitud debe exigir iniciar sesión de nuevo.
6. Configurar Google y probar alta, ingreso y vinculación según la sección siguiente.

## Configuración de Google

Implementado con autorización de servidor, PKCE S256 y estado aleatorio ligado al navegador. El intento vence a los diez minutos y solo puede utilizarse una vez. Los tokens de Google se usan únicamente en el backend para consultar UserInfo; no se guardan ni se entregan al frontend.

Crear un cliente OAuth de tipo **Aplicación web** en Google Cloud y configurar su pantalla de consentimiento. En modo de pruebas de Google, agregar las cuentas de prueba que podrán acceder.

Para la vista local de esta entrega, registrar exactamente esta URI de redirección:

```text
http://127.0.0.1:8082/api/v1/auth/google/callback
```

Definir en `.env` y reiniciar el backend:

```dotenv
GOOGLE_CLIENT_ID=<id del cliente web>
GOOGLE_CLIENT_SECRET=<secreto del cliente web>
GOOGLE_REDIRECT_URL=http://127.0.0.1:8082/api/v1/auth/google/callback
FRONTEND_URL=http://127.0.0.1:3000
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
```

En producción usar URLs HTTPS del dominio definitivo. El secreto nunca debe usar el prefijo `NUXT_PUBLIC_`. Ambos Compose pasan las variables de Google solamente al backend.

Si faltan las tres variables de Google, el acceso queda deshabilitado y se informa en la pantalla. Una configuración incompleta impide arrancar, evitando aparentar que la integración está disponible.

Si el correo ya tiene cuenta local, el login con Google solicita entrar con contraseña y vincular desde **Mi perfil**. La vinculación exige el mismo correo, una sesión vigente y protección CSRF. No asigna ni cambia roles. Una cuenta desactivada tampoco puede ingresar con Google.

Referencia: [Google — OAuth para aplicaciones web](https://developers.google.com/identity/protocols/oauth2/web-server) y [OpenID Connect / UserInfo](https://developers.google.com/identity/openid-connect/openid-connect).

## API incorporada

| Método | Ruta | Acceso |
|---|---|---|
| GET | `/api/v1/auth/options` | Público: disponibilidad de Google y registro |
| POST | `/api/v1/auth/google/start` | Público: iniciar Google |
| GET | `/api/v1/auth/google/callback` | Estado OAuth y cookie del navegador |
| POST | `/api/v1/me/google/start` | Sesión + CSRF: vincular Google |
| POST | `/api/v1/admin/users` | Admin + CSRF |
| PATCH | `/api/v1/admin/users/:id` | Admin + CSRF, modificación atómica |

Se mantienen las rutas de autenticación, perfil, listado/detalle y cambios de rol/estado del starter. El correo no se modifica desde edición de perfil para no introducir un flujo de cambio de identidad sin verificación.

Migración `000002_users_google.sql`: convierte `user` a `turista`, amplía los roles y agrega los datos de perfil, identidad Google e intentos OAuth. No elimina usuarios ni cambia contraseñas.

## Verificación

```powershell
# Backend
go test ./...
go vet ./...

# Frontend
npm test
npm run typecheck
npm run build
```

Las pruebas de integración requieren `TEST_DATABASE_DSN` en formato PostgreSQL de palabras clave. Crean y eliminan únicamente esquemas `module1_test_<identificador>`; no modifican las tablas de la aplicación. Sin esa variable se omiten explícitamente.

Cubren migraciones repetibles, registro sin escalamiento de rol, duplicados, permisos de los cuatro roles, CSRF, edición, revocación de sesiones, cuentas inactivas y cambio de contraseña. Las pruebas de Google cubren correo verificado, identidad estable, prohibición de vinculación automática, vinculación explícita y rechazo de callbacks repetidos.

La autenticación real contra Google queda pendiente de configurar las credenciales OAuth. Las pruebas del proveedor usan respuestas simuladas. La publicación con dominio, TLS, correo y validación integral del proxy corresponde a la etapa de despliegue.
