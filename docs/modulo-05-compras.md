# Módulo 5: compra y revisión de pagos

## Flujo acordado

1. El turista elige una salida concreta, cantidades de adultos nacionales y extranjeros, y registra únicamente la edad y condición de extranjero de cada menor.
2. El sistema toma de la agencia la edad mínima de pago. Los menores por debajo de esa edad quedan registrados, pero no pagan ni descuentan cupo. Desde esa edad pagan, consumen cupo y reciben el recargo si son extranjeros.
3. El sistema calcula el 100 % del importe con las tarifas guardadas en el paquete. La compra conserva una copia de los precios y datos del turista para que un cambio posterior no altere el historial.
4. El turista paga por QR o transferencia y adjunta un comprobante PNG/JPEG de hasta 5 MB.
5. Al enviar el comprobante, la operación bloquea la salida, vuelve a comprobar la disponibilidad y descuenta los cupos como retenidos. Si otra compra tomó los últimos cupos durante el pago, se rechaza la creación sin sobreventa.
6. La compra queda en **pago en revisión**. Los cupos retenidos reducen la disponibilidad, pero no cuentan para alcanzar el mínimo de pasajeros.
7. El encargado confirma el pago, solicita corregir el comprobante o envía el caso a reembolso. La corrección conserva el cupo retenido; el reembolso lo libera.
8. Al confirmar, los cupos pasan de retenidos a confirmados. La salida cambia a confirmada cuando la suma confirmada alcanza el cupo mínimo.

No existe una reserva previa mientras el turista realiza la transferencia. El cupo se asegura únicamente cuando el servidor acepta el comprobante. Esto mantiene la compra directa solicitada y evita reservas abandonadas.

## Estados y trazabilidad

- `payment_review`: comprobante recibido y cupo retenido.
- `correction_requested`: la agencia indicó un motivo; el turista puede reemplazar el comprobante y el cupo se conserva.
- `confirmed`: pago validado y cupo confirmado.
- `refund_pending`: compra enviada a devolución y cupo liberado.
- `refunded`: estado preparado para registrar la devolución completada en una iteración posterior.

Todas las revisiones usan versión optimista para rechazar decisiones basadas en datos antiguos. Se conserva referencia única, fecha, responsable de la revisión, motivo y fotografías del comprobante sin exponerlas en listados. Solo el turista propietario y el encargado de la agencia correspondiente pueden consultar la compra y su comprobante.

Cuando una agencia cancela una salida, las compras en revisión, en corrección o confirmadas pasan a `refund_pending` con el motivo de la salida. La salida libera los cupos retenidos y confirmados dentro de la misma transacción.

## API compartida con Flutter

Turista:

- `GET /api/v1/me/packages/{id}/payment-options`
- `GET/POST /api/v1/me/purchases`
- `GET /api/v1/me/purchases/{id}`
- `GET/PATCH /api/v1/me/purchases/{id}/proof`

Agencia:

- `GET /api/v1/agency/purchases`
- `GET /api/v1/agency/purchases/{id}`
- `GET /api/v1/agency/purchases/{id}/proof`
- `PATCH /api/v1/agency/purchases/{id}/review`

La interfaz web está en `/packages/{id}/purchase`, `/app/purchases` y `/agency/purchases`.

## Verificación

- Prueba de integración PostgreSQL para precios, menores, recargo extranjero, medios de pago, aislamiento por turista y agencia, corrección, confirmación, reembolso y cancelación de salida.
- Carrera sincronizada sobre el último cupo: una solicitud crea la compra y la otra recibe conflicto, sin sobreventa.
- Protección contra decisión duplicada o desactualizada mediante `version`.
- Tres compras reales de desarrollo sobre paquetes con frecuencias y cupos distintos; el catálogo reflejó 6/8, 12/14 y 4/5 cupos disponibles.
- Revisión pública en navegador sin errores de consola, 15 pruebas frontend, typecheck y compilación de producción.
- Informe reproducible en `.medusa-auditor/results/purchases-report.md`.

La transferencia o lectura bancaria sigue siendo externa: la agencia compara manualmente el comprobante. La cancelación, los destinos QR o bancarios y la constancia de devolución se describen en [Módulo 6](modulo-06-cancelaciones-reembolsos.md).
