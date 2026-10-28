# Revisión de seguridad de Andaria: 20 criterios

Fecha: 23/10/2026, America/La_Paz. Alcance: fuentes actuales de backend Go, frontend Nuxt/Vue, app Flutter/Android, migraciones, composición de producción, Nginx, CI e historial Git disponible.

La revisión encuentra una base de autenticación y autorización bien implementada, pero **requiere actualizar dependencias y reforzar la configuración operativa**. Que las pruebas funcionales pasen no significa que las versiones instaladas estén libres de vulnerabilidades.

No se modificaron fuentes de la aplicación, dependencias, credenciales ni configuración del VPS. Las pruebas de base de datos crearon y eliminaron esquemas temporales locales. Las consultas al sitio público fueron cinco solicitudes GET sin autenticación, sin modificar registros ni realizar pruebas de carga.

## Resultado ejecutivo

- **11 criterios implementados en el alcance comprobado**, 7 parciales, 1 que requiere corrección y 1 no aplicable a esta arquitectura. Estas cantidades no son un porcentaje de seguridad ni una certificación.
- `npm audit`: **24 paquetes con alertas**: 3 críticos, 19 altos y 2 moderados. Incluye alertas transitivas y propagadas: no equivale a 24 fallos explotables en producción.
- `govulncheck`, con Linux/amd64, CGO deshabilitado y **Go 1.25.12**, como el Dockerfile: **30 avisos distintos; 22 con símbolos alcanzables según el análisis estático** del ejecutable `cmd/api`. La configuración de protocolos y los datos que llegan a cada función condicionan su explotabilidad.
- Flutter: **102 paquetes alojados de `pubspec.lock` consultados en OSV, sin avisos devueltos**. Esto no cubre el SDK Flutter, dependencias Maven de Android ni vulnerabilidades todavía no publicadas.
- Gitleaks: 45 commits locales y 396 archivos de texto versionables del árbol actual. Las alertas revisadas correspondieron a valores de prueba y hashes públicos; **no se identificaron credenciales reales entre esas alertas**.
- Las tres variables secretas locales comparables no aparecieron literalmente en los **115 archivos de código/datos del frontend exportado** examinados. No se guardaron sus valores en el informe.
- La suite Go con PostgreSQL local, las **22 pruebas del frontend** y las **19 pruebas móviles de seguridad seleccionadas** finalizaron correctamente.

## Matriz de los 20 criterios

“Implementado” describe evidencia concreta de fuentes y pruebas, con las limitaciones de alcance indicadas; no garantiza ausencia de cualquier variante del ataque.

| N.º | Criterio de la imagen | Resultado | Evidencia y alcance |
|---|---|---|---|
| 1 | Ocultar API keys | Implementado | El secreto OAuth permanece en el backend; `runtimeConfig.public` no lo exporta. La API móvil entrega el **Client ID público**, que no es una credencial secreta. Sin coincidencias de los secretos locales comparados en la salida web examinada. |
| 2 | Eliminar secretos de Git | Implementado en lo examinado | Sin archivos `.env` reales, claves privadas o keystores entre los archivos rastreados identificados. Gitleaks revisó 45 commits y el texto versionable actual; sus alertas se clasificaron como fixtures o checksums. No se auditó el almacenamiento del proveedor Git ni objetos remotos no disponibles localmente. |
| 3 | Key pública para DB | No aplica | Los clientes consumen la API Go; no acceden a una API pública de base de datos como Supabase. No corresponde introducir una clave anónima de DB. Las credenciales PostgreSQL deben seguir exclusivamente en el servidor; revisar H03. |
| 4 | Row-Level Security | Parcial | No se encontraron políticas RLS en las migraciones. La separación se implementa mediante roles y filtros de propietario del backend, con pruebas de aislamiento aprobadas. RLS sería una defensa adicional, no un sustituto de esos filtros. El rol de DB debe limitarse antes de plantear RLS. |
| 5 | Encriptar datos sensibles | Parcial | HTTPS, hashes de tokens y almacenamiento seguro móvil presentes. Documento de identidad, datos de reembolso y comprobantes se guardan sin cifrado de campo en la aplicación. No se comprobó cifrado de discos, volúmenes o backups del VPS. Véase H06. |
| 6 | Forzar autenticación | Implementado | Subrouter protegido con autenticador y roles. Las rutas públicas de catálogo son intencionales. En el sitio público, `/api/v1/me` y `/api/v1/admin/users?limit=1` respondieron 401 sin sesión. |
| 7 | Restringir acceso a registros | Implementado | Filtros por `tourist_id`, `agency_id` y `manager_id`; fotografías y comprobantes también comprueban alcance. Pruebas de aislamiento de agencias, turistas y encargados aprobadas. |
| 8 | Bloquear manipulación de campos | Implementado | DTO de entrada, rechazo de campos desconocidos y listas explícitas de campos actualizables. El rol del registro público se fija en servidor. Precios, cupos, nacionalidad que paga y estado de compra se recalculan o validan en backend. |
| 9 | Proteger sesión/cookies | Parcial | Token aleatorio de 256 bits, hash en DB, cookie HttpOnly/Secure en producción, SameSite configurable, CSRF, expiración absoluta/inactiva y revocación al cambiar contraseña. El cierre de sesión ignora errores de revocación: H07. |
| 10 | Hashear contraseñas | Implementado | bcrypt con coste configurable, 12 por defecto, política de 12–72 bytes y comparación del hash. No se devuelve el hash en JSON. |
| 11 | Rate limiting | Parcial | Límites en Go y Nginx; observados 10/min en opciones de autenticación y 120/min en rutas API. Contadores por proceso/IP y contador de intentos no atómico: H08. La identidad IP a través de Cloudflare y el proxy TLS requiere validación en VPS. |
| 12 | Protección contra bots | Parcial | Limitación de solicitudes y bloqueo temporal presentes. No se encontró desafío adaptativo ni verificación de propiedad del correo para el registro con contraseña. Google sí valida correo verificado. Véase H09. |
| 13 | Parametrizar queries | Implementado | GORM con parámetros `?`/listas parametrizadas; ordenación por opciones conocidas. SQL manual revisado usa constantes o parámetros. El esquema dinámico de pruebas usa bytes aleatorios, no entrada de usuarios. |
| 14 | Validar inputs | Implementado | Validación de tipos, longitudes y reglas de negocio en servidor; JSON único, campos desconocidos rechazados y límite de cuerpo. No depende de la validación web o móvil. |
| 15 | Sanitizar contenido | Implementado en el código revisado | Texto presentado con interpolación Vue o widgets Text; no se encontró `v-html`/`innerHTML` en las fuentes web examinadas. Las imágenes se decodifican y recodifican. El motor SSR tiene actualizaciones pendientes: H02. |
| 16 | Restringir archivos | Implementado | PNG/JPEG verificados por decodificación real, hasta 5 MiB y 4096×4096 para fotografías/comprobantes; QR de agencia con límites menores. Se recodifica la imagen, se limita el cuerpo y se comprueba autorización para archivos privados. |
| 17 | Devolver datos necesarios | Implementado en las rutas revisadas | DTO públicos del catálogo, exclusión de hashes, Google subject y bytes de comprobantes de JSON. Comprobantes y QR privados tienen rutas autorizadas y `private, no-store`. No se observó acceso anónimo a datos de usuarios. |
| 18 | Security headers | Parcial | Presentes CSP, nosniff, DENY, referrer y permissions policy en respuestas públicas comprobadas. La CSP desplegada permite `script-src 'unsafe-inline'` y difiere de la versión del repositorio: H04. |
| 19 | Forzar HTTPS | Parcial | HTTP redirige 301 a HTTPS y la conexión HTTPS validó su certificado durante la consulta. La app release rechaza una API HTTP. No se observó HSTS en las respuestas examinadas: H05. |
| 20 | Escanear dependencias | Requiere corrección | Se ejecutaron escaneos, con alertas relevantes en Go y npm. El workflow de calidad no incorpora estos controles como puerta de seguridad. Se requiere actualizar, reconstruir y volver a escanear: H01, H02 y H10. |

## Hallazgos y correcciones priorizadas

### H01 — Alta: toolchain Go y librerías con avisos de seguridad

**Evidencia:** `backend/Dockerfile:1`, `backend/go.mod:3` y los resultados de `govulncheck-api-linux-go125.jsonl`. El builder y CI fijan Go 1.25.12. La política oficial deja de mantener una rama cuando existen dos versiones principales posteriores; ya existen Go 1.26 y 1.27. [Política e historial oficial de Go](https://go.dev/doc/devel/release).

El análisis específico del ejecutable de API obtuvo 30 avisos de módulo y 22 con rutas de símbolos. Hay avisos de la biblioteca estándar (`net/http`, `crypto/tls`, `crypto/x509`, entre otros) y de dependencias como gRPC y `golang.org/x/text`. Las múltiples trazas de un mismo aviso no se cuentan como vulnerabilidades distintas.

La presencia de una función en el grafo **no demuestra por sí sola un ataque posible por Internet**. Por ejemplo, [GO-2026-6089](https://pkg.go.dev/vuln/GO-2026-6089) requiere HTTP/2 sin cifrar configurado; el código revisado no habilita expresamente ese protocolo. Otros avisos dependen de las entradas o del almacén de certificados; [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809) incorpora rechazo de intermediarios comprometidos. Se deben actualizar igualmente los componentes afectados y comprobar las condiciones de cada aviso.

**Corrección:** migrar Dockerfile, CI y toolchain a una rama Go soportada con parches actuales; actualizar las dependencias indicadas por el escáner y reconstruir los binarios. Las versiones 1.26.10 y 1.27.3 incluyen el arreglo de certificados citado; esto es una referencia mínima de ese aviso, no una garantía sobre futuros avisos. Repetir el análisis para Linux/amd64 y la suite PostgreSQL antes del despliegue.

### H02 — Alta: dependencias web alertadas, incluido el runtime SSR

**Evidencia:** `frontend/package-lock.json`, `frontend/.output/server/package.json` y `npm-audit.json`. El árbol instalado contiene Nuxt 4.5.2, Vue y `@vue/server-renderer` 3.5.41. Este último aparece en el manifiesto de producción generado, por lo que no todos los hallazgos son exclusivamente de desarrollo.

La alerta [GHSA-g2v6-rqmx-r4w6](https://github.com/advisories/GHSA-g2v6-rqmx-r4w6) afecta al renderizador SSR anterior a 3.5.42 cuando recibe nombres de atributos controlados por un atacante. No se identificó en esta revisión una ruta propia que permita inyectar esos nombres; no se afirma una explotación de la web.

También se detectaron alertas críticas de `simple-git`, `@simple-git/argv-parser` y `shell-quote` en el árbol de herramientas. No aparecen como dependencias declaradas del manifiesto de runtime generado examinado, lo que reduce su exposición directa en el contenedor final, pero mantiene riesgo para desarrollo y construcción bajo sus condiciones de uso. Ejemplo y corrección del parser: [GHSA-v5rq-49vh-5v5c](https://github.com/advisories/GHSA-v5rq-49vh-5v5c).

Nuxt tiene además [avisos de route rules](https://github.com/advisories/GHSA-vjh5-4x84-xrqp). No se encontraron reglas propias de Nuxt que se estén usando como barrera de autorización; los permisos están en Go.

**Corrección:** actualizar versiones y lockfile de forma compatible, construir la salida final y volver a comprobar el árbol y el artefacto que se desplegará. La versión efectiva del VPS no se determinó a partir de las cabeceras públicas: estas alertas describen el árbol local auditado.

### H03 — Alta: el despliegue reutiliza la cuenta de arranque de PostgreSQL

**Evidencia:** `docker-compose.prod.yml:7`, `:36` y `:81`: `POSTGRES_USER`, backend y migrador utilizan la misma variable `DB_USER`. La imagen oficial crea `POSTGRES_USER` con privilegios de superusuario en una inicialización nueva. [Documentación oficial de la imagen](https://raw.githubusercontent.com/docker-library/docs/master/postgres/README.md).

No se encontraron pasos que separen después un usuario de ejecución limitado. En una instalación creada con esta composición y sin endurecimiento externo, una vulnerabilidad en el backend podría alcanzar privilegios de administración de DB. **Los privilegios reales de la base ya existente en el VPS están pendientes de consulta**; la configuración indica el riesgo, no prueba que el VPS carezca de un ajuste manual.

**Corrección:** separar la cuenta de migraciones y la cuenta de ejecución; conceder a la API solo las operaciones necesarias, sin superusuario, creación de roles ni administración del esquema. Evaluar RLS después de esta separación. Los superusuarios eluden RLS aunque existan políticas. [PostgreSQL: Row Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html).

### H04 — Media: la CSP pública permite scripts en línea

**Evidencia:** GET de `https://diplomado.andaria.site/login`: `script-src 'self' 'unsafe-inline'`. En cambio, `nginx/default.conf:15` declara `script-src 'self'`. Existe diferencia entre lo versionado y lo desplegado.

Permitir scripts en línea reduce la protección de CSP frente a una inyección de HTML. No se demostró una inyección explotable en esta auditoría.

**Corrección:** identificar la configuración efectiva de Nginx/proxy/CDN, preparar nonces o hashes compatibles con la hidratación de Nuxt, probar ingreso y navegación, y retirar el permiso general de scripts en línea. Registrar esa configuración en el proyecto para evitar divergencias.

### H05 — Media: falta HSTS en las respuestas públicas examinadas

**Evidencia:** HTTP devuelve 301 hacia HTTPS, pero no se observó `Strict-Transport-Security` en la página de acceso ni en las rutas API consultadas.

**Corrección:** configurar HSTS en la capa que entrega HTTPS al visitante, tras comprobar certificado y renovación. Aplicar `includeSubDomains` únicamente si todos los subdominios pertinentes funcionan por HTTPS. La ausencia de HSTS no significa que la conexión HTTPS actual estuviera sin cifrar.

### H06 — Media: no se pudo cerrar la protección de datos en reposo

**Evidencia:** `backend/internal/models/user.go`, `backend/internal/models/purchase.go` y migraciones 10/12: documentos, datos de reembolso y comprobantes se persisten como texto o bytes normales. `docs/despliegue-produccion.md` pide respaldos cifrados y restauraciones de prueba, pero el repositorio no demuestra su ejecución.

Esto demuestra ausencia de cifrado de campo en la aplicación; **no demuestra que el disco o los backups del proveedor estén sin cifrar**, pues no se inspeccionaron.

**Corrección:** comprobar cifrado del volumen y de los backups, destino externo, permisos, separación de claves y una restauración verificable. Definir conservación y eliminación de documentos/comprobantes. Evaluar cifrado de campos concretos según necesidad de acceso y amenaza, sin confundir bcrypt con cifrado reversible de datos personales.

### H07 — Media: el cierre de sesión confirma éxito aunque falle la revocación

**Evidencia de código:** `backend/internal/handlers/auth.go:183`: `Logout` descarta el error del `UPDATE revoked_at`, borra la cookie y devuelve éxito.

Si la escritura de revocación falla mientras la sesión sigue vigente en DB, un cliente que conserve el token anterior puede seguir utilizándolo al recuperarse el acceso a la base. No se comprobó este fallo mediante una caída inducida ni se afectó producción.

**Corrección:** comprobar el resultado de revocación, registrar el fallo sin tokens ni datos personales y devolver un estado que no confirme una revocación no persistida. Mantener el borrado local de la cookie y añadir una prueba con fallo de persistencia.

### H08 — Media: bloqueo de intentos susceptible a concurrencia y límites por proceso/IP

**Evidencia de código:** `backend/internal/handlers/auth.go:129` lee `FailedLoginAttempts`, calcula el incremento en memoria y escribe el valor; `:135` ignora el error. Solicitudes concurrentes pueden leer el mismo valor y perder incrementos. La limitación por IP reduce el abuso, pero no sustituye un contador por cuenta consistente.

`backend/internal/middleware/rate_limit.go` mantiene los límites en memoria por proceso. Son adecuados como primera barrera de una réplica, pero no globales para varias réplicas. Nginx sobrescribe `X-Real-IP`, lo que evita aceptar sin más la cabecera enviada por el visitante en la topología revisada; falta comprobar qué IP ve realmente detrás del proxy TLS y Cloudflare. No se demostró un bypass del límite en el sitio público ni se enviaron ráfagas.

**Corrección:** actualizar intentos de forma atómica y comprobar errores; probar concurrencia. Validar la cadena de proxies confiables y la IP observada. Incorporar un almacén compartido si se escala a varias réplicas y combinar señales por cuenta/IP en autenticación.

### H09 — Media: registro con contraseña sin comprobar propiedad del correo

**Evidencia:** `backend/internal/handlers/auth.go:64` crea directamente un turista activo con el correo indicado. Google, en cambio, exige `email_verified`. No se encontró verificación por correo ni desafío adaptativo en el registro tradicional.

Una persona puede registrar un correo ajeno y dejarlo ocupado, afectando posteriormente el acceso o vinculación de su propietario. No obtiene una contraseña de Google ni privilegios administrativos por hacerlo. El rate limiting limita velocidad, pero no acredita propiedad del correo.

**Corrección:** verificar correo antes de confiar en la dirección para identidad/contacto; definir un flujo seguro para resolver cuentas no verificadas. Añadir protección adaptativa frente a abuso si las métricas lo justifican. La mera ausencia de CAPTCHA no se considera por sí sola una vulnerabilidad.

### H10 — Media: CI no bloquea secretos ni vulnerabilidades conocidas

**Evidencia:** `.github/workflows/quality.yml` ejecuta formato, análisis, compilaciones y pruebas, pero no se encontraron pasos de Gitleaks, govulncheck o auditoría npm/OSV.

**Corrección:** incorporar estos escaneos con políticas explícitas, revisión de falsos positivos y actualización periódica. Escanear también el artefacto y las imágenes por digest, además de las dependencias Maven/Gradle de la APK. El chequeo de calidad funcional actual no detecta toda la deuda de seguridad encontrada aquí.

## Verificaciones y límites de evidencia

| Verificación | Resultado |
|---|---|
| Suite Go con PostgreSQL local | Aprobada; esquemas temporales aislados y limpieza automática. Incluye permisos, CSRF, sesiones, roles, propietarios, precios, cupos, imágenes y concurrencia de operaciones de negocio. |
| Frontend | 22 pruebas aprobadas en esta revisión. |
| Flutter de seguridad | 19 pruebas aprobadas: sesiones, aislamiento de cuenta, Google, imágenes e idempotencia. |
| Gitleaks histórico | 45 commits; 3 alertas de una constante hexadecimal repetida utilizada en pruebas. No se publican sus valores. |
| Gitleaks árbol actual | 396 archivos de texto de hasta 6 MB; 59 alertas: 3 fixtures, 54 coincidencias con checksums públicos de módulos y 2 hashes SHA256 de documentos. |
| Comparación de secretos con salida web | 3 variables locales, 115 archivos exportados examinados; sin coincidencias literales. Es una comprobación complementaria, no un detector universal de secretos transformados. |
| OSV/Pub | 102 versiones alojadas examinadas, sin avisos devueltos. |
| npm/Go | Con hallazgos; ver anexos. |

Las 59 alertas del árbol actual no son 59 credenciales filtradas. Las de archivos JSON de revisiones reproducen checksums o código de prueba; dos pertenecen al registro de hashes de documentos. Se revisaron las clases de coincidencia sin divulgar valores.

Se verificaron 234 archivos de aplicación del conjunto exportado sin cambios durante la auditoría. Los escáneres y evidencias locales se guardan en `.local/security-review-2026-10-23/`, excluido de Git. Los valores de credenciales no se incluyen en este informe ni en el registro de revisiones.

### Observación pública, sin sesión

| Solicitud | Resultado observado |
|---|---|
| `http://diplomado.andaria.site/` | 301 a `https://diplomado.andaria.site/`. |
| `https://diplomado.andaria.site/login` | 200; nosniff, DENY, referrer y permissions policy presentes; CSP permite scripts en línea; sin HSTS observado. |
| `/api/v1/auth/options` | 200, `Cache-Control: no-store`, límite publicado 10/min. |
| `/api/v1/me` | 401, sin datos privados, límite publicado 120/min. |
| `/api/v1/admin/users?limit=1` | 401, sin datos privados. |

No se iniciaron sesiones reales ni se consultaron registros personales del VPS. Las cookies efectivas de un ingreso real, los privilegios PostgreSQL, la configuración del proxy/CDN, cifrado y permisos del host, restauración de backups y versión exacta de las imágenes desplegadas requieren validación adicional en el servidor. No se realizaron explotación, pruebas de carga, barridos de puertos ni rotación de credenciales.

## Orden recomendado de trabajo

1. **Actualizar Go y dependencias web/backend**; reconstruir, probar y volver a escanear el artefacto antes de desplegar.
2. **Separar privilegios PostgreSQL** y confirmar en el VPS la cuenta usada por la API.
3. **Corregir revocación e intentos concurrentes**, con pruebas de fallos y concurrencia.
4. **Unificar la configuración de seguridad desplegada**: CSP con nonces/hashes, HSTS y cadena de IP confiable.
5. **Cerrar operación y prevención de abuso**: cifrado/restauración, conservación de datos, verificación de correo y controles de CI.

La autorización y los filtros actuales deben conservarse al hacer estos cambios. No hay motivo para duplicar el sistema de usuarios ni exponer PostgreSQL a los clientes para cumplir la lista.

## Anexos de dependencias

Las tablas siguientes corresponden al estado local examinado y a las bases consultadas en esta revisión. “Alcanzable” significa un resultado del grafo estático, no una explotación confirmada ni que se cumplan todas las condiciones del aviso.

### A. Paquetes npm alertados

Los paquetes marcados como declarados en el manifiesto de runtime se comprobaron en el artefacto local generado; un “no” no demuestra ausencia de código empaquetado dentro de otros bundles.

| Paquete | Versión en lockfile | Gravedad npm | Declarado en manifiesto de runtime |
|---|---|---|---|
| `@simple-git/argv-parser` | 1.1.1 | critical | No |
| `shell-quote` | 1.10.0 | critical | No |
| `simple-git` | 3.36.0 | critical | No |
| `@nuxt/cli` | 3.37.0 | high | No |
| `@nuxt/devtools` | 3.4.1 | high | No |
| `@nuxt/nitro-server` | 4.5.2 | high | No |
| `@vue/server-renderer` | 3.5.41 | high | Sí |
| `brace-expansion` | 2.1.4, 5.0.9 | high | No |
| `braces` | 3.0.3 | high | No |
| `fast-glob` | 3.3.3 | high | No |
| `globby` | 16.2.3 | high | No |
| `launch-editor` | 2.14.1 | high | No |
| `listhen` | 1.10.1 | high | No |
| `micromatch` | 4.0.8 | high | No |
| `nitropack` | 2.13.4 | high | No |
| `node-forge` | 1.4.0 | high | No |
| `nuxt` | 4.5.2 | high | No |
| `seroval` | 1.6.2 | high | No |
| `source-map-js` | 1.2.1 | high | Sí |
| `undici` | 8.10.0 | high | No |
| `vite` | 8.2.1 | high | No |
| `vue` | 3.5.41 | high | Sí |
| `@nuxt/vite-builder` | 4.5.2 | moderate | No |
| `postcss-selector-parser` | 7.1.5 | moderate | No |

Los resultados completos y los enlaces de cada aviso están en el artefacto local `npm-audit.json`. No se ejecutó `npm audit fix` ni se cambió el lockfile.

### B. Avisos Go para la API Linux con Go 1.25.12

La columna de corrección procede del escáner para la rama aplicable; cuando la rama antigua ya no recibe un arreglo, aparece una versión de una rama posterior. La actualización debe comprobar la compatibilidad del conjunto.

| Aviso | Componente | Versión | Corrección indicada | Símbolo alcanzable |
|---|---|---|---|---|
| [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) | `golang.org/x/crypto` | v0.53.0 | Sin versión correctiva indicada | No observado |
| [GO-2026-5972](https://pkg.go.dev/vuln/GO-2026-5972) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6061](https://pkg.go.dev/vuln/GO-2026-6061) | `google.golang.org/grpc` | v1.82.0 | v1.82.1 | Sí, condicionado a configuración/entrada |
| [GO-2026-6088](https://pkg.go.dev/vuln/GO-2026-6088) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6089](https://pkg.go.dev/vuln/GO-2026-6089) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6090](https://pkg.go.dev/vuln/GO-2026-6090) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6091](https://pkg.go.dev/vuln/GO-2026-6091) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6218](https://pkg.go.dev/vuln/GO-2026-6218) | `stdlib` | v1.25.12 | v1.25.13 | Sí, condicionado a configuración/entrada |
| [GO-2026-6303](https://pkg.go.dev/vuln/GO-2026-6303) | `golang.org/x/crypto` | v0.53.0 | v0.55.0 | No observado |
| [GO-2026-6348](https://pkg.go.dev/vuln/GO-2026-6348) | `google.golang.org/grpc` | v1.82.0 | v1.83.1 | Sí, condicionado a configuración/entrada |
| [GO-2026-6354](https://pkg.go.dev/vuln/GO-2026-6354) | `golang.org/x/crypto` | v0.53.0 | v0.56.0 | No observado |
| [GO-2026-6355](https://pkg.go.dev/vuln/GO-2026-6355) | `golang.org/x/crypto` | v0.53.0 | v0.56.0 | No observado |
| [GO-2026-6441](https://pkg.go.dev/vuln/GO-2026-6441) | `google.golang.org/grpc` | v1.82.0 | v1.83.1 | No observado |
| [GO-2026-6443](https://pkg.go.dev/vuln/GO-2026-6443) | `google.golang.org/grpc` | v1.82.0 | v1.82.2 | No observado |
| [GO-2026-6599](https://pkg.go.dev/vuln/GO-2026-6599) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6600](https://pkg.go.dev/vuln/GO-2026-6600) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6603](https://pkg.go.dev/vuln/GO-2026-6603) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6604](https://pkg.go.dev/vuln/GO-2026-6604) | `stdlib` | v1.25.12 | v1.26.9 | No observado |
| [GO-2026-6605](https://pkg.go.dev/vuln/GO-2026-6605) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6607](https://pkg.go.dev/vuln/GO-2026-6607) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6608](https://pkg.go.dev/vuln/GO-2026-6608) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6609](https://pkg.go.dev/vuln/GO-2026-6609) | `stdlib` | v1.25.12 | v1.26.9 | No observado |
| [GO-2026-6610](https://pkg.go.dev/vuln/GO-2026-6610) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6611](https://pkg.go.dev/vuln/GO-2026-6611) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6612](https://pkg.go.dev/vuln/GO-2026-6612) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6613](https://pkg.go.dev/vuln/GO-2026-6613) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617) | `stdlib` | v1.25.12 | v1.26.9 | Sí, condicionado a configuración/entrada |
| [GO-2026-6629](https://pkg.go.dev/vuln/GO-2026-6629) | `golang.org/x/text` | v0.39.0 | v0.41.0 | Sí, condicionado a configuración/entrada |
| [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809) | `stdlib` | v1.25.12 | v1.26.10 | Sí, condicionado a configuración/entrada |

### C. Alcance de los escáneres

Gitleaks 8.30.1; govulncheck 1.9.0 y base Go actualizada al 21/10/2026; npm audit contra el registro; OSV querybatch para paquetes Pub. El primer análisis Go usó el toolchain del escáner y se descartó como referencia de producción: los resultados publicados utilizan expresamente Linux/amd64, CGO=0, Go 1.25.12 y `./cmd/api`. El análisis de `./...` queda conservado solo como evidencia complementaria. Las salidas JSON se interpretaron por sus hallazgos, no únicamente por el código de terminación del proceso.
