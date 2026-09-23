# Preparación y despliegue de Andaria

Estado: configuración preparada y validada localmente el 25 de septiembre de 2026. La publicación real queda pendiente del dominio y del acceso al servidor.

## Arquitectura prevista

```text
Internet HTTPS
  └─ proxy TLS o balanceador
       └─ 127.0.0.1:8088 (Nginx de Andaria)
            ├─ frontend Nuxt
            └─ API Go
                 └─ PostgreSQL privado
```

PostgreSQL, frontend y API no publican puertos. Nginx escucha por defecto únicamente en `127.0.0.1:8088`, para que un proxy con certificados gestione HTTPS. El backend también usa una red de salida aislada para comunicarse con Google OAuth sin exponer un puerto adicional.

## Datos necesarios antes de publicar

- dominio definitivo y acceso a sus registros DNS;
- sistema operativo y forma de acceso al servidor;
- proxy que terminará TLS: servicio del proveedor, Caddy, Traefik u otro;
- correo de Google Cloud y credenciales OAuth, si Google estará activo en la primera publicación;
- política de copias de seguridad del proveedor o destino cifrado para `pg_dump`.

## Configuración

1. Instala Docker Engine y Docker Compose en el servidor.
2. Copia el proyecto y crea `.env.production` desde `.env.production.example`.
3. Sustituye `andaria.example`, `DB_PASSWORD` y `SESSION_SECRET`. Genera los secretos con una fuente criptográfica y no los envíes por chat ni los guardes en Git.
4. Conserva `DB_SSLMODE=disable` cuando PostgreSQL sea el contenedor privado incluido. Usa `require` o el modo exigido por el proveedor para una base externa.
5. Si habilitas Google, registra exactamente esta redirección:

   ```text
   https://DOMINIO/api/v1/auth/google/callback
   ```

6. Valida la composición antes de iniciar:

   ```powershell
   docker compose --env-file .env.production -f docker-compose.prod.yml config --quiet
   ```

## Publicación

El proxy TLS debe enviar el dominio público a `http://127.0.0.1:8088`. Después ejecuta:

```powershell
docker compose --env-file .env.production -f docker-compose.prod.yml pull
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

El servicio `migrate` aplica las migraciones antes de habilitar la API. No uses el seed de desarrollo en producción.

## Verificación posterior

Comprueba desde otra conexión de Internet:

- `https://DOMINIO/health` responde `200`;
- `/`, `/attractions` y `/packages` cargan sin contenido mixto;
- una ruta privada redirige al ingreso;
- registro e ingreso por correo funcionan;
- Google vuelve al dominio correcto, si está habilitado;
- una imagen cercana a 5 MB puede enviarse como comprobante;
- las cabeceras incluyen `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options` y `Referrer-Policy`;
- el certificado es válido y la renovación está programada.

## Copias y recuperación

Antes de cada actualización crea una copia lógica con `pg_dump` y verifica que el archivo no esté vacío. Conserva al menos una copia fuera del servidor y cifra el almacenamiento. Ensaya periódicamente la restauración en una base separada; una copia que nunca fue restaurada no constituye una recuperación comprobada.

El volumen `postgres_data` conserva datos entre recreaciones de contenedores, pero no reemplaza un backup. Fotografías, comprobantes y QR se almacenan actualmente en PostgreSQL y quedan incluidos en `pg_dump`.

## Actualización y reversión

1. Crea backup y registra la versión desplegada.
2. Construye las imágenes y deja que `migrate` termine antes del backend.
3. Ejecuta las comprobaciones posteriores.
4. Si la aplicación falla, conserva la base, vuelve al código o imagen anterior y vuelve a levantar los servicios. No reviertas migraciones manualmente sin un procedimiento específico para esa versión.

## Límites operativos actuales

- El rate limiting vive en cada proceso; se requiere Redis antes de ejecutar varias réplicas de API.
- Las imágenes viven en la base de datos; conviene migrarlas a almacenamiento de objetos cuando aumente el volumen.
- El mapa depende de OpenFreeMap y OpenStreetMap; se debe supervisar su disponibilidad y respetar sus políticas de uso.
- Los pagos y reembolsos se validan manualmente; Andaria registra evidencia, pero no confirma movimientos bancarios.
- No hay todavía correo de recuperación de contraseña ni recordatorios automáticos.
