# Refuerzo de seguridad de Andaria

Implementación local del 23/10/2026, revisión 015. Complementa la [auditoría inicial](revision-seguridad-20-criterios-2026-10-23.md). No equivale a una certificación ni confirma que el VPS ya ejecute estas fuentes.

## Cambios implementados

- Go 1.27.3 en toolchain, Docker y CI; actualizaciones de `x/crypto`, `x/net`, `x/text`, gRPC y dependencias relacionadas. Nuxt 4.6.2 y Vue 3.5.44. Nginx de producción pasa a 1.30.6.
- Intentos de contraseña serializados con bloqueo de fila y comprobación de cambios concurrentes de contraseña. Los errores de escritura devuelven indisponibilidad y no se ocultan. Un intento adicional no prolonga un bloqueo vigente.
- Logout informa un fallo de revocación con 503, elimina la cookie local y distingue ese resultado en la web. Una revocación confirmada impide reutilizar el token anterior.
- Registro pendiente separado de `users`: no reserva el correo ni impide al propietario entrar con Google. Token aleatorio de 256 bits, solo hash almacenado, vigencia de 30 minutos, uso único y espera de un minuto entre solicitudes. Confirmación mediante POST explícito, sin iniciar sesión automáticamente. Los datos pendientes vencidos se eliminan periódicamente.
- La web y Android entienden `verification_required` y esperan la confirmación. El enlace guarda el token en un fragmento, lo retira del historial visible y evita transmitirlo en la solicitud HTTP inicial o el referrer.
- SMTP exige TLS con certificado válido, STARTTLS o TLS implícito en 465, con límites de conexión. En producción no se permite desactivar la verificación de nuevos registros con contraseña. Sin SMTP configurado, ese registro devuelve 503; Google y los accesos existentes conservan sus flujos. **Las credenciales de correo se incorporarán al final, por decisión del propietario.**
- `email_verified_at` conserva `NULL` para cuentas antiguas sin prueba de verificación. No se declaran verificadas automáticamente ni se vinculan cuentas por coincidencia de correo. Los registros y vínculos nuevos de Google aprovechan su identidad verificada.
- La API de producción rechaza cuentas PostgreSQL con privilegios administrativos, creación de esquemas/base o pertenencia al rol propietario de tablas. Las migraciones no se ejecutan desde la API en producción.
- CSP con nonce criptográfico por respuesta SSR, sin `unsafe-inline` para scripts. Solo se autorizan fragmentos de scripts del renderer; el cuerpo de contenido no recibe privilegios. HTML `private, no-store`, fuentes y mapas permitidos explícitamente y API HTTPS configurada. Nginx deja de duplicar la CSP de Nuxt/Go y añade HSTS inicial de 300 segundos.
- Cabeceras de IP aceptadas únicamente desde peers incluidos en `TRUSTED_PROXY_CIDRS`. El limitador tiene un máximo de entradas y no descarta límites activos para admitir nuevas IP.
- CI incluye historial de secretos redactado, análisis Go, puerta npm, revisión de imágenes finales y de PostgreSQL/Nginx, inventario y escaneo Android/Dart. Dependabot propone actualizaciones semanales; no se configura fusión automática.

## Matriz de los 20 criterios

| N.º | Criterio | Implementación y condición de cierre |
|---|---|---|
| 1 | Ocultar claves | Secretos exclusivos del backend; configuración pública y Client ID OAuth siguen siendo públicos. Escaneo de fuentes y build web. |
| 2 | Secretos en Git | Gitleaks en historial y fuentes versionables, con valores redactados. Excepciones limitadas a fixtures y hashes públicos concretos. No se reescribió historial ni se rotaron credenciales sin una filtración identificada. |
| 3 | Clave pública de DB | No aplica: clientes web/móvil acceden a la API, nunca a PostgreSQL. |
| 4 | RLS | Se conserva aislamiento en la API y se limita su rol PostgreSQL. No se habilita RLS globalmente: exige identidad por transacción y políticas para catálogos públicos, turistas, agencias y administradores. Requiere una entrega específica con pruebas del pool antes de activarse. |
| 5 | Cifrado | HTTPS y almacenamiento seguro móvil conservados. Scripts de backup cifrado y ensayo de restauración preparados. Falta verificar cifrado real de discos/volúmenes, destino externo y restauración con claves del propietario; los campos actuales no ganan cifrado de aplicación en esta entrega. |
| 6 | Autenticación | Rutas privadas protegidas; catálogo público intencional. Registro pendiente sin sesión antes de verificar. |
| 7 | Acceso a registros | Filtros por propietario/rol y pruebas de aislamiento conservados. DB de ejecución separada. |
| 8 | Manipulación de campos | DTO, rechazo de campos desconocidos y cálculo comercial en servidor conservados; regresiones de permisos y compras ejecutadas. |
| 9 | Sesiones | Revocación comprobada, manejo de errores, cookies y CSRF existentes conservados; prueba de replay. |
| 10 | Contraseñas | bcrypt y política existente conservados. Tokens de verificación también almacenados como hashes. |
| 11 | Límites | Contador de fallos atómico, límite de memoria, peers explícitos y respuesta Nginx 429. Falta comprobar la cadena de IP del VPS. No ejecutar varias réplicas sin un limitador compartido. |
| 12 | Bots | Registro con verificación, espera de reenvío, límites por IP y bloqueo por cuenta. No se añadió un CAPTCHA general ni un proveedor de desafíos; escalar la protección según evidencia de abuso. |
| 13 | Consultas | Parametrización y ordenación por listas permitidas conservadas. Valores DSN escapados, sin concatenar credenciales como opciones. |
| 14 | Entradas | Validación backend existente, validación de tokens y de configuración SMTP/proxy. |
| 15 | Contenido | Interpolación de texto y recodificación de imágenes conservadas; renderizador SSR actualizado y CSP reforzada. |
| 16 | Archivos | Decodificación, formato, tamaño, dimensiones y autorización existentes conservados. |
| 17 | Datos mínimos | DTO y endpoints privados conservados. Nuevos tokens/estados internos no se incorporan al usuario público. Pendientes vencidos eliminados. |
| 18 | Cabeceras | CSP de Nuxt comprobada localmente; API conserva su política restrictiva. El proxy/CDN no debe reemplazarlas por una política permisiva ni cachear HTML. |
| 19 | HTTPS | HSTS inicial en configuración de producción, sin `includeSubDomains` ni preload. Verificar el resultado público después de desplegar. |
| 20 | Dependencias | Actualización y escaneos locales, workflows y caducidad de excepciones. Los jobs remotos y contenedores requieren ejecutarse en CI/VPS; Docker Engine no está disponible en el entorno local. |

## Resultados locales y excepciones

- Suites Go con PostgreSQL y esquemas aislados aprobadas, incluidas concurrencia, errores de DB, revocación/replay y verificación vencida/de un solo uso.
- Bootstrap PostgreSQL probado en bases desechables con datos existentes y desde una base nueva. La cuenta API permite CRUD e inserciones tras nuevas migraciones, pero rechaza DDL, TRUNCATE y escritura del ledger. Bases y roles de prueba eliminados.
- `govulncheck` recompilado con Go 1.27.3 y ejecutado para Linux/amd64, CGO deshabilitado: salida satisfactoria, sin funciones vulnerables alcanzables. Persiste un aviso de módulo, [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932), sobre OpenPGP de `x/crypto`, que no se importa en Andaria. No se suprime el análisis de funciones.
- La última consulta npm informa **0 entradas críticas y 7 altas**, propagadas desde un único aviso raíz: `node-forge` 1.4.0, ausente del artefacto SSR examinado. `tools/security/npm-exceptions.json` documenta versión exacta, responsable, razón y vencimiento **06/11/2026**. La puerta falla si aparece en runtime, cambia a una versión no revisada, vence la excepción o se añade otro aviso alto/crítico. No significa ausencia de riesgo en las herramientas de construcción. Los recuentos pueden cambiar al actualizarse la base de avisos.
- Pruebas de la puerta npm aprobadas; pruebas web, typecheck y build de producción ejecutados. La validación del formulario en el navegador funciona con CSP activa. Nonces y `no-store` comprobados en respuestas SSR.
- Gitleaks: historial de 45 commits y exportación local de fuentes versionables sin alertas pendientes tras revisar fixtures y hashes públicos. No se afirma cobertura de otros repositorios, objetos remotos o archivos ignorados.
- Comparación adicional del secreto local de longitud suficiente contra 459 archivos del build web y las entradas descomprimidas de la APK: sin coincidencias literales ni archivos de credenciales empaquetados. Una contraseña de DB de desarrollo de cuatro caracteres se excluyó porque produce coincidencias naturales en código/binarios; no permite una comprobación literal concluyente. Producción rechaza contraseñas de DB de menos de 16 caracteres y el bootstrap nuevo exige al menos 24.
- Flutter: análisis sin incidencias y 41 pruebas aprobadas, incluida espera de confirmación sin login. Se corrigió una prueba previa que tocaba un elemento antes de completar su desplazamiento en una pantalla de 320 px.
- OSV: 102 paquetes Pub alojados y 77 módulos Maven del classpath release de Android, sin avisos devueltos. No cubre por sí solo el SDK Flutter, el sistema operativo del teléfono ni fallos aún no publicados.
- APK **0.1.5+6** compilada y verificada con `apksigner`; mismo certificado de la evidencia anterior e identificador `bo.andaria.andaria_mobile`. Entrega en `output/seguridad-2026-10-23/`, con SHA256. Compilar después de pruebas sin `--no-pub`: esa opción impide regenerar el registro nativo por variante y puede dejar una referencia de `integration_test` en release. No se modifican manualmente fuentes generadas ni se incluye el plugin de pruebas en producción.
- `docker compose config --quiet` aprobado con el archivo de ejemplo; scripts de backup/restauración comprobados sintácticamente. Esto no ejecuta una copia, una restauración cifrada ni el escaneo de imágenes.

## Despliegue obligatorio: transición de cuentas de DB

**No recrear directamente la API con el archivo de entorno antiguo.** La nueva API rechazará la cuenta administrativa que antes compartía con PostgreSQL. La preparación es aditiva; no elimina datos ni cambia automáticamente la contraseña administrativa existente.

1. Comprobar un respaldo cifrado recuperable y registrar imágenes/configuración actuales en almacenamiento privado. Mantener disponibles las imágenes anteriores. Leer la configuración y privilegios reales del VPS antes de modificarlos.
2. En `.env.production`, conservar el usuario/contraseña de arranque existentes bajo `DB_ADMIN_USER` y `DB_ADMIN_PASSWORD`. Cambiar variables de la API a `DB_USER=andaria_api` y una contraseña nueva aleatoria de al menos 24 caracteres. Añadir `DB_MIGRATION_USER=andaria_migrator` y otra contraseña distinta de al menos 24 caracteres. No cambiar nombre, volumen ni proyecto Compose de la instalación existente.
3. Con PostgreSQL en marcha, ejecutar `deploy/database-roles.sql` como su administrador en la base **dedicada a Andaria**. Las variables `ANDARIA_DB_RUNTIME_PASSWORD` y `ANDARIA_DB_MIGRATION_PASSWORD` del proceso deben contener las nuevas contraseñas guardadas en el entorno privado. El script también prepara una instalación nueva; crea solo el ledger antes de las migraciones.

   Ejemplo de ejecución Linux, con las dos variables ya exportadas mediante entrada oculta/gestor de secretos:

   ```bash
   docker exec -i \
     -e ANDARIA_DB_RUNTIME_PASSWORD -e ANDARIA_DB_MIGRATION_PASSWORD \
     diplomado-andaria-postgres-1 \
     sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
     < deploy/database-roles.sql
   ```

   El script transfiere tablas/secuencias/funciones de `public` al migrador. No usarlo en una base compartida con otros proyectos. No otorga membresía del migrador a la API. Si se reutilizan roles creados manualmente, revisar sus membresías previas; el arranque de la API rechaza las que permiten actuar como propietario.

4. Construir y aplicar migraciones antes de recrear los servicios:

   ```bash
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml config --quiet
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml build backend frontend
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml run --rm --no-deps migrate
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml up -d --no-deps --no-build --force-recreate --wait --wait-timeout 120 backend frontend
   ```

5. Revisar IP real y proxy antes de servir tráfico normal. `TRUST_PROXY=false` es seguro ante suplantación, pero agrupa solicitudes bajo la IP del proxy; es un valor de preparación, no la configuración final para muchos usuarios. Identificar mediante Docker el peer inmediato que conecta Nginx con Go y configurar únicamente sus CIDR en `TRUSTED_PROXY_CIDRS`, con `TRUST_PROXY=true`. Ajustar `nginx/trusted-proxies.conf` para confiar exclusivamente en el proxy TLS real, que debe reemplazar las cabeceras entrantes de los clientes. Comprobar desde dos conexiones externas que los límites distinguen clientes y rechazan cabeceras IP falsas.
6. Revisar certificados/renovación. Cargar los dos archivos de Nginx, actualizar su imagen y validar la configuración antes de habilitarla. El nuevo include requiere recrear Nginx para incorporar el montaje; un simple reload del contenedor antiguo no añade el volumen.

   ```bash
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml pull nginx
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml run --rm --no-deps nginx nginx -t
   docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml up -d --no-deps --force-recreate nginx
   ```

7. Comprobar salud, catálogo, login web/Google/Android, permisos por rol, compras, comprobantes y cierre/replay. Verificar en la URL HTTPS pública CSP con nonce y HSTS. Si se usa `NUXT_CSP_REPORT_ONLY=true` para observar compatibilidad, la CSP todavía no bloquea scripts; regresar a `false` después de comprobarla. No cachear HTML en el CDN ni inyectar `unsafe-inline` para resolver errores. Cambiar hosts de mapas requiere revisar la política.
8. Aumentar HSTS solo después de confirmar funcionamiento y renovación; no activar subdominios/preload sin comprobarlos. HSTS puede persistir en el navegador durante su vigencia aunque se revierta el servidor.

Para revertir aplicación, conservar los roles limitados y las migraciones aditivas; utilizar las imágenes anteriores con la configuración de ejecución limitada. No devolver privilegios administrativos a la API ni borrar la base como mecanismo de reversión.

## Correo, respaldos y datos pendientes del propietario

- Completar `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD` y `SMTP_FROM` únicamente en el archivo privado del servidor. Para un correo convencional, usar las credenciales de aplicación o el mecanismo admitido por su proveedor; no desactivar TLS. Confirmar entrega y reenvío con una cuenta de prueba antes de ofrecer registro por contraseña. Este cambio no implementa recuperación de contraseña.
- Las cuentas antiguas con correo sin verificar mantienen su acceso. Hace falta definir y desplegar una campaña de verificación antes de exigirla a esas cuentas, sin asumir propiedad del correo ni vincularlas automáticamente con Google.
- `deploy/backup-encrypted.sh` transmite `pg_dump` directamente a `age`, sin crear un dump temporal en claro. Requiere `AGE_RECIPIENT` público y un `BACKUP_DIRECTORY` privado absoluto. Las identidades de descifrado deben guardarse fuera del repositorio y separadas del servidor respaldado.
- `deploy/restore-check.sh ARCHIVO.dump.age` solo restaura en una base desechable nueva de PostgreSQL en loopback y la elimina al terminar. Requiere `RESTORE_CHECK_USER`, credenciales locales privadas y `AGE_IDENTITY_FILE`. No ejecutarlo sobre una base de producción.
- Falta elegir destino externo, custodiar la clave y realizar una restauración cifrada real. Verificar además cifrado del volumen del proveedor; cifrar backups no cifra automáticamente PostgreSQL en uso. Definir plazos de conservación para documentos, cuentas y comprobantes antes de automatizar eliminación de datos comerciales.

## Decisión sobre RLS

Se mantiene la autorización explícita en Go como barrera actual, respaldada por pruebas entre turistas, agencias, encargados y administradores. PostgreSQL deja de otorgar administración al proceso API. No se agregan políticas RLS incompletas para marcar un requisito como terminado: las consultas públicas y el worker de revisión también necesitan políticas correctas. La activación futura requiere identidad transaccional (`SET LOCAL`/`set_config` local), rol sin bypass, pruebas de conexiones reutilizadas y cobertura de operaciones anidadas. Este punto continúa identificado como defensa adicional pendiente.

Fuentes de versiones y herramientas: [Go](https://go.dev/doc/devel/release), [NGINX](https://nginx.org/en/CHANGES-1.30), [Trivy](https://github.com/aquasecurity/trivy/releases/tag/v0.76.1), [Gitleaks](https://github.com/gitleaks/gitleaks/releases/tag/v8.30.1).
