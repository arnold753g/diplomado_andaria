# Medusa Auditor report

- Project: Andaria Diplomado
- Module: Cancelaciones y reembolsos
- Target: http://localhost:3000
- Started: 2026-09-23

## Summary

| Passed | Failed | Blocked | Not run | Total |
|---:|---:|---:|---:|---:|
| 22 | 0 | 0 | 0 | 22 |

## Executed cases

| ID | Status | Severity | Title |
|---|---|---|---|
| REF-001 | passed | critical | El cierre bajo el cupo mínimo solo marca la salida para revisión |
| REF-002 | passed | critical | La agencia debe confirmar expresamente la cancelación por cupo mínimo |
| REF-003 | passed | critical | La confirmación por mínimo cancela la salida e inicia los reembolsos afectados |
| REF-004 | passed | critical | La cancelación válida del turista es automática y libera sus cupos |
| REF-005 | passed | high | La cancelación fuera del plazo se rechaza sin alterar la compra ni su capacidad |
| REF-006 | passed | high | El plazo de devolución es exactamente de 72 horas |
| REF-007 | passed | high | El turista puede elegir un QR válido como destino de devolución |
| REF-008 | passed | high | El turista puede elegir una cuenta bancaria como destino de devolución |
| REF-009 | passed | high | Los datos bancarios incompletos se rechazan |
| REF-010 | passed | critical | La agencia no puede completar una devolución sin comprobante |
| REF-011 | passed | critical | La agencia no puede completar una devolución sin destino del turista |
| REF-012 | passed | critical | Completar la devolución conserva comprobante, referencia, fecha y responsable |
| REF-013 | passed | high | Una segunda finalización con versión antigua se rechaza |
| REF-014 | passed | critical | Solo el turista propietario puede cancelar o cambiar su destino |
| REF-015 | passed | critical | Solo la agencia asignada puede confirmar el mínimo o completar el reembolso |
| REF-016 | passed | critical | El QR de devolución queda aislado entre turistas y agencias |
| REF-017 | passed | critical | El comprobante de devolución queda aislado entre turistas y agencias |
| REF-018 | passed | high | La base de datos restringe estados, destino, vencimiento y tamaño de imágenes |
| REF-019 | passed | medium | La bandeja del turista presenta política, destino, plazo y comprobante final |
| REF-020 | passed | medium | La bandeja de la agencia presenta vencidos, destino y registro del comprobante |
| REF-021 | passed | medium | Las etiquetas distinguen QR, cuenta bancaria y destino pendiente |
| REF-022 | passed | high | La batería técnica completa y la recarga pública finalizan correctamente |

## Findings

No confirmed failures were recorded.
## Untested risks

- El movimiento de dinero ocurre fuera de Andaria; el sistema registra la devolución, pero no verifica la operación contra un banco.
- Todavía no se envían recordatorios por correo cuando se acerca o vence el plazo de 72 horas.
- La interfaz autenticada de cancelación y devolución quedó cubierta por tipos, compilación, lógica y API; en este ciclo la revisión visual del navegador se realizó como visitante.
- No se ejecutaron pruebas de carga ni inyección de fallos porque la configuración de auditoría las deshabilita.
- La futura aplicación Flutter que compartirá esta API todavía no está implementada.
- La compilación conserva la advertencia de un bloque de MapLibre mayor a 500 kB; no bloquea este flujo.
