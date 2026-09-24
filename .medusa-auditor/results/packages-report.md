# Medusa Auditor report

- Project: Andaria Diplomado
- Module: Paquetes turísticos: gestión, programación y catálogo público
- Target: http://localhost:3000
- Started: 2026-09-22

## Summary

| Passed | Failed | Blocked | Not run | Total |
|---:|---:|---:|---:|---:|
| 22 | 0 | 0 | 0 | 22 |

## Executed cases

| ID | Status | Severity | Title |
|---|---|---|---|
| PKG-001 | passed | critical | El catálogo solo publica paquetes visibles de agencias activas y publicadas |
| PKG-002 | passed | high | Los borradores y paquetes cancelados no se exponen al turista |
| PKG-003 | passed | high | El detalle y las fotografías respetan el mismo alcance público |
| PKG-004 | passed | medium | Los filtros inválidos de dificultad, precio y fecha responden 422 |
| PKG-005 | passed | critical | El cupo disponible descuenta retenciones y confirmaciones |
| PKG-006 | passed | critical | Solo las compras confirmadas cuentan para alcanzar el cupo mínimo |
| PKG-007 | passed | high | La venta exige salida futura, ventana abierta, estado permitido y cupo |
| PKG-008 | passed | medium | El itinerario público omite atracciones que dejaron de estar publicadas |
| PKG-009 | passed | critical | Una agencia no puede administrar paquetes de otra agencia |
| PKG-010 | passed | critical | Turista, administrador y encargado de atracción no acceden a la gestión de paquetes |
| PKG-011 | passed | critical | Dos actualizaciones simultáneas de la misma salida producen un único cambio |
| PKG-012 | passed | high | Una excepción manual de salida sobrevive a la regeneración de la programación |
| PKG-013 | passed | high | Las frecuencias única, diaria y por días de semana generan salidas concretas acotadas |
| PKG-014 | passed | high | La cancelación conserva motivo y trazabilidad y deja de publicar la salida |
| PKG-015 | passed | high | La publicación exige datos comerciales, fotografía e itinerario válidos |
| PKG-016 | passed | medium | El catálogo público muestra filtros, tarjetas, estado vacío y restauración |
| PKG-017 | passed | medium | La ficha pública muestra galería, itinerario, políticas y próximas salidas |
| PKG-018 | passed | medium | La navegación separa exploración pública de gestión según el rol |
| PKG-019 | passed | medium | El visitante ve ingreso y registro en vez de una acción de cierre de sesión |
| PKG-020 | passed | low | Los importes, duración y estados de salida se presentan de forma consistente |
| PKG-021 | passed | high | La API aplica cabeceras de seguridad y no presenta errores de consola en las vistas auditadas |
| PKG-022 | passed | high | Pruebas Go, go vet, tipos, pruebas frontend y compilación de producción finalizan correctamente |

## Findings

No confirmed failures were recorded.
## Untested risks

- La creación de compra, retención corta de cupos, comprobante, validación de pago y reembolso pertenecen a los módulos 5 y 6 y aún no existen.
- La aplicación Flutter que compartirá este backend todavía no está implementada.
- No se ejecutaron pruebas de carga ni inyección de fallos; el entorno auditado es de desarrollo y la configuración evita acciones disruptivas.
- La compilación advierte que el bloque de MapLibre supera 500 kB; no impide el funcionamiento, pero su tiempo de carga debe medirse antes de publicar.
- La navegación de turista autenticado se verificó con pruebas unitarias y sesiones API; la auditoría visual de este ciclo se realizó como visitante.
