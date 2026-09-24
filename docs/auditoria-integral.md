# Auditoría integral de Andaria

Fecha de cierre: 25 de septiembre de 2026.

Esta revisión cubre usuarios y acceso, agencias, atracciones, paquetes, compras, cancelaciones, reembolsos, paneles por rol, rutas públicas y preparación para producción. El detalle ejecutable se encuentra en `.medusa-auditor/results/integrated-release-report.md`.

## Resultado

| Estado | Cantidad |
| --- | ---: |
| Aprobado | 19 |
| Fallido | 0 |
| Bloqueado por el entorno | 1 |
| Pendiente del servidor real | 1 |

Los informes especializados mantienen además 22 casos aprobados para paquetes, 28 para compras y 22 para cancelaciones y reembolsos.

## Correcciones realizadas durante la revisión

- Las rutas privadas ahora resuelven la sesión durante SSR. Esto evita renderizar primero un panel protegido y redirigir después en el cliente, que producía desajustes de hidratación.
- La cookie de sesión usa `Path=/`, por lo que Nuxt puede recibirla y validar una solicitud directa a `/app`, `/admin` o `/agency`.
- Nginx acepta hasta 8 MB en los endpoints de comprobantes. El margen contempla la expansión de archivos de 5 MB al codificarlos para la solicitud.
- El backend conserva salida a Internet para Google OAuth mediante una red exclusiva, sin publicar su puerto.
- El punto de entrada de Nginx escucha por defecto en `127.0.0.1:8088`, listo para quedar detrás del proxy HTTPS del servidor.
- La configuración de producción rechaza orígenes HTTP, contraseñas de base de datos cortas y valores de ejemplo evidentes.
- Se actualizaron las dependencias que producían avisos de seguridad; `npm audit --omit=dev` terminó sin vulnerabilidades conocidas.
- La plantilla `.env.production.example` puede versionarse, mientras `.env.production` continúa excluido de Git.

## Verificación técnica

- `go test ./... -count=1`: aprobado con esquemas PostgreSQL aislados.
- `go vet ./...`: aprobado.
- `npm test`: 16 de 16 pruebas aprobadas.
- `npm run typecheck`: aprobado.
- `npm run build`: aprobado.
- `npm audit --omit=dev`: 0 vulnerabilidades.
- validación estática de `docker-compose.prod.yml`: aprobada.
- revisión visual de portada, atracciones, paquetes, acceso y registro: aprobada.
- revisión móvil del catálogo de paquetes en 390 × 844: aprobada.
- rutas privadas sin sesión: redirección correcta al acceso y sin errores de hidratación.
- solicitud SSR con sesión real: conserva `/app`; la sesión de prueba fue revocada al terminar.
- cabeceras CSP, `nosniff`, `DENY`, política de referencia y permisos: presentes.

## Pendientes del entorno definitivo

La construcción real de los contenedores y `nginx -t` quedaron bloqueados porque Docker Engine no está disponible en este equipo. La composición sí fue validada de forma estática.

Cuando estén disponibles el dominio y el servidor se debe comprobar DNS, certificado y renovación TLS, Google OAuth con credenciales reales, envío de un comprobante cercano a 5 MB, copias de seguridad y restauración. La secuencia está documentada en `docs/despliegue-produccion.md`.

## Riesgos conocidos

- El mapa produce un bloque JavaScript grande; funciona, pero conviene diferir y optimizar su carga cuando aumente el tráfico.
- Los pagos y reembolsos siguen dependiendo de transferencias externas y revisión humana.
- El rate limiting es local a cada proceso y requerirá un almacén compartido antes de escalar la API horizontalmente.
- La aplicación Flutter que consumirá este backend corresponde a una etapa posterior.
