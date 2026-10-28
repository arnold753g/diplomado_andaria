# PreparaciÃ³n y despliegue de Andaria

**Gestionar pagos (migracion 16):** esta version requiere aplicar tambien la autorizacion manual de salidas antes de iniciar la API. Reconstruir backend, migrador y frontend. Ver [implementacion y verificaciones](gestion-pagos.md).

**Esquema en español (migración 15):** seguir el [procedimiento coordinado](base-datos-espanol.md) al publicar esta versión. Reconstruir backend y migrador, detener las aplicaciones, migrar y después recrear los servicios. El backend anterior y el esquema nuevo no son compatibles; no es necesario borrar datos.

**ActualizaciÃ³n de seguridad 015:** antes de usar el procedimiento general de esta guÃ­a, aplicar la [transiciÃ³n de roles, CSP y proxies](seguridad-implementacion.md). La API nueva exige un usuario PostgreSQL limitado; el archivo de entorno antiguo no permite recrearla sin esa preparaciÃ³n. El registro nuevo con contraseÃ±a espera la configuraciÃ³n SMTP.

Estado actualizado en la revisiÃ³n 003 del 11/10/2026: plataforma pÃºblica HTTPS y monitor web UptimeRobot verificados. La inspecciÃ³n SSH de lectura encontrÃ³ los archivos de salud anteriores en `/opt/apps/diplomado-andaria` y 404 en `/api/v1/salud` dentro del backend. La correcciÃ³n PostgreSQL estÃ¡ probada localmente; requiere copiar los archivos correctos y reconstruir/recrear backend con el proyecto Compose `diplomado-andaria`. El monitor API y la prueba de notificaciones siguen pendientes. Consulte [salud y monitoreo](monitoreo-y-despliegue-salud.md). El procedimiento de configuraciÃ³n original se conserva abajo.

## Arquitectura prevista

```text
Internet HTTPS
  â””â”€ proxy TLS o balanceador
       â””â”€ 127.0.0.1:8088 (Nginx de Andaria)
            â”œâ”€ frontend Nuxt
            â””â”€ API Go
                 â””â”€ PostgreSQL privado
```

PostgreSQL, frontend y API no publican puertos. Nginx escucha por defecto Ãºnicamente en `127.0.0.1:8088`, para que un proxy con certificados gestione HTTPS. El backend tambiÃ©n usa una red de salida aislada para comunicarse con Google OAuth sin exponer un puerto adicional.

## Datos necesarios antes de publicar

- dominio definitivo y acceso a sus registros DNS;
- sistema operativo y forma de acceso al servidor;
- proxy que terminarÃ¡ TLS: servicio del proveedor, Caddy, Traefik u otro;
- correo de Google Cloud y credenciales OAuth, si Google estarÃ¡ activo en la primera publicaciÃ³n;
- polÃ­tica de copias de seguridad del proveedor o destino cifrado para `pg_dump`.

## ConfiguraciÃ³n

1. Instala Docker Engine y Docker Compose en el servidor.
2. Copia el proyecto y crea `.env.production` desde `.env.production.example`.
3. Sustituye `andaria.example`, `DB_PASSWORD` y `SESSION_SECRET`. Genera los secretos con una fuente criptogrÃ¡fica y no los envÃ­es por chat ni los guardes en Git.
4. Conserva `DB_SSLMODE=disable` cuando PostgreSQL sea el contenedor privado incluido. Usa `require` o el modo exigido por el proveedor para una base externa.
5. Si habilitas Google, registra exactamente esta redirecciÃ³n:

   ```text
   https://DOMINIO/api/v1/auth/google/callback
   ```

6. Valida la composiciÃ³n antes de iniciar:

   ```powershell
   docker compose --env-file .env.production -f docker-compose.prod.yml config --quiet
   ```

## PublicaciÃ³n

El proxy TLS debe enviar el dominio pÃºblico a `http://127.0.0.1:8088`. DespuÃ©s ejecuta:

```powershell
docker compose --env-file .env.production -f docker-compose.prod.yml pull
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

El servicio `migrate` aplica las migraciones antes de habilitar la API. No uses el seed de desarrollo en producciÃ³n.

## VerificaciÃ³n posterior

Comprueba desde otra conexiÃ³n de Internet:

- `https://DOMINIO/health` responde `200`;
- `/`, `/attractions` y `/packages` cargan sin contenido mixto;
- una ruta privada redirige al ingreso;
- registro e ingreso por correo funcionan;
- Google vuelve al dominio correcto, si estÃ¡ habilitado;
- una imagen cercana a 5 MB puede enviarse como comprobante;
- las cabeceras incluyen `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options` y `Referrer-Policy`;
- el certificado es vÃ¡lido y la renovaciÃ³n estÃ¡ programada.

## Copias y recuperaciÃ³n

Antes de cada actualizaciÃ³n crea una copia lÃ³gica con `pg_dump` y verifica que el archivo no estÃ© vacÃ­o. Conserva al menos una copia fuera del servidor y cifra el almacenamiento. Ensaya periÃ³dicamente la restauraciÃ³n en una base separada; una copia que nunca fue restaurada no constituye una recuperaciÃ³n comprobada.

El volumen `postgres_data` conserva datos entre recreaciones de contenedores, pero no reemplaza un backup. FotografÃ­as, comprobantes y QR se almacenan actualmente en PostgreSQL y quedan incluidos en `pg_dump`.

## ActualizaciÃ³n y reversiÃ³n

1. Crea backup y registra la versiÃ³n desplegada.
2. Construye las imÃ¡genes y deja que `migrate` termine antes del backend.
3. Ejecuta las comprobaciones posteriores.
4. Si la aplicaciÃ³n falla, conserva la base, vuelve al cÃ³digo o imagen anterior y vuelve a levantar los servicios. No reviertas migraciones manualmente sin un procedimiento especÃ­fico para esa versiÃ³n.

## LÃ­mites operativos actuales

- El rate limiting vive en cada proceso; se requiere Redis antes de ejecutar varias rÃ©plicas de API.
- Las imÃ¡genes viven en la base de datos; conviene migrarlas a almacenamiento de objetos cuando aumente el volumen.
- El mapa depende de OpenFreeMap y OpenStreetMap; se debe supervisar su disponibilidad y respetar sus polÃ­ticas de uso.
- Los pagos y reembolsos se validan manualmente; Andaria registra evidencia, pero no confirma movimientos bancarios.
- No hay todavÃ­a correo de recuperaciÃ³n de contraseÃ±a ni recordatorios automÃ¡ticos.
