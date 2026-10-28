# Base de datos en espaÃ±ol

ImplementaciÃ³n local del 26/10/2026. Esquema final: **19 tablas de aplicaciÃ³n + `migraciones_esquema`, 229 columnas y 140 restricciones**. Ver [diccionario fÃ­sico](diccionario-datos.md) y [evidencia de catÃ¡logo](evidencias/esquema-espanol-2026-10-26/esquema-fisico.json).

## Alcance y compatibilidad

Se traducen los identificadores de tablas, columnas, restricciones, Ã­ndices, secuencias, disparadores y la funciÃ³n de actualizaciÃ³n de fechas. Se usa `snake_case` ASCII, sin tildes ni Ã±. `id`, `version`, Google, OAuth, QR y hash conservan sus identificadores tÃ©cnicos cuando corresponde.

El nombre lÃ³gico de la base de datos (`DB_NAME`), los roles de conexiÃ³n (`DB_USER`, `DB_MIGRATION_USER`), las credenciales y el esquema PostgreSQL `public` no cambian. Los catÃ¡logos internos y las palabras reservadas pertenecen a PostgreSQL.

La API mantiene sus rutas, campos JSON y valores de estados (`active`, `payment_review`, `single`, etc.). Los documentos JSONB mantienen sus claves y valores. Los datos aportados por usuarios no se traducen. Los modelos GORM declaran las columnas explÃ­citamente: `gorm:"column:correo" json:"email"`. La web y la APK siguen usando el mismo contrato; este cambio de persistencia no necesita una nueva firma ni nuevas credenciales de Google.

## Migraciones y actualizaciÃ³n

Las migraciones 1â€“14 se conservan intactas. `000015_esquema_espanol.sql` renombra la estructura en una transacciÃ³n, conservando filas, claves forÃ¡neas, secuencias, comprobantes y vÃ­nculos de Google. Se reemplaza el cuerpo del disparador porque PostgreSQL no reescribe los nombres dentro de PL/pgSQL. El registro tÃ©cnico se traduce antes de consultar versiones para no repetir migraciones ya aplicadas. Se conservan sus versiones y fechas originales; si hay dos registros de migraciÃ³n, se detiene con error.

El bootstrap de permisos soporta tanto una base vacÃ­a como el registro antiguo, y puede repetirse despuÃ©s de la traducciÃ³n. La API de producciÃ³n continÃºa sin permisos DDL. La API comprueba al arrancar que existe la migraciÃ³n 15 y la estructura espaÃ±ola; si falta, pide ejecutar el migrador y no acepta trÃ¡fico con una estructura incompatible.

**No es necesario borrar datos para adoptar esta versiÃ³n.** Una base vacÃ­a aplica 1â€“15; una base anterior aplica solo las versiones pendientes. No se debe ejecutar Ãºnicamente el archivo 15 con `psql`: se utiliza el migrador para mantener correctamente el registro.

## Despliegue coordinado en el VPS

Procedimiento preparado; **no se ha ejecutado en producciÃ³n**. Se necesita copiar estas fuentes y disponer de los roles/configuraciÃ³n de [la actualizaciÃ³n de seguridad](seguridad-implementacion.md). No activar el registro con contraseÃ±a hasta configurar SMTP; esta traducciÃ³n no modifica esa decisiÃ³n.

1. Conservar `.env.production`, claves y configuraciÃ³n TLS/OAuth. Tomar una copia privada de la base antes de la actualizaciÃ³n y conservar las imÃ¡genes anteriores para recuperaciÃ³n.
2. Revisar que los roles separados ya estÃ©n provisionados mediante `deploy/database-roles.sql`, siguiendo la guÃ­a de seguridad. No ejecutar el seed de desarrollo en producciÃ³n.
3. Desde `/opt/apps/diplomado-andaria`, reconstruir **backend y migrate** y detener las aplicaciones antes de traducir tablas; el backend anterior no puede operar sobre el esquema nuevo.

```sh
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml config --quiet
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml build backend migrate
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml stop backend frontend
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml run --rm migrate
```

4. **Solo si el migrador termina con cÃ³digo 0**, levantar las aplicaciones y actualizar el destino de Nginx:

```sh
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml up -d --no-deps --wait --wait-timeout 120 backend frontend
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml exec -T nginx nginx -t
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml exec -T nginx nginx -s reload
curl --fail --silent --show-error https://diplomado.andaria.site/api/v1/salud
```

5. Comprobar el registro con una consulta de lectura (no muestra credenciales ni registros personales):

```sh
docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml exec -T postgres sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT version, nombre FROM migraciones_esquema ORDER BY version;"'
```

6. Validar inicio de sesiÃ³n normal y con Google, roles, catÃ¡logos, creaciÃ³n/ediciÃ³n de paquetes y salidas, compras y reembolsos desde web/APK. Un `migrate` terminado en `Exited (0)` es esperado: es una tarea puntual.

Si la migraciÃ³n falla, conservar el servicio detenido, revisar el error y corregir la causa. La migraciÃ³n 15 se revierte por transacciÃ³n, pero puede haber migraciones anteriores pendientes ya aplicadas. Para volver al backend anterior despuÃ©s de una traducciÃ³n completada, restaurar la copia de la base anterior junto con su imagen; cambiar Ãºnicamente la imagen no es suficiente. No existe un `down` automÃ¡tico que traduzca los nombres de vuelta.

### InstalaciÃ³n desde cero opcional

Si se decide eliminar los datos desechables del despliegue, primero identificar y comprobar el volumen conectado a `/var/lib/postgresql/data` del contenedor **postgres del proyecto `diplomado-andaria`**. Con los servicios detenidos, recrear Ãºnicamente ese volumen, provisionar roles, aplicar las migraciones y crear el administrador inicial mediante un procedimiento controlado. Se perderÃ¡n todas las cuentas, relaciones con Google, paquetes, compras y comprobantes.

No eliminar volÃºmenes globalmente ni ejecutar `docker system prune`. No borrar archivos de entorno, certificados, copias o claves de firma. El camino recomendado para esta versiÃ³n es la migraciÃ³n sin borrado, que conserva tambiÃ©n al administrador actual.

## Correspondencia de tablas

| Nombre anterior | Nombre actual |
|---|---|
| `users` | `usuarios` |
| `auth_sessions` | `sesiones_autenticacion` |
| `oauth_attempts` | `intentos_oauth` |
| `agencies` | `agencias` |
| `attractions` | `atracciones` |
| `attraction_photos` | `fotos_atracciones` |
| `attraction_favorites` | `favoritos_atracciones` |
| `attraction_categories` | `categorias_atracciones` |
| `attraction_subcategories` | `subcategorias_atracciones` |
| `attraction_classifications` | `clasificaciones_atracciones` |
| `tour_packages` | `paquetes_turisticos` |
| `tour_package_photos` | `fotos_paquete` |
| `tour_package_itinerary_days` | `dias_itinerario_paquete` |
| `tour_package_itinerary_attractions` | `atracciones_itinerario_paquete` |
| `tour_package_schedules` | `programaciones_paquete` |
| `tour_package_departures` | `salidas_paquete` |
| `tour_package_purchases` | `compras_paquete` |
| `purchase_requests` | `solicitudes_compra` |
| `pending_registrations` | `registros_pendientes` |
| `schema_migrations` | `migraciones_esquema` |


El glosario completo de columnas estÃ¡ en `backend/internal/database/nombres_esquema.json`. Las pruebas contrastan el esquema vivo con ese glosario y verifican que todos los identificadores propios cumplan el formato y el lÃ­mite de 63 bytes de PostgreSQL.

## Verificaciones

Resultados locales sobre PostgreSQL 17.5 (producciÃ³n declara PostgreSQL 16):

- InstalaciÃ³n vacÃ­a, actualizaciÃ³n desde 1â€“14 con datos ficticios y segunda ejecuciÃ³n del migrador sin cambios.
- ConservaciÃ³n de compras, vÃ­nculos de Google, sesiones, comprobantes, fechas del registro, secuencias y restricciones referenciales.
- ActualizaciÃ³n de fechas mediante el disparador, ejecutando SQL directamente.
- Permisos de migrador/API tanto en instalaciÃ³n nueva como en actualizaciÃ³n; API sin DDL, TRUNCATE ni escritura sobre el registro de migraciones.
- Suite Go completa con PostgreSQL local y `go vet`.
- Web: 25 pruebas y comprobaciÃ³n de tipos.
- Flutter: anÃ¡lisis sin incidencias y 41 pruebas.

Las verificaciones locales no sustituyen la comprobaciÃ³n del VPS despuÃ©s de desplegar.
