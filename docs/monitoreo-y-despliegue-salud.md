# Salud de API y configuración de UptimeRobot

Actualizado en la revisión 003 del 11/10/2026. El monitor web ya está activo según captura original del propietario. El acceso SSH se utilizó sólo para inspección; no se modificaron archivos, servicios, datos ni credenciales del VPS.

## Diagnóstico confirmado por SSH

Proyecto: `/opt/apps/diplomado-andaria`, nombre Compose `diplomado-andaria`. La rama `main` conserva el commit `2f26d105e6890d67f721765ea395ee52e23c8ac6`. Los archivos `health.go` y `router.go` corresponden a la versión anterior: no hay consulta `SELECT 1` ni alias de salud de API. El contenedor backend fue creado el 10/10/2026 a las 04:51 UTC y seguía ejecutando esa versión. Desde el propio contenedor, `/health` devuelve 200 con `status=ok`; `/api/v1/health` y `/api/v1/salud` devuelven 404. La consulta pública coincide. El fallo del alias ocurre antes de pasar por Nginx.

Backend, frontend, PostgreSQL y Nginx estaban en ejecución y saludables según Docker. Migraciones estaba terminado con código 0; es un proceso que finaliza, no un servicio permanente. La configuración del VPS tiene ajustes locales en Compose y Nginx. Deben conservarse al actualizar. El chequeo de Docker anterior sólo acredita `/health`; no comprueba PostgreSQL.

Respaldos: `docs/evidencias/revision-2026-10-11/vps-lectura-revision03.json` y `salud-publica-revision03.json`. No se inspeccionaron ni publicaron variables `.env`, contraseñas o claves. El tramo completo de TLS al origen y la política de copias requieren una comprobación específica; no se deducen del estado saludable de Docker.

## Archivos preparados para la actualización

El ZIP de corrección contiene únicamente dos archivos de ejecución y esta guía. La revisión del código está probada localmente: los tres alias ejecutan `SELECT 1` con límite de dos segundos; devuelven 200 con `data.status=ok` y `data.database=ok`, o 503 `SERVICE_UNAVAILABLE` sin detalles internos. Se mantienen sesiones, rutas comerciales y esquema de datos.

Los cambios siguen siendo locales, sin commit ni push: **`git pull` todavía no los trae**. Las huellas SHA-256 de los dos archivos corregidos son:

| Archivo | SHA-256 esperado |
|---|---|
| `backend/internal/handlers/health.go` | `02d8df54b717c7c598eb8c4aa2e9865b4c734376d7854da3fb09cfb04247c8a6` |
| `backend/internal/httpapi/router.go` | `80eebd43284f1a791aa2756dc2a9d0f98947e565f15aef102161cc8f25c18f52` |

Si se transfiere en modo texto y cambian los saltos de línea, la huella cambia; el ZIP conserva los bytes exactos.

## Procedimiento preparado para que lo ejecute el propietario

Estos comandos **no se ejecutaron durante la inspección**. Primero subir el ZIP nuevo a `/tmp/Correccion-salud-API-Andaria.zip` y conservar una copia de los dos archivos anteriores y de la imagen backend. Desde el proyecto indicado:

```bash
cd /opt/apps/diplomado-andaria
unzip -o /tmp/Correccion-salud-API-Andaria.zip backend/internal/handlers/health.go backend/internal/httpapi/router.go
sha256sum backend/internal/handlers/health.go backend/internal/httpapi/router.go
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml config --quiet
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml build backend
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml up -d --no-deps --force-recreate backend
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml ps backend
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml exec -T backend wget -S -O- http://127.0.0.1:8080/api/v1/salud
curl --fail-with-body https://diplomado.andaria.site/api/v1/salud
curl --fail-with-body https://diplomado.andaria.site/health
```

Comprobar las huellas antes de construir. La actualización queda verificada al observar 200 y `database: ok` en la URL pública; registrar fecha e identificación de la imagen nueva. No se necesitan migraciones ni reconstruir web o APK. No detener PostgreSQL productivo para simular un fallo; ese caso ya pasó en aislamiento.

## UptimeRobot: siguiente paso

El monitor web 804238365 ya muestra Arriba y comprobaciones cada cinco minutos. Abrir **Edit** y comprobar que tenga seleccionado el contacto de alertas del propietario; pulsar **Test Notification** y confirmar que llega el mensaje. La captura no demuestra esa entrega. Conservar la fecha real de alta desde el panel, si está disponible; no reconstruirla a partir de la hora de una captura.

Preparar el segundo monitor con estos campos. Mantenerlo sin activar hasta corregir y verificar la API:

| Campo | Salud de API Andaria |
|---|---|
| Nombre | Andaria - API y PostgreSQL |
| Tipo | Keyword |
| URL | https://diplomado.andaria.site/api/v1/salud |
| Método | GET |
| Intervalo | 5 minutos |
| Palabra esperada | `"database":"ok"` |
| Condición | Avisar si falta la palabra; respuesta saludable HTTP 200 |
| Alertas | Caída y recuperación al contacto verificado del propietario |
| Autenticación | Pública, sin credenciales |

La documentación oficial indica monitores HTTP/Keyword cada cinco minutos en el plan gratuito: [plan gratuito](https://help.uptimerobot.com/en/articles/11604710-who-should-use-uptimerobot-s-free-plan), [creación de monitores](https://help.uptimerobot.com/en/articles/11358364-how-to-create-your-first-monitor-on-uptimerobot-quick-setup-guide).

En **Add New Monitor**, seleccionar Keyword y aviso cuando la palabra esté ausente. Activar sólo después de obtener la respuesta pública correcta. Si el perímetro devuelve 403, investigar el acceso de UptimeRobot; no convertir ese bloqueo en un resultado satisfactorio.

## Evidencia disponible y pendiente

Captura original: `uptimerobot-web-usuario-2026-10-11.png`; estado de ambos monitores: `monitores-estado-revision03.json`. El panel muestra 1 h 6 min 30 s desde el estado Arriba, sin incidentes. Sus porcentajes de 24 horas, siete o treinta días no acreditan ventanas completas. No se afirma cumplimiento del objetivo mensual del 99 %.

Pendiente: respuesta pública del chequeo PostgreSQL, alta del monitor API, fecha real de creación de ambos, prueba de entrega de alertas y captura conjunta. Mantener el monitor web funcionando y conservar historial; el objetivo mensual se evaluará con un mes completo de observaciones.
