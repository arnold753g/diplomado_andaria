# Medusa Auditor report

- Project: Andaria Diplomado
- Module: Compra de paquetes y revisión de pagos
- Target: http://localhost:3000
- Started: 2026-09-23

## Summary

| Passed | Failed | Blocked | Not run | Total |
|---:|---:|---:|---:|---:|
| 28 | 0 | 0 | 0 | 28 |

## Executed cases

| ID | Status | Severity | Title |
|---|---|---|---|
| PUR-001 | passed | high | Los paquetes diarios, únicos y por días específicos generan salidas y cupos diferentes |
| PUR-002 | passed | critical | El turista compra una salida concreta y la compra queda en pago en revisión |
| PUR-003 | passed | critical | El precio utiliza copias de la tarifa nacional y del recargo extranjero |
| PUR-004 | passed | high | El recargo extranjero se aplica también a menores que ya pagan |
| PUR-005 | passed | high | Los menores bajo la edad mínima quedan registrados sin pagar ni ocupar cupo |
| PUR-006 | passed | high | La compra exige un comprobante de imagen válido |
| PUR-007 | passed | critical | El comprobante aceptado retiene cupos dentro de la misma transacción |
| PUR-008 | passed | critical | Los cupos retenidos reducen disponibilidad pero no cuentan para el mínimo |
| PUR-009 | passed | critical | Confirmar el pago mueve cupos retenidos a confirmados |
| PUR-010 | passed | high | Las compras confirmadas se acumulan hasta alcanzar el cupo mínimo |
| PUR-011 | passed | critical | Una compra que excede la disponibilidad se rechaza sin modificar cupos |
| PUR-012 | passed | critical | Dos compras concurrentes por el último cupo no producen sobreventa |
| PUR-013 | passed | high | Solicitar corrección conserva los cupos retenidos y registra el motivo |
| PUR-014 | passed | high | El turista propietario reemplaza el comprobante y devuelve la compra a revisión |
| PUR-015 | passed | critical | Enviar una compra a reembolso libera sus cupos |
| PUR-016 | passed | critical | Cancelar una salida registra las compras afectadas como reembolso pendiente |
| PUR-017 | passed | critical | Un turista no puede consultar ni corregir la compra de otro turista |
| PUR-018 | passed | critical | Una agencia no puede consultar ni revisar compras de otra agencia |
| PUR-019 | passed | critical | Visitante, administrador, encargado de agencia y encargado de atracción no pueden comprar como turista |
| PUR-020 | passed | critical | Solo el encargado de la agencia correspondiente puede revisar un pago |
| PUR-021 | passed | critical | Solo el turista propietario puede corregir su comprobante |
| PUR-022 | passed | high | Una decisión repetida con versión antigua se rechaza |
| PUR-023 | passed | high | Los medios de pago solo se entregan a turistas autenticados para paquetes públicos |
| PUR-024 | passed | medium | La interfaz calcula participantes, importe y muestra el flujo de comprobante |
| PUR-025 | passed | medium | Turista y agencia disponen de bandejas separadas de compras |
| PUR-026 | passed | medium | El catálogo y la ficha pública muestran los paquetes de prueba sin errores de consola |
| PUR-027 | passed | high | Las tres compras de desarrollo actualizan correctamente cupos distintos |
| PUR-028 | passed | high | La batería técnica completa termina correctamente |

## Findings

No confirmed failures were recorded.
## Untested risks

- La transferencia bancaria o pago por QR ocurre fuera de Andaria; la agencia todavía compara el comprobante manualmente.
- La finalización del reembolso, su comprobante y el QR de devolución del turista quedan para la siguiente profundización del módulo.
- La revisión de confirmación, corrección y reembolso se ejecutó contra PostgreSQL aislado; no se alteró la contraseña del encargado de la agencia de desarrollo para repetirla en su sesión real.
- No se ejecutaron pruebas de carga ni inyección de fallos porque la configuración de auditoría las deshabilita.
- La futura aplicación Flutter que compartirá esta API todavía no está implementada.
- La compilación mantiene la advertencia de un chunk de MapLibre mayor a 500 kB; no bloquea el flujo, pero conviene medir su carga antes de publicar.
