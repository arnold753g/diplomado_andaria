# Medusa Auditor report

- Project: Andaria Diplomado
- Module: Auditoria integral y preparacion de produccion
- Target: http://localhost:3000
- Started: 2026-09-25

## Summary

| Passed | Failed | Blocked | Not run | Total |
|---:|---:|---:|---:|---:|
| 19 | 0 | 1 | 1 | 21 |

## Executed cases

| ID | Status | Severity | Title |
|---|---|---|---|
| REL-001 | passed | critical | Registro, login, cierre de sesion y revocacion conservan el aislamiento de cuentas y roles |
| REL-002 | passed | critical | Las operaciones administrativas concurrentes protegen al ultimo administrador activo |
| REL-003 | passed | high | El flujo Google valida estado, identidad y vinculacion mediante proveedor simulado |
| REL-004 | passed | critical | El administrador crea y asigna agencias sin poder editar sus politicas comerciales |
| REL-005 | passed | critical | Cada encargado de agencia solo administra su agencia, pagos y politicas |
| REL-006 | passed | critical | Atracciones, asignaciones, catalogo y favoritos respetan publicacion y propiedad |
| REL-007 | passed | critical | El modulo de paquetes conserva 22 casos funcionales aprobados |
| REL-008 | passed | critical | El modulo de compras conserva 28 casos funcionales aprobados |
| REL-009 | passed | critical | El modulo de cancelaciones y reembolsos conserva 22 casos funcionales aprobados |
| REL-010 | passed | high | Los paneles muestran metricas y acciones propias de cada rol |
| REL-011 | passed | high | Portada, atracciones, paquetes, login y registro cargan mediante interfaz real |
| REL-012 | passed | high | El catalogo de paquetes mantiene uso correcto en viewport movil de 390 por 844 |
| REL-013 | passed | critical | Las rutas privadas redirigen durante SSR sin desajustes de hidratacion |
| REL-014 | passed | critical | Una sesion autenticada se reconoce en SSR y conserva la ruta privada solicitada |
| REL-015 | passed | high | Las respuestas incluyen cabeceras de seguridad y la API rechaza acceso administrativo anonimo |
| REL-016 | passed | high | Las dependencias de produccion del frontend no presentan vulnerabilidades conocidas |
| REL-017 | passed | critical | La configuracion de produccion exige HTTPS y secretos de base de datos no triviales |
| REL-018 | passed | high | El proxy admite comprobantes de 5 MB y el backend conserva salida para Google OAuth |
| REL-019 | passed | critical | La suite tecnica completa, tipos y compilacion de produccion finalizan correctamente |
| REL-020 | blocked | high | La construccion real de imagenes y validacion de Nginx en contenedor requieren Docker Engine |
| REL-021 | not_run | critical | HTTPS, DNS y Google OAuth reales se verificaran en el servidor definitivo |

## Findings

No confirmed failures were recorded.
## Untested risks

- No se probo aun el dominio real, su certificado TLS ni la renovacion automatica.
- La construccion de contenedores y nginx -t quedaron bloqueados porque Docker Engine no esta disponible en este equipo.
- Google OAuth se valido con un proveedor simulado; falta la credencial y redireccion del dominio definitivo.
- No se realizaron pruebas de carga ni inyeccion de fallos para evitar afectar el entorno local compartido.
- El mapa genera un bloque JavaScript superior a 500 kB; funciona, pero conviene optimizar su carga antes de crecer el trafico.
- Los pagos y reembolsos dependen de transferencias externas y conservan una revision humana.
- La aplicacion movil Flutter que compartira el backend pertenece a una etapa posterior.
