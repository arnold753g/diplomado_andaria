# Módulo 6: cancelaciones y reembolsos

## Cancelación solicitada por el turista

Una compra en revisión, con corrección solicitada o confirmada puede cancelarse cuando el paquete permite cancelaciones y todavía no venció la anticipación configurada. El servidor vuelve a validar ambas condiciones dentro de la misma transacción que libera los cupos.

El turista registra un motivo y elige dónde recibirá el 100 % de la devolución:

- QR PNG/JPEG de hasta 5 MB; o
- banco, titular y número de cuenta.

La cancelación no necesita aprobación de la agencia. La compra pasa directamente a `refund_pending`, se registra la fecha de solicitud y se fija un vencimiento de 72 horas. Si el plazo ya terminó, el servidor rechaza la operación y conserva la compra y sus cupos.

## Cupo mínimo de la salida

Un proceso del backend revisa cada minuto las salidas cuyo cierre de compras ya ocurrió. Cuando los cupos confirmados no alcanzan el mínimo, la salida pasa a `minimum_review`. Esta evaluación no cancela la salida ni modifica sus compras.

El encargado de la agencia debe revisar la salida y confirmar expresamente la cancelación. Solo entonces la salida queda cancelada, libera sus cupos y las compras en revisión, en corrección o confirmadas pasan a reembolso pendiente. Si antes de esa decisión se confirman pagos suficientes para alcanzar el mínimo, la salida vuelve a `confirmed`.

La operación usa bloqueo de la salida y versión optimista para impedir decisiones simultáneas o desactualizadas.

## Registro de la devolución

La bandeja de la agencia muestra el motivo, el vencimiento, los datos bancarios o el QR indicado por el turista y señala los casos vencidos. La agencia solo puede completar el reembolso cuando existe un destino válido y adjunta un comprobante PNG/JPEG de hasta 5 MB. Puede añadir una referencia bancaria.

Al completar la devolución se conserva:

- fecha y hora;
- usuario de la agencia que la registró;
- comprobante;
- referencia de la operación;
- destino utilizado;
- motivo y origen de la cancelación.

El turista puede consultar posteriormente el comprobante. Los QR y comprobantes se entregan únicamente al propietario de la compra o al encargado de la agencia correspondiente, con caché privada deshabilitada.

## API compartida con Flutter

Turista:

- `POST /api/v1/me/purchases/{id}/cancel`
- `PATCH /api/v1/me/purchases/{id}/refund-destination`
- `GET /api/v1/me/purchases/{id}/refund-qr`
- `GET /api/v1/me/purchases/{id}/refund-proof`

Agencia:

- `POST /api/v1/agency/packages/{id}/departures/{departure}/minimum-refund`
- `GET /api/v1/agency/purchases/{id}/refund-qr`
- `GET /api/v1/agency/purchases/{id}/refund-proof`
- `PATCH /api/v1/agency/purchases/{id}/refund`

## Verificación

- Cancelación válida con liberación de cupo y vencimiento exacto de 72 horas.
- Cancelación fuera de plazo rechazada sin modificar capacidad.
- Destinos QR y cuenta bancaria validados y aislados por propietario y agencia.
- Reembolso completado únicamente con destino y comprobante.
- Evaluación del cupo mínimo separada de la confirmación de la agencia.
- Cancelación por mínimo insuficiente con todas las compras afectadas en seguimiento.
- Conflictos de versión, roles y acceso cruzado cubiertos en PostgreSQL.
- 16 pruebas frontend, typecheck y compilación de producción aprobados.

El movimiento de dinero continúa ocurriendo fuera de Andaria. El sistema registra y prueba el flujo operativo y su evidencia, pero no consulta automáticamente al banco.
